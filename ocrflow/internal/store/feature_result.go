package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
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

	// Base query
	q := `
SELECT
  r.id, r.name, r.description, r.created_at, r.updated_at,
  r.dataset_id, r.annotation_id, r.feature_id, r.page_key,
  r.source_resp, r.source_id, r.source_revision, r.source_name
FROM feature_results r
WHERE r.dataset_id = ? AND r.annotation_id = ?
`
	args := []any{datasetID, annotationID}

	// Optional filters
	if len(keys) > 0 {
		q += " AND r.page_key IN (" + placeholders(len(keys)) + ")"
		for _, k := range keys {
			args = append(args, k)
		}
	}
	if len(features) > 0 {
		q += " AND r.feature_id IN (" + placeholders(len(features)) + ")"
		for _, f := range features {
			args = append(args, f)
		}
	}

	q += " ORDER BY r.created_at DESC"

	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("list feature results: %w", err)
	}
	defer rows.Close()

	var out []*feature.Result
	for rows.Next() {
		res, err := scanFeatureResult(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("list feature results scan: %w", err)
		}

		vals, err := s.listValuesForResult(res.ID)
		if err != nil {
			return nil, fmt.Errorf("list feature results values for %s: %w", res.ID, err)
		}
		res.Values = vals

		out = append(out, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list feature results rows: %w", err)
	}

	return out, nil
}

