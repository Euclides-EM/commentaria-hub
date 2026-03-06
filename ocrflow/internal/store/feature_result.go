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

func (s *FeatureResultSql) listQueryFallbackToOrigin(datasetID, annotationID string, keys []string, features []string) (query string, args []any) {
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

func (s *FeatureResultSql) listQueryNoFallback(datasetID, annotationID string, keys []string, features []string) (query string, args []any) {
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

func (s *FeatureResultSql) List(datasetID, annotationID string, keys []string, features []string, fallbackToOrigin bool) ([]*feature.Result, error) {
	if datasetID == "" || annotationID == "" {
		return nil, errors.New("list feature results: missing dataset_id or annotation_id")
	}

	query, args := s.listQueryNoFallback(datasetID, annotationID, keys, features)
	if fallbackToOrigin {
		query, args = s.listQueryFallbackToOrigin(datasetID, annotationID, keys, features)
	}

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
				DatasetID:    dsID,
				AnnotationID: annID,
				FeatureID:    featID,
				PageKey:      pageKey,
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
	if res.DatasetID == "" || res.AnnotationID == "" || res.FeatureID == "" || res.PageKey == "" {
		return errors.New("create feature result: missing dataset_id, annotation_id, feature_id, or page_key")
	}
	if res.Source.Resp == "" {
		return errors.New("create feature result: missing source.resp")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create feature result: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	targetAnnotationID, err := resolveTargetAnnotationIDTx(tx, res.DatasetID, res.AnnotationID, pushToOrigin)
	if err != nil {
		return fmt.Errorf("create feature result: resolve target annotation: %w", err)
	}

	targetRes := cloneResultWithAnnotationID(res, targetAnnotationID)

	if err := upsertResult(tx, targetRes); err != nil {
		return err
	}
	if err := replaceValues(tx, targetRes); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create feature result: commit: %w", err)
	}
	return nil
}

func (s *FeatureResultSql) CreateBatch(results []*feature.Result, pushToOrigin bool) error {
	if len(results) == 0 {
		return nil
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

	upsertStmt, err := tx.Prepare(`
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
`)
	if err != nil {
		return fmt.Errorf("create batch feature results: prepare upsert: %w", err)
	}
	defer upsertStmt.Close()

	delValsStmt, err := tx.Prepare(`
DELETE FROM feature_result_values
WHERE dataset_id = ? AND annotation_id = ? AND feature_id = ? AND page_key = ?
`)
	if err != nil {
		return fmt.Errorf("create batch feature results: prepare delete values: %w", err)
	}
	defer delValsStmt.Close()

	insValStmt, err := tx.Prepare(`
INSERT INTO feature_result_values (
  dataset_id, feature_id, annotation_id, page_key,
  surface
) VALUES (?, ?, ?, ?, ?)
`)
	if err != nil {
		return fmt.Errorf("create batch feature results: prepare insert value: %w", err)
	}
	defer insValStmt.Close()

	for i, res := range results {
		if res == nil {
			return fmt.Errorf("create batch feature results: results[%d] is nil", i)
		}
		if res.DatasetID == "" || res.AnnotationID == "" || res.FeatureID == "" || res.PageKey == "" {
			return fmt.Errorf("create batch feature results: results[%d] missing ids", i)
		}
		if res.Source.Resp == "" {
			return fmt.Errorf("create batch feature results: results[%d] missing source.resp", i)
		}

		targetAnnotationID, err := resolveTarget(res.DatasetID, res.AnnotationID)
		if err != nil {
			return fmt.Errorf("create batch feature results: resolve target annotation for results[%d]: %w", i, err)
		}

		createdAt, updatedAt := any(nil), any(nil)
		if !res.CreatedAt.IsZero() {
			createdAt = res.CreatedAt
		}
		if !res.UpdatedAt.IsZero() {
			updatedAt = res.UpdatedAt
		}

		if _, err := upsertStmt.Exec(
			res.Name, res.Description, createdAt, updatedAt,
			res.DatasetID, res.FeatureID, targetAnnotationID, res.PageKey,
			res.Source.Resp, lo.EmptyableToPtr(res.Source.Id), lo.EmptyableToPtr(res.Source.Revision), lo.EmptyableToPtr(res.Source.Name),
		); err != nil {
			return fmt.Errorf(
				"create batch feature results: upsert (%s/%s/%s/%s -> %s): %w",
				res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, targetAnnotationID, err,
			)
		}

		if _, err := delValsStmt.Exec(res.DatasetID, targetAnnotationID, res.FeatureID, res.PageKey); err != nil {
			return fmt.Errorf(
				"create batch feature results: delete values (%s/%s/%s/%s -> %s): %w",
				res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, targetAnnotationID, err,
			)
		}

		for _, v := range res.Values {
			if _, err := insValStmt.Exec(
				res.DatasetID, res.FeatureID, targetAnnotationID, res.PageKey,
				v.Surface,
			); err != nil {
				return fmt.Errorf(
					"create batch feature results: insert value (%s/%s/%s/%s -> %s): %w",
					res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, targetAnnotationID, err,
				)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create batch feature results: commit: %w", err)
	}
	return nil
}

func (s *FeatureResultSql) CopyResults(datasetID string, srcAnnID string, dstAnnID string) error {
	if strings.TrimSpace(datasetID) == "" || strings.TrimSpace(srcAnnID) == "" || strings.TrimSpace(dstAnnID) == "" {
		return errors.New("copy feature results: missing dataset_id, src annotation id, or dst annotation id")
	}
	if srcAnnID == dstAnnID {
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
	if err := ensureAnnotationExistsTx(tx, datasetID, dstAnnID); err != nil {
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
`, datasetID, dstAnnID, datasetID, srcAnnID)
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
  dataset_id, feature_id, ?, page_key,
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
`, dstAnnID, datasetID, srcAnnID)
	if err != nil {
		return fmt.Errorf("copy feature results: upsert destination results: %w", err)
	}

	// Copy child value rows from source to destination.
	_, err = tx.Exec(`
INSERT INTO feature_result_values (
  dataset_id, feature_id, annotation_id, page_key, surface
)
SELECT
  dataset_id, feature_id, ?, page_key, surface
FROM feature_result_values
WHERE dataset_id = ?
  AND annotation_id = ?
`, dstAnnID, datasetID, srcAnnID)
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

func upsertResult(tx *sql.Tx, res *feature.Result) error {
	createdAt, updatedAt := any(nil), any(nil)

	if !res.CreatedAt.IsZero() {
		createdAt = res.CreatedAt
	}
	if !res.UpdatedAt.IsZero() {
		updatedAt = res.UpdatedAt
	}

	_, err := tx.Exec(`
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
		res.Name, res.Description, createdAt, updatedAt,
		res.DatasetID, res.FeatureID, res.AnnotationID, res.PageKey,
		res.Source.Resp, lo.EmptyableToPtr(res.Source.Id), lo.EmptyableToPtr(res.Source.Revision), lo.EmptyableToPtr(res.Source.Name),
	)
	if err != nil {
		return fmt.Errorf("create feature result: upsert (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
	}
	return nil
}

func replaceValues(tx *sql.Tx, res *feature.Result) error {
	_, err := tx.Exec(`
DELETE FROM feature_result_values
WHERE dataset_id = ? AND annotation_id = ? AND feature_id = ? AND page_key = ?
`, res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey)
	if err != nil {
		return fmt.Errorf("create feature result: delete values (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
	}

	for _, v := range res.Values {
		_, err := tx.Exec(`
INSERT INTO feature_result_values (
  dataset_id, feature_id, annotation_id, page_key,
  surface
) VALUES (?, ?, ?, ?, ?)
`, res.DatasetID, res.FeatureID, res.AnnotationID, res.PageKey, v.Surface)
		if err != nil {
			return fmt.Errorf("create feature result: insert value (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
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
	clone.AnnotationID = annotationID

	if res.Values != nil {
		clone.Values = append([]feature.ResultValue(nil), res.Values...)
	}

	return &clone
}
