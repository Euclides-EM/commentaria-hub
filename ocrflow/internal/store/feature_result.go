package store

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureResultSQL is the SQL store for feature results (table: feature_results).
type FeatureResultSQL struct {
	db *sql.DB
}

// NewFeatureResultSQL returns a new FeatureResultSQL store using the given DB.
func NewFeatureResultSQL(db *sql.DB) *FeatureResultSQL {
	return &FeatureResultSQL{db: db}
}

func (s *FeatureResultSQL) List(datasetID, annotationID string, keys []string, features []string) ([]*feature.Result, error) {
	query := `SELECT dataset_id, annotation_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json FROM feature_results WHERE dataset_id = ? AND annotation_id = ?`
	args := []any{datasetID, annotationID}

	if len(keys) > 0 {
		query += ` AND key IN (`
		for i, key := range keys {
			if i > 0 {
				query += `, `
			}
			query += `?`
			args = append(args, key)
		}
		query += `)`
	}

	if len(features) > 0 {
		query += ` AND feature IN (`
		for i, feature := range features {
			if i > 0 {
				query += `, `
			}
			query += `?`
			args = append(args, feature)
		}
		query += `)`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*feature.Result
	for rows.Next() {
		res, err := scanFeatureResult(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, res)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FeatureResultSQL) Create(res *feature.Result) error {
	if res == nil {
		return fmt.Errorf("feature result is nil")
	}
	if res.DatasetID == "" {
		return fmt.Errorf("feature result dataset_id is empty")
	}
	if res.AnnotationID == "" {
		return fmt.Errorf("feature result annotation_id is empty")
	}
	if res.Feature == "" {
		return fmt.Errorf("feature result feature is empty")
	}
	if res.Key == "" {
		return fmt.Errorf("feature result key is empty")
	}

	valuesJSON := "[]"
	if len(res.Values) > 0 {
		valuesBytes, err := json.Marshal(res.Values)
		if err != nil {
			return fmt.Errorf("failed to marshal values: %w", err)
		}
		valuesJSON = string(valuesBytes)
	}

	_, err := s.db.Exec(`
		INSERT INTO feature_results (dataset_id, annotation_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(dataset_id, annotation_id, feature, key) DO UPDATE SET
			note = excluded.note,
			source_resp = excluded.source_resp,
			source_id = excluded.source_id,
			source_revision = excluded.source_revision,
			source_name = excluded.source_name,
			values_json = excluded.values_json
	`, res.DatasetID, res.AnnotationID, res.Feature, res.Key, res.Note, res.Source.Resp, res.Source.Id, res.Source.Revision, res.Source.Name, valuesJSON)
	return err
}

func (s *FeatureResultSQL) CreateBatch(results []*feature.Result) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
		INSERT INTO feature_results (dataset_id, annotation_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(dataset_id, annotation_id, feature, key) DO UPDATE SET
			note = excluded.note,
			source_resp = excluded.source_resp,
			source_id = excluded.source_id,
			source_revision = excluded.source_revision,
			source_name = excluded.source_name,
			values_json = excluded.values_json
	`)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, res := range results {
		if res == nil {
			continue
		}
		valuesJSON := "[]"
		if len(res.Values) > 0 {
			valuesBytes, err := json.Marshal(res.Values)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to marshal values: %w", err)
			}
			valuesJSON = string(valuesBytes)
		}
		if _, err := stmt.Exec(res.DatasetID, res.AnnotationID, res.Feature, res.Key, res.Note, res.Source.Resp, res.Source.Id, res.Source.Revision, res.Source.Name, valuesJSON); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func scanFeatureResult(scanner func(...any) error) (*feature.Result, error) {
	res := &feature.Result{}
	var valuesJSON string
	err := scanner(
		&res.DatasetID,
		&res.AnnotationID,
		&res.Feature,
		&res.Key,
		&res.Note,
		&res.Source.Resp,
		&res.Source.Id,
		&res.Source.Revision,
		&res.Source.Name,
		&valuesJSON,
	)
	if err != nil {
		return nil, err
	}

	if valuesJSON != "" && valuesJSON != "[]" {
		// Try to unmarshal as array of strings first (legacy format)
		var stringValues []string
		if err := json.Unmarshal([]byte(valuesJSON), &stringValues); err == nil {
			// Convert strings to FeatureResultValue objects
			res.Values = make([]feature.ResultValue, len(stringValues))
			for i, s := range stringValues {
				res.Values[i] = feature.ResultValue{
					Root: s,
				}
			}
		} else {
			// Try to unmarshal as array of FeatureResultValue objects
			if err := json.Unmarshal([]byte(valuesJSON), &res.Values); err != nil {
				return nil, fmt.Errorf("failed to parse values JSON: %w", err)
			}
		}
	}

	return res, nil
}