func (s *FeatureResultSql) Create(res *feature.Result) error {
	if res == nil {
		return errors.New("create feature result: nil result")
	}
	if res.ID == "" {
		return errors.New("create feature result: missing id")
	}
	if res.DatasetID == "" || res.AnnotationID == "" || res.FeatureID == "" || res.PageKey == "" {
		return errors.New("create feature result: missing dataset_id, annotation_id, feature_id, or page_key")
	}
	if strings.TrimSpace(res.Source.Resp) == "" {
		return errors.New("create feature result: missing source.resp")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create feature result: begin tx: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const qRes = `
INSERT INTO feature_results (
  id, name, description, dataset_id, feature_id, annotation_id, page_key,
  source_resp, source_id, source_revision, source_name,
  created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
`
	_, err = tx.Exec(
		qRes,
		res.ID,
		res.Name,
		res.Description,
		res.DatasetID,
		res.FeatureID,
		res.AnnotationID,
		res.PageKey,
		res.Source.Resp,
		nullIfEmpty(res.Source.Id),
		nullIfEmpty(res.Source.Revision),
		nullIfEmpty(res.Source.Name),
	)
	if err != nil {
		return fmt.Errorf("create feature result: insert feature_results: %w", err)
	}

	const qVal = `
INSERT INTO result_values (result_id, surface, properties)
VALUES (?, ?, ?)
`
	for _, v := range res.Values {
		propsJSON, err := json.Marshal(v.Properties)
		if err != nil {
			return fmt.Errorf("create feature result: marshal value properties: %w", err)
		}
		_, err = tx.Exec(qVal, res.ID, v.Surface, string(propsJSON))
		if err != nil {
			return fmt.Errorf("create feature result: insert result_values: %w", err)
		}
	}

	// Refresh timestamps from DB (optional but handy for callers)
	created, err := s.getByIDTx(tx, res.DatasetID, res.AnnotationID, res.FeatureID, res.PageKey)
	if err == nil && created != nil {
		res.CreatedAt = created.CreatedAt
		res.UpdatedAt = created.UpdatedAt
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

	// Validate first so the tx is not opened for a doomed batch.
	for i, res := range results {
		if res == nil {
			return fmt.Errorf("create batch feature results: nil result at index %d", i)
		}
		if res.ID == "" {
			return fmt.Errorf("create batch feature results: missing id at index %d", i)
		}
		if res.DatasetID == "" || res.AnnotationID == "" || res.FeatureID == "" || res.PageKey == "" {
			return fmt.Errorf("create batch feature results: missing dataset_id, annotation_id, feature_id, or page_key at index %d", i)
		}
		if strings.TrimSpace(res.Source.Resp) == "" {
			return fmt.Errorf("create batch feature results: missing source.resp at index %d", i)
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("create batch feature results: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const qRes = `
INSERT INTO feature_results (
  id, name, description, dataset_id, feature_id, annotation_id, page_key,
  source_resp, source_id, source_revision, source_name,
  created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
`
	const qVal = `
INSERT INTO result_values (result_id, surface, properties)
VALUES (?, ?, ?)
`

	stmtRes, err := tx.Prepare(qRes)
	if err != nil {
		return fmt.Errorf("create batch feature results: prepare feature_results: %w", err)
	}
	defer stmtRes.Close()

	stmtVal, err := tx.Prepare(qVal)
	if err != nil {
		return fmt.Errorf("create batch feature results: prepare result_values: %w", err)
	}
	defer stmtVal.Close()

	for i, res := range results {
		_, err := stmtRes.Exec(
			res.ID,
			res.Name,
			res.Description,
			res.DatasetID,
			res.FeatureID,
			res.AnnotationID,
			res.PageKey,
			res.Source.Resp,
			nullIfEmpty(res.Source.Id),
			nullIfEmpty(res.Source.Revision),
			nullIfEmpty(res.Source.Name),
		)
		if err != nil {
			return fmt.Errorf("create batch feature results: insert feature_results at index %d (id=%s): %w", i, res.ID, err)
		}

		for j, v := range res.Values {
			propsJSON, err := json.Marshal(v.Properties)
			if err != nil {
				return fmt.Errorf("create batch feature results: marshal value properties at index %d value %d (id=%s): %w", i, j, res.ID, err)
			}
			_, err = stmtVal.Exec(res.ID, v.Surface, string(propsJSON))
			if err != nil {
				return fmt.Errorf("create batch feature results: insert result_values at index %d value %d (id=%s): %w", i, j, res.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("create batch feature results: commit: %w", err)
	}
	return nil
}

func (s *FeatureResultSql) listValuesForResult(resultID string) ([]feature.ResultValue, error) {
	const q = `
SELECT surface, properties
FROM result_values
WHERE result_id = ?
ORDER BY id ASC
`
	rows, err := s.db.Query(q, resultID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []feature.ResultValue
	for rows.Next() {
		var surface string
		var props string
		if err := rows.Scan(&surface, &props); err != nil {
			return nil, err
		}
		if props == "" {
			props = "{}"
		}
		m := map[string]string{}
		if err := json.Unmarshal([]byte(props), &m); err != nil {
			return nil, fmt.Errorf("unmarshal result value properties: %w", err)
		}
		out = append(out, feature.ResultValue{
			Surface:    surface,
			Properties: m,
		})
	}
	return out, rows.Err()
}

// scanFeatureResult scans a parent feature_results row (no child values).
func scanFeatureResult(scanner func(...any) error) (*feature.Result, error) {
	var (
		id         string
		name       string
		desc       string
		createdAt  time.Time
		updatedAt  time.Time
		datasetID  string
		annotation string
		featureID  string
		pageKey    string
		sourceResp string
		sourceID   sql.NullString
		sourceRev  sql.NullString
		sourceName sql.NullString
	)

	if err := scanner(
		&id, &name, &desc, &createdAt, &updatedAt,
		&datasetID, &annotation, &featureID, &pageKey,
		&sourceResp, &sourceID, &sourceRev, &sourceName,
	); err != nil {
		return nil, err
	}

	return &feature.Result{
		Meta: common.Meta{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		DatasetID:    datasetID,
		AnnotationID: annotation,
		FeatureID:    featureID,
		PageKey:      pageKey,
		Source: feature.ResultSource{
			Resp:     sourceResp,
			Id:       sourceID.String,
			Revision: sourceRev.String,
			Name:     sourceName.String,
		},
		Values: nil, // loaded separately
	}, nil
}

// getByIDTx finds a result by its natural key within an existing tx (used for timestamp refresh).
func (s *FeatureResultSql) getByIDTx(tx *sql.Tx, datasetID, annotationID, featureID, pageKey string) (*feature.Result, error) {
	const q = `
SELECT
  r.id, r.name, r.description, r.created_at, r.updated_at,
  r.dataset_id, r.annotation_id, r.feature_id, r.page_key,
  r.source_resp, r.source_id, r.source_revision, r.source_name
FROM feature_results r
WHERE r.dataset_id = ? AND r.annotation_id = ? AND r.feature_id = ? AND r.page_key = ?
LIMIT 1
`
	row := tx.QueryRow(q, datasetID, annotationID, featureID, pageKey)
	r, err := scanFeatureResult(row.Scan)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, 0, n*2-1)
	for i := 0; i < n; i++ {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, '?')
	}
	return string(b)
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
