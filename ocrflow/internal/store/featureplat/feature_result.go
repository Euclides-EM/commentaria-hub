package featureplat

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// FeatureResultSQL is the SQL store for feature results (table: feature_results).
type FeatureResultSQL struct {
	db *sql.DB
}

// NewFeatureResultSQL returns a new FeatureResultSQL store using the given DB.
func NewFeatureResultSQL(db *sql.DB) *FeatureResultSQL {
	return &FeatureResultSQL{db: db}
}

func (s *FeatureResultSQL) List(collectionID string, keys []string, features []string) ([]*featureplat.FeatureResult, error) {
	query := `SELECT collection_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json FROM feature_results WHERE collection_id = ?`
	args := []any{collectionID}

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

	var out []*featureplat.FeatureResult
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

func (s *FeatureResultSQL) Create(res *featureplat.FeatureResult) error {
	if res == nil {
		return fmt.Errorf("feature result is nil")
	}
	if res.CollectionID == "" {
		return fmt.Errorf("feature result collection_id is empty")
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
		INSERT INTO feature_results (collection_id, feature, key, note, source_resp, source_id, source_revision, source_name, values_json)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(collection_id, feature, key) DO UPDATE SET
			note = excluded.note,
			source_resp = excluded.source_resp,
			source_id = excluded.source_id,
			source_revision = excluded.source_revision,
			source_name = excluded.source_name,
			values_json = excluded.values_json
	`, res.CollectionID, res.Feature, res.Key, res.Note, res.Source.Resp, res.Source.Id, res.Source.Revision, res.Source.Name, valuesJSON)
	return err
}

func scanFeatureResult(scanner func(...any) error) (*featureplat.FeatureResult, error) {
	res := &featureplat.FeatureResult{}
	var valuesJSON string
	err := scanner(
		&res.CollectionID,
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
		if err := json.Unmarshal([]byte(valuesJSON), &res.Values); err != nil {
			return nil, fmt.Errorf("failed to parse values JSON: %w", err)
		}
	}

	return res, nil
}
