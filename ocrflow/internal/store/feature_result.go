package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/samber/lo"
)

// FeatureResultSql is the SQL store for feature results (table: feature_results).
type FeatureResultSql struct {
	db *sql.DB
}

// NewFeatureResultSQL returns a new FeatureResultSql store using the given DB.
func NewFeatureResultSQL(db *sql.DB) *FeatureResultSql {
	return &FeatureResultSql{db: db}
}

// ListEditionFeatureValues returns the current values of an edition-scoped
// feature, grouped by edition. Edition feature results are upserted when a
// feature is rerun, so this always reflects the latest persisted result.
func (s *FeatureResultSql) ListEditionFeatureValues(featureID string, editionIDs []string) (map[string][]string, error) {
	valuesByEdition := make(map[string][]string, len(editionIDs))
	const batchSize = 500
	for start := 0; start < len(editionIDs); start += batchSize {
		end := min(start+batchSize, len(editionIDs))
		batch, err := s.listEditionFeatureValuesBatch(featureID, editionIDs[start:end])
		if err != nil {
			return nil, err
		}
		for editionID, values := range batch {
			valuesByEdition[editionID] = append(valuesByEdition[editionID], values...)
		}
	}
	return valuesByEdition, nil
}

func (s *FeatureResultSql) listEditionFeatureValuesBatch(featureID string, editionIDs []string) (map[string][]string, error) {
	valuesByEdition := make(map[string][]string, len(editionIDs))
	query := `
SELECT v.edition_id, v.surface
FROM edition_feature_result_values v
JOIN edition_feature_results r
  ON r.scope = v.scope
 AND r.edition_id = v.edition_id
 AND r.feature_id = v.feature_id
WHERE v.scope = ?
  AND v.feature_id = ?
  AND v.edition_id IN (` + strings.TrimSuffix(strings.Repeat("?, ", len(editionIDs)), ", ") + `)
ORDER BY v.edition_id, v.id
`
	args := []any{feature.ScopeTypeEditions, featureID}
	args = append(args, lo.ToAnySlice(editionIDs)...)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list edition feature values: query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var editionID, surface string
		if err := rows.Scan(&editionID, &surface); err != nil {
			return nil, fmt.Errorf("list edition feature values: scan: %w", err)
		}
		valuesByEdition[editionID] = append(valuesByEdition[editionID], surface)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list edition feature values: rows: %w", err)
	}
	return valuesByEdition, nil
}

func (s *FeatureResultSql) listDatasetsQueryFallbackToOrigin(datasetID, annotationID string, keys []string, features []string) (query string, args []any) {
	query = `
WITH requested_annotation AS (
  SELECT origin_annotation_id
  FROM annotations
  WHERE id = ? AND dataset_id = ?
),
ranked_results AS (
  SELECT
    r.name,
    r.description,
    r.created_at,
    r.updated_at,
    r.dataset_id,
    r.feature_id,
    r.annotation_id AS actual_annotation_id,
    ? AS effective_annotation_id,
    r.page_key,
    r.source_resp,
    r.source_id,
    r.source_revision,
    r.source_name,
    ROW_NUMBER() OVER (
      PARTITION BY r.feature_id, r.page_key
      ORDER BY CASE WHEN r.annotation_id = ? THEN 0 ELSE 1 END
    ) AS rn
  FROM feature_results r
  WHERE r.dataset_id = ?
    AND (
      r.annotation_id = ?
      OR (
        r.annotation_id = (
          SELECT origin_annotation_id
          FROM requested_annotation
        )
        AND (
          SELECT origin_annotation_id
          FROM requested_annotation
        ) <> ''
      )
    )
),
chosen_results AS (
  SELECT *
  FROM ranked_results
  WHERE rn = 1
)
SELECT
  c.name,
  c.description,
  c.created_at,
  c.updated_at,
  c.dataset_id,
  c.feature_id,
  c.effective_annotation_id,
  c.page_key,
  c.source_resp,
  c.source_id,
  c.source_revision,
  c.source_name,
  v.surface
FROM chosen_results c
LEFT JOIN feature_result_values v
  ON v.dataset_id = c.dataset_id
 AND v.feature_id = c.feature_id
 AND v.annotation_id = c.actual_annotation_id
 AND v.page_key = c.page_key
WHERE 1=1
`
	args = []any{
		annotationID, // requested_annotation lookup
		datasetID,
		annotationID, // effective_annotation_id
		annotationID, // prefer direct result over origin
		datasetID,
		annotationID,
	}

	if len(keys) > 0 {
		query += " AND c.page_key IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(keys)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(keys)...)
	}
	if len(features) > 0 {
		query += " AND c.feature_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(features)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(features)...)
	}

	query += "ORDER BY c.feature_id, c.page_key, v.id\n"
	return query, args
}

