package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureResultSql is the SQL store for feature results (table: feature_results).
type FeatureResultSql struct {
	db *sql.DB
}

// NewFeatureResultSQL returns a new FeatureResultSql store using the given DB.
func NewFeatureResultSQL(db *sql.DB) *FeatureResultSql {
	return &FeatureResultSql{db: db}
}

func (s *FeatureResultSql) List(datasetID, annotationID string, keys []string, features []string) ([]*feature.Result, error) {
	if datasetID == "" || annotationID == "" {
		return nil, errors.New("list feature results: missing dataset_id or annotation_id")
	}

	// One query: parent rows + optional child rows.
	base := `
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
	args := []any{datasetID, annotationID}

	if len(keys) > 0 {
		in, inArgs := makeInClause(keys)
		base += " AND r.page_key IN " + in + "\n"
		args = append(args, inArgs...)
	}
	if len(features) > 0 {
		in, inArgs := makeInClause(features)
		base += " AND r.feature_id IN " + in + "\n"
		args = append(args, inArgs...)
	}

	base += "ORDER BY r.feature_id, r.page_key, v.id\n"

	rows, err := s.db.Query(base, args...)
	if err != nil {
		return nil, fmt.Errorf("list feature results: query: %w", err)
	}
	defer rows.Close()

	type parentKey struct {
		d, a, f, k string
	}

	byKey := make(map[parentKey]*feature.Result)
	order := make([]parentKey, 0)

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

		k := parentKey{d: dsID, a: annID, f: featID, k: pageKey}
		r, ok := byKey[k]
		if !ok {
			r = &feature.Result{
				DatasetID:    dsID,
				AnnotationID: annID,
				FeatureID:    featID,
				PageKey:      pageKey,
				Source: feature.ResultSource{
					Resp:     sourceResp,
					Id:       nullStr(sourceID),
					Revision: nullStr(sourceRev),
					Name:     nullStr(sourceName),
				},
				Values: nil,
			}

			// common.Meta is embedded, so set fields via promoted names (if present).
			// These assignments assume Meta contains Name, Description, CreatedAt, UpdatedAt.
			// If Meta differs, adjust these four lines.
			r.Name = name
			r.Description = desc
			r.CreatedAt = createdAt
			r.UpdatedAt = updatedAt

			byKey[k] = r
			order = append(order, k)
		}

		// Child row present?
		if surfaceNS.Valid {
			val := feature.ResultValue{
				Surface: surfaceNS.String,
			}
			r.Values = append(r.Values, val)
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

func (s *FeatureResultSql) Create(res *feature.Result) error {
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

	if err := upsertResult(tx, res); err != nil {
		return err
	}
	if err := replaceValues(tx, res); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create feature result: commit: %w", err)
	}
	return nil
}

func (s *FeatureResultSql) CreateBatch(results []*feature.Result) error {
	if len(results) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create batch feature results: begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Prepare once.
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

		createdAt, updatedAt := any(nil), any(nil)

		// If Meta timestamps exist and are non-zero, keep them. Otherwise use CURRENT_TIMESTAMP.
		// If Meta differs in your project, adjust these two blocks.
		if !res.CreatedAt.IsZero() {
			createdAt = res.CreatedAt
		}
		if !res.UpdatedAt.IsZero() {
			updatedAt = res.UpdatedAt
		}

		if _, err := upsertStmt.Exec(
			res.Name, res.Description, createdAt, updatedAt,
			res.DatasetID, res.FeatureID, res.AnnotationID, res.PageKey,
			res.Source.Resp, emptyToNull(res.Source.Id), emptyToNull(res.Source.Revision), emptyToNull(res.Source.Name),
		); err != nil {
			return fmt.Errorf("create batch feature results: upsert (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
		}

		if _, err := delValsStmt.Exec(res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey); err != nil {
			return fmt.Errorf("create batch feature results: delete values (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
		}

		for _, v := range res.Values {
			if _, err := insValStmt.Exec(
				res.DatasetID, res.FeatureID, res.AnnotationID, res.PageKey,
				v.Surface,
			); err != nil {
				return fmt.Errorf("create batch feature results: insert value (%s/%s/%s/%s): %w", res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create batch feature results: commit: %w", err)
	}
	return nil
}

func (s *FeatureResultSql) listValuesForResult(datasetID, annotationID, featureID, pageKey string) ([]feature.ResultValue, error) {
	if datasetID == "" || annotationID == "" || featureID == "" || pageKey == "" {
		return nil, errors.New("list values: missing ids")
	}

	rows, err := s.db.Query(`
SELECT surface
FROM feature_result_values
WHERE dataset_id = ? AND annotation_id = ? AND feature_id = ? AND page_key = ?
ORDER BY id
`, datasetID, annotationID, featureID, pageKey)
	if err != nil {
		return nil, fmt.Errorf("list values: query: %w", err)
	}
	defer rows.Close()

	var out []feature.ResultValue
	for rows.Next() {
		var surface string
		if err := rows.Scan(&surface); err != nil {
			return nil, fmt.Errorf("list values: scan: %w", err)
		}

		val := feature.ResultValue{Surface: surface}
		out = append(out, val)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list values: rows: %w", err)
	}
	return out, nil
}

func upsertResult(tx *sql.Tx, res *feature.Result) error {
	createdAt, updatedAt := any(nil), any(nil)

	// If Meta timestamps exist and are non-zero, keep them. Otherwise use CURRENT_TIMESTAMP.
	// If Meta differs in your project, adjust these two blocks.
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
		res.Source.Resp, emptyToNull(res.Source.Id), emptyToNull(res.Source.Revision), emptyToNull(res.Source.Name),
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

func makeInClause(vals []string) (string, []any) {
	placeholders := make([]string, 0, len(vals))
	args := make([]any, 0, len(vals))
	for _, v := range vals {
		placeholders = append(placeholders, "?")
		args = append(args, v)
	}
	return "(" + strings.Join(placeholders, ",") + ")", args
}

func emptyToNull(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}