func (s *FeatureResultSql) listDatasetsQueryNoFallback(datasetID, annotationID string, keys []string, features []string) (query string, args []any) {
	query = `
SELECT
  r.name, r.description, r.created_at, r.updated_at,
  r.dataset_id, r.feature_id, r.annotation_id, r.page_key,
  r.source_resp, r.source_id, r.source_revision, r.source_name,
  v.surface
FROM feature_results r
LEFT JOIN feature_result_values v
  ON v.dataset_id = r.dataset_id
 AND v.feature_id = r.feature_id
 AND v.annotation_id = r.annotation_id
 AND v.page_key = r.page_key
WHERE r.dataset_id = ? AND r.annotation_id = ?
`
	args = []any{datasetID, annotationID}

	if len(keys) > 0 {
		query += " AND r.page_key IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(keys)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(keys)...)
	}
	if len(features) > 0 {
		query += " AND r.feature_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(features)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(features)...)
	}

	query += "ORDER BY r.feature_id, r.page_key, v.id\n"
	return query, args
}

func (s *FeatureResultSql) List(scope feature.ExecScope, keys []string, features []string, fallbackToOrigin bool) ([]*feature.Result, error) {
	switch scope.Type {
	case feature.ScopeTypeDataset:
		return s.listDataset(scope.DatasetID, scope.AnnotationID, keys, features, fallbackToOrigin)
	case feature.ScopeTypeEditions:
		return s.listEdition(keys, features)
	default:
		return nil, fmt.Errorf("list feature results: unsupported scope type: %s", scope.Type)
	}
}

func (s *FeatureResultSql) listDataset(datasetID, annotationID string, keys []string, features []string, fallbackToOrigin bool) ([]*feature.Result, error) {
	if datasetID == "" || annotationID == "" {
		return nil, errors.New("list feature results: missing dataset_id or annotation_id")
	}

	query, args := s.listDatasetsQueryNoFallback(datasetID, annotationID, keys, features)
	if fallbackToOrigin {
		query, args = s.listDatasetsQueryFallbackToOrigin(datasetID, annotationID, keys, features)
	}

	return s.listDatasetsByQuery(query, args)
}

func (s *FeatureResultSql) listEdition(keys []string, features []string) ([]*feature.Result, error) {
	query := `
SELECT
  r.name, r.description, r.created_at, r.updated_at,
  r.scope, r.edition_id, r.feature_id,
  r.source_resp, r.source_id, r.source_revision, r.source_name,
  v.surface
FROM edition_feature_results r
LEFT JOIN edition_feature_result_values v
  ON v.scope = r.scope
 AND v.edition_id = r.edition_id
 AND v.feature_id = r.feature_id
WHERE r.scope = ?
`
	args := []any{feature.ScopeTypeEditions}

	if len(keys) > 0 {
		query += " AND r.edition_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(keys)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(keys)...)
	}

	if len(features) > 0 {
		query += " AND r.feature_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(features)), ", ") + ")\n"
		args = append(args, lo.ToAnySlice(features)...)
	}

	query += "ORDER BY r.feature_id, v.id\n"
	return s.listEditionByQuery(query, args)
}

func (s *FeatureResultSql) ListForExecutionPolicy(scope feature.ExecScope, keys []string, features []string, pushToOrigin bool) ([]*feature.Result, error) {
	switch scope.Type {
	case feature.ScopeTypeDataset:
		return s.listDatasetForExecutionPolicy(scope.DatasetID, scope.AnnotationID, keys, features, pushToOrigin)
	case feature.ScopeTypeEditions:
		return s.listEditionsForExecutionPolicy(keys, features)
	default:
		return nil, fmt.Errorf("list feature results for execution policy: unsupported scope type: %s", scope.Type)

	}
}

func (s *FeatureResultSql) listDatasetForExecutionPolicy(datasetID, annotationID string, keys []string, features []string, pushToOrigin bool) ([]*feature.Result, error) {
	if datasetID == "" || annotationID == "" {
		return nil, errors.New("list feature results for execution policy: missing dataset_id or annotation_id")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("list feature results for execution policy: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	targetAnnotationID, err := resolveTargetAnnotationIDTx(tx, datasetID, annotationID, pushToOrigin)
	if err != nil {
		return nil, fmt.Errorf("list feature results for execution policy: resolve target annotation: %w", err)
	}

	query, args := s.listDatasetsQueryNoFallback(datasetID, targetAnnotationID, keys, features)
	return s.listDatasetsByQuery(query, args)
}

func (s *FeatureResultSql) listEditionsForExecutionPolicy(keys []string, features []string) ([]*feature.Result, error) {
	if len(keys) == 0 || len(features) == 0 {
		return nil, nil
	}

	query := `
SELECT
  r.name, r.description, r.created_at, r.updated_at,
  r.scope, r.edition_id, r.feature_id,
  r.source_resp, r.source_id, r.source_revision, r.source_name,
  v.surface
FROM edition_feature_results r
LEFT JOIN edition_feature_result_values v
  ON v.scope = r.scope
 AND v.edition_id = r.edition_id
 AND v.feature_id = r.feature_id
WHERE r.scope = ?
`
	args := []any{feature.ScopeTypeEditions}

	query += " AND r.edition_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(keys)), ", ") + ")\n"
	args = append(args, lo.ToAnySlice(keys)...)
	query += " AND r.feature_id IN (" + strings.TrimSuffix(strings.Repeat("?, ", len(features)), ", ") + ")\n"
	args = append(args, lo.ToAnySlice(features)...)

	query += "ORDER BY r.feature_id, r.edition_id, v.id\n"
	return s.listEditionByQuery(query, args)
}

func (s *FeatureResultSql) listEditionByQuery(query string, args []any) ([]*feature.Result, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list edition feature results: query: %w", err)
	}
	defer rows.Close()

	byKey := make(map[string]*feature.Result)
	order := make([]string, 0)

	for rows.Next() {
		var (
			name, desc                      string
			createdAt, updatedAt            time.Time
			scope                           feature.ScopeType
			editionID, featID               string
			sourceResp                      string
			sourceID, sourceRev, sourceName sql.NullString
			surfaceNS                       sql.NullString
		)

		if err := rows.Scan(
			&name, &desc, &createdAt, &updatedAt,
			&scope, &editionID, &featID,
			&sourceResp, &sourceID, &sourceRev, &sourceName,
			&surfaceNS,
		); err != nil {
			return nil, fmt.Errorf("list edition feature results: scan: %w", err)
		}

		k := fmt.Sprintf("scope_%s_edition_%s_feat_%s", scope, editionID, featID)
		r, ok := byKey[k]
		if !ok {
			r = &feature.Result{
				Scope:     feature.ExecScope{DefScope: feature.DefScope{Type: scope}},
				FeatureID: featID,
				Key:       editionID,
				Source: feature.ResultSource{
					Resp:     sourceResp,
					Id:       lo.Ternary(sourceID.Valid, sourceID.String, ""),
					Revision: lo.Ternary(sourceRev.Valid, sourceRev.String, ""),
					Name:     lo.Ternary(sourceName.Valid, sourceName.String, ""),
				},
				Values: nil,
			}

			r.Name = name
			r.Description = desc
			r.CreatedAt = createdAt
			r.UpdatedAt = updatedAt

			byKey[k] = r
			order = append(order, k)
		}

		if surfaceNS.Valid {
			r.Values = append(r.Values, feature.ResultValue{
				Surface: surfaceNS.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list edition feature results: rows: %w", err)
	}

	out := make([]*feature.Result, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out, nil
}

func (s *FeatureResultSql) listDatasetsByQuery(query string, args []any) ([]*feature.Result, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list feature results: query: %w", err)
	}
	defer rows.Close()

	byKey := make(map[string]*feature.Result)
	order := make([]string, 0)

	for rows.Next() {
		var (
			name, desc                      string
			createdAt, updatedAt            time.Time
			dsID, featID, annID, pageKey    string
			sourceResp                      string
			sourceID, sourceRev, sourceName sql.NullString
			surfaceNS                       sql.NullString
		)

		if err := rows.Scan(
			&name, &desc, &createdAt, &updatedAt,
			&dsID, &featID, &annID, &pageKey,
			&sourceResp, &sourceID, &sourceRev, &sourceName,
			&surfaceNS,
		); err != nil {
			return nil, fmt.Errorf("list feature results: scan: %w", err)
		}

		k := fmt.Sprintf("ds_%s_ann_%s_feat_%s_page_or_key_%s", dsID, annID, featID, pageKey)
		r, ok := byKey[k]
		if !ok {
			r = &feature.Result{
				Scope:     feature.NewDatasetExecScope(dsID, annID),
				FeatureID: featID,
				Key:       pageKey,
				Source: feature.ResultSource{
					Resp:     sourceResp,
					Id:       lo.Ternary(sourceID.Valid, sourceID.String, ""),
					Revision: lo.Ternary(sourceRev.Valid, sourceRev.String, ""),
					Name:     lo.Ternary(sourceName.Valid, sourceName.String, ""),
				},
				Values: nil,
			}

			r.Name = name
			r.Description = desc
			r.CreatedAt = createdAt
			r.UpdatedAt = updatedAt

			byKey[k] = r
			order = append(order, k)
		}

		if surfaceNS.Valid {
			r.Values = append(r.Values, feature.ResultValue{
				Surface: surfaceNS.String,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list feature results: rows: %w", err)
	}

	out := make([]*feature.Result, 0, len(order))
	for _, k := range order {
		out = append(out, byKey[k])
	}
	return out, nil
}

func (s *FeatureResultSql) Create(res *feature.Result, pushToOrigin bool) error {
	if res == nil {
		return errors.New("create feature result: nil result")
	}
	if res.Scope.Type == feature.ScopeTypeEditions {
		return s.createOne(res, editionResultWriter, func(tx *sql.Tx, res *feature.Result) (*feature.Result, error) {
			return res, nil
		})
	}
	return s.createOne(res, datasetResultWriter, func(tx *sql.Tx, res *feature.Result) (*feature.Result, error) {
		targetAnnotationID, err := resolveTargetAnnotationIDTx(tx, res.Scope.DatasetID, res.Scope.AnnotationID, pushToOrigin)
		if err != nil {
			return nil, fmt.Errorf("create feature result: resolve target annotation: %w", err)
		}
		return cloneResultWithAnnotationID(res, targetAnnotationID), nil
	})
}

func (s *FeatureResultSql) CreateBatch(results []*feature.Result, pushToOrigin bool) error {
	if len(results) == 0 {
		return nil
	}
	allEditions := true
	for _, res := range results {
		if res == nil || res.Scope.Type != feature.ScopeTypeEditions {
			allEditions = false
			break
		}
	}
	if allEditions {
		return s.createBatch(results, editionResultWriter, passthroughResult)
	}

	for _, res := range results {
		if res != nil && res.Scope.Type == feature.ScopeTypeEditions {
			return s.createMixedBatch(results, pushToOrigin)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create batch feature results: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Resolve target annotation ids once per distinct annotation.
	resolvedAnnotationIDs := make(map[string]string)

	resolveTarget := func(datasetID, annotationID string) (string, error) {
		key := datasetID + "::" + annotationID
		if resolved, ok := resolvedAnnotationIDs[key]; ok {
			return resolved, nil
		}
		resolved, err := resolveTargetAnnotationIDTx(tx, datasetID, annotationID, pushToOrigin)
		if err != nil {
			return "", err
		}
		resolvedAnnotationIDs[key] = resolved
		return resolved, nil
	}

	if err := writeFeatureResultsBatchTx(tx, results, datasetResultWriter, func(i int, res *feature.Result) (*feature.Result, error) {
		targetAnnotationID, err := resolveTarget(res.Scope.DatasetID, res.Scope.AnnotationID)
		if err != nil {
			return nil, fmt.Errorf("create batch feature results: resolve target annotation for results[%d]: %w", i, err)
		}
		return cloneResultWithAnnotationID(res, targetAnnotationID), nil
	}); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create batch feature results: commit: %w", err)
	}
	return nil
}

type featureResultWriter struct {
	label           string
	upsertSQL       string
	deleteValuesSQL string
	insertValueSQL  string
	validate        func(*feature.Result) error
	upsertArgs      func(*feature.Result) []any
	valueKeyArgs    func(*feature.Result) []any
	insertValueArgs func(*feature.Result, feature.ResultValue) []any
	identity        func(*feature.Result) string
}

var datasetResultWriter = featureResultWriter{
	label: "feature result",
	upsertSQL: `
INSERT INTO feature_results (
  name, description, created_at, updated_at,
  dataset_id, feature_id, annotation_id, page_key,
  source_resp, source_id, source_revision, source_name
) VALUES (
  ?, ?, COALESCE(?, CURRENT_TIMESTAMP), COALESCE(?, CURRENT_TIMESTAMP),
  ?, ?, ?, ?,
  ?, ?, ?, ?
)
ON CONFLICT(dataset_id, feature_id, annotation_id, page_key) DO UPDATE SET
  name = excluded.name,
  description = excluded.description,
  updated_at = CURRENT_TIMESTAMP,
  source_resp = excluded.source_resp,
  source_id = excluded.source_id,
  source_revision = excluded.source_revision,
  source_name = excluded.source_name
`,
	deleteValuesSQL: `
DELETE FROM feature_result_values
WHERE dataset_id = ? AND annotation_id = ? AND feature_id = ? AND page_key = ?
`,
	insertValueSQL: `
INSERT INTO feature_result_values (
  dataset_id, feature_id, annotation_id, page_key,
  surface
) VALUES (?, ?, ?, ?, ?)
`,
	validate: func(res *feature.Result) error {
		if res.Scope.DatasetID == "" || res.Scope.AnnotationID == "" || res.FeatureID == "" || res.Key == "" {
			return errors.New("missing dataset_id, annotation_id, feature_id, or page_key")
		}
		return validateResultSource(res)
	},
	upsertArgs: func(res *feature.Result) []any {
		createdAt, updatedAt := resultTimestamps(res)
		return []any{
			res.Name, res.Description, createdAt, updatedAt,
			res.Scope.DatasetID, res.FeatureID, res.Scope.AnnotationID, res.Key,
			res.Source.Resp, lo.EmptyableToPtr(res.Source.Id), lo.EmptyableToPtr(res.Source.Revision), lo.EmptyableToPtr(res.Source.Name),
		}
	},
	valueKeyArgs: func(res *feature.Result) []any {
		return []any{res.Scope.DatasetID, res.Scope.AnnotationID, res.FeatureID, res.Key}
	},
	insertValueArgs: func(res *feature.Result, v feature.ResultValue) []any {
		return []any{res.Scope.DatasetID, res.FeatureID, res.Scope.AnnotationID, res.Key, v.Surface}
	},
	identity: func(res *feature.Result) string {
		return fmt.Sprintf("%s/%s/%s/%s", res.Scope.DatasetID, res.Scope.AnnotationID, res.FeatureID, res.Key)
	},
}

var editionResultWriter = featureResultWriter{
	label: "edition feature result",
	upsertSQL: `
INSERT INTO edition_feature_results (
  name, description, created_at, updated_at,
  scope, edition_id, feature_id,
  source_resp, source_id, source_revision, source_name
) VALUES (
  ?, ?, COALESCE(?, CURRENT_TIMESTAMP), COALESCE(?, CURRENT_TIMESTAMP),
  ?, ?, ?,
  ?, ?, ?, ?
)
ON CONFLICT(scope, edition_id, feature_id) DO UPDATE SET
  name = excluded.name,
  description = excluded.description,
  updated_at = CURRENT_TIMESTAMP,
  source_resp = excluded.source_resp,
  source_id = excluded.source_id,
  source_revision = excluded.source_revision,
  source_name = excluded.source_name
`,
	deleteValuesSQL: `
DELETE FROM edition_feature_result_values
WHERE scope = ? AND edition_id = ? AND feature_id = ?
`,
	insertValueSQL: `
INSERT INTO edition_feature_result_values (
  scope, edition_id, feature_id, surface
) VALUES (?, ?, ?, ?)
`,
	validate: func(res *feature.Result) error {
		if res.Key == "" || res.FeatureID == "" {
			return errors.New("missing edition_id or feature_id")
		}
		return validateResultSource(res)
	},
	upsertArgs: func(res *feature.Result) []any {
		createdAt, updatedAt := resultTimestamps(res)
		return []any{
			res.Name, res.Description, createdAt, updatedAt,
			feature.ScopeTypeEditions, res.Key, res.FeatureID,
			res.Source.Resp, lo.EmptyableToPtr(res.Source.Id), lo.EmptyableToPtr(res.Source.Revision), lo.EmptyableToPtr(res.Source.Name),
		}
	},
	valueKeyArgs: func(res *feature.Result) []any {
		return []any{feature.ScopeTypeEditions, res.Key, res.FeatureID}
	},
	insertValueArgs: func(res *feature.Result, v feature.ResultValue) []any {
		return []any{feature.ScopeTypeEditions, res.Key, res.FeatureID, v.Surface}
	},
	identity: func(res *feature.Result) string {
		return fmt.Sprintf("%s/%s", res.Key, res.FeatureID)
	},
}

func resultTimestamps(res *feature.Result) (any, any) {
	createdAt, updatedAt := any(nil), any(nil)
	if !res.CreatedAt.IsZero() {
		createdAt = res.CreatedAt
	}
	if !res.UpdatedAt.IsZero() {
		updatedAt = res.UpdatedAt
	}
	return createdAt, updatedAt
}

func validateResultSource(res *feature.Result) error {
	if res.Source.Resp == "" {
		return errors.New("missing source.resp")
	}
	return nil
}

func passthroughResult(_ int, res *feature.Result) (*feature.Result, error) {
	return res, nil
}

func (s *FeatureResultSql) createOne(
	res *feature.Result,
	writer featureResultWriter,
	transform func(*sql.Tx, *feature.Result) (*feature.Result, error),
) error {
	if err := validateFeatureResult(res, writer); err != nil {
		return err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create %s: begin: %w", writer.label, err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err = transform(tx, res)
	if err != nil {
		return err
	}
	if err := writeFeatureResultTx(tx, res, writer); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create %s: commit: %w", writer.label, err)
	}
	return nil
}

func (s *FeatureResultSql) createBatch(
	results []*feature.Result,
	writer featureResultWriter,
	transform func(int, *feature.Result) (*feature.Result, error),
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create batch %ss: begin: %w", writer.label, err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := writeFeatureResultsBatchTx(tx, results, writer, transform); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create batch %ss: commit: %w", writer.label, err)
	}
	return nil
}

func (s *FeatureResultSql) createMixedBatch(results []*feature.Result, pushToOrigin bool) error {
	for _, res := range results {
		if err := s.Create(res, pushToOrigin); err != nil {
			return err
		}
	}
	return nil
}

func (s *FeatureResultSql) CopyResults(datasetID, srcAnnID, dstDatasetID, dstAnnID string) error {
	if strings.TrimSpace(datasetID) == "" || strings.TrimSpace(srcAnnID) == "" || strings.TrimSpace(dstDatasetID) == "" || strings.TrimSpace(dstAnnID) == "" {
		return errors.New("copy feature results: missing dataset_id, destination dataset_id src annotation id, or dst annotation id")
	}
	if srcAnnID == dstAnnID && datasetID == dstDatasetID {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("copy feature results: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Optional but useful: make sure both annotations exist in the dataset.
	if err := ensureAnnotationExistsTx(tx, datasetID, srcAnnID); err != nil {
		return fmt.Errorf("copy feature results: source annotation: %w", err)
	}
	if err := ensureAnnotationExistsTx(tx, dstDatasetID, dstAnnID); err != nil {
		return fmt.Errorf("copy feature results: destination annotation: %w", err)
	}

	// Remove destination value rows for keys that are about to be overwritten.
	_, err = tx.Exec(`
DELETE FROM feature_result_values
WHERE dataset_id = ?
  AND annotation_id = ?
  AND (feature_id, page_key) IN (
    SELECT feature_id, page_key
    FROM feature_results
    WHERE dataset_id = ?
      AND annotation_id = ?
  )
`, dstDatasetID, dstAnnID, datasetID, srcAnnID)
	if err != nil {
		return fmt.Errorf("copy feature results: delete destination values: %w", err)
	}

	// Upsert parent result rows from source to destination.
	_, err = tx.Exec(`
INSERT INTO feature_results (
  name, description, created_at, updated_at,
  dataset_id, feature_id, annotation_id, page_key,
  source_resp, source_id, source_revision, source_name
)
SELECT
  name, description, created_at, CURRENT_TIMESTAMP,
  ?, feature_id, ?, page_key,
  source_resp, source_id, source_revision, source_name
FROM feature_results
WHERE dataset_id = ?
  AND annotation_id = ?
ON CONFLICT(dataset_id, feature_id, annotation_id, page_key) DO UPDATE SET
  name = excluded.name,
  description = excluded.description,
  updated_at = CURRENT_TIMESTAMP,
  source_resp = excluded.source_resp,
  source_id = excluded.source_id,
  source_revision = excluded.source_revision,
  source_name = excluded.source_name
`, dstDatasetID, dstAnnID, datasetID, srcAnnID)
	if err != nil {
		return fmt.Errorf("copy feature results: upsert destination results: %w", err)
	}

	// Copy child value rows from source to destination.
	_, err = tx.Exec(`
INSERT INTO feature_result_values (
  dataset_id, feature_id, annotation_id, page_key, surface
)
SELECT
  ?, feature_id, ?, page_key, surface
FROM feature_result_values
WHERE dataset_id = ?
  AND annotation_id = ?
`, dstDatasetID, dstAnnID, datasetID, srcAnnID)
	if err != nil {
		return fmt.Errorf("copy feature results: insert destination values: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("copy feature results: commit: %w", err)
	}
	return nil
}

func ensureAnnotationExistsTx(tx *sql.Tx, datasetID, annotationID string) error {
	var exists int
	err := tx.QueryRow(`
SELECT 1
FROM annotations
WHERE id = ? AND dataset_id = ?
`, annotationID, datasetID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("annotation not found: dataset_id=%s annotation_id=%s", datasetID, annotationID)
		}
		return err
	}
	return nil
}

func validateFeatureResult(res *feature.Result, writer featureResultWriter) error {
	if res == nil {
		return fmt.Errorf("create %s: nil result", writer.label)
	}
	if err := writer.validate(res); err != nil {
		return fmt.Errorf("create %s: %w", writer.label, err)
	}
	return nil
}

func writeFeatureResultTx(tx *sql.Tx, res *feature.Result, writer featureResultWriter) error {
	if err := validateFeatureResult(res, writer); err != nil {
		return err
	}

	if _, err := tx.Exec(writer.upsertSQL, writer.upsertArgs(res)...); err != nil {
		return fmt.Errorf("create %s: upsert (%s): %w", writer.label, writer.identity(res), err)
	}
	if _, err := tx.Exec(writer.deleteValuesSQL, writer.valueKeyArgs(res)...); err != nil {
		return fmt.Errorf("create %s: delete values (%s): %w", writer.label, writer.identity(res), err)
	}
	for _, v := range res.Values {
		if _, err := tx.Exec(writer.insertValueSQL, writer.insertValueArgs(res, v)...); err != nil {
			return fmt.Errorf("create %s: insert value (%s): %w", writer.label, writer.identity(res), err)
		}
	}
	return nil
}

func writeFeatureResultsBatchTx(
	tx *sql.Tx,
	results []*feature.Result,
	writer featureResultWriter,
	transform func(int, *feature.Result) (*feature.Result, error),
) error {
	upsertStmt, err := tx.Prepare(writer.upsertSQL)
	if err != nil {
		return fmt.Errorf("create batch %ss: prepare upsert: %w", writer.label, err)
	}
	defer upsertStmt.Close()

	delValsStmt, err := tx.Prepare(writer.deleteValuesSQL)
	if err != nil {
		return fmt.Errorf("create batch %ss: prepare delete values: %w", writer.label, err)
	}
	defer delValsStmt.Close()

	insValStmt, err := tx.Prepare(writer.insertValueSQL)
	if err != nil {
		return fmt.Errorf("create batch %ss: prepare insert value: %w", writer.label, err)
	}
	defer insValStmt.Close()

	for i, res := range results {
		if err := validateFeatureResult(res, writer); err != nil {
			return fmt.Errorf("create batch %ss: results[%d]: %w", writer.label, i, err)
		}

		res, err = transform(i, res)
		if err != nil {
			return err
		}
		if err := validateFeatureResult(res, writer); err != nil {
			return fmt.Errorf("create batch %ss: results[%d]: %w", writer.label, i, err)
		}

		if _, err := upsertStmt.Exec(writer.upsertArgs(res)...); err != nil {
			return fmt.Errorf("create batch %ss: upsert (%s): %w", writer.label, writer.identity(res), err)
		}

		if _, err := delValsStmt.Exec(writer.valueKeyArgs(res)...); err != nil {
			return fmt.Errorf("create batch %ss: delete values (%s): %w", writer.label, writer.identity(res), err)
		}

		for _, v := range res.Values {
			if _, err := insValStmt.Exec(writer.insertValueArgs(res, v)...); err != nil {
				return fmt.Errorf("create batch %ss: insert value (%s): %w", writer.label, writer.identity(res), err)
			}
		}
	}
	return nil
}

func resolveTargetAnnotationIDTx(tx *sql.Tx, datasetID, annotationID string, pushToOrigin bool) (string, error) {
	if !pushToOrigin {
		return annotationID, nil
	}

	var originAnnotationID string
	err := tx.QueryRow(`
SELECT origin_annotation_id
FROM annotations
WHERE id = ? AND dataset_id = ?
`, annotationID, datasetID).Scan(&originAnnotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("annotation not found: dataset_id=%s annotation_id=%s", datasetID, annotationID)
		}
		return "", err
	}

	if strings.TrimSpace(originAnnotationID) == "" {
		return annotationID, nil
	}

	return originAnnotationID, nil
}

func cloneResultWithAnnotationID(res *feature.Result, annotationID string) *feature.Result {
	if res == nil {
		return nil
	}
	clone := *res
	clone.Scope.AnnotationID = annotationID

	if res.Values != nil {
		clone.Values = append([]feature.ResultValue(nil), res.Values...)
	}

	return &clone
}
