package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureSQL is the SQL store for features (table: features).
type FeatureSQL struct {
	db *sql.DB
}

// NewFeatureSQL returns a new FeatureSQL store using the given DB.
func NewFeatureSQL(db *sql.DB) *FeatureSQL {
	return &FeatureSQL{db: db}
}

func (s *FeatureSQL) ListFeatures(scope feature.DefScope) ([]*feature.Feature, error) {
	rows, err := s.db.Query(`
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, is_default, is_list, color, properties
FROM features
WHERE scope = ? AND dataset_id = ?
ORDER BY created_at DESC
`, scope.Type, scope.DatasetID)
	if err != nil {
		return nil, fmt.Errorf("list features: %w", err)
	}
	defer rows.Close()

	var out []*feature.Feature
	for rows.Next() {
		f, err := scanFeature(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("list features scan: %w", err)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list features rows: %w", err)
	}
	return out, nil
}

func (s *FeatureSQL) GetFeatureByID(id string) (*feature.Feature, error) {
	row := s.db.QueryRow(`
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, is_default, is_list, color, properties
FROM features
WHERE id = ?
LIMIT 1
`, id)
	f, err := scanFeature(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature not found: id=%s: %w", id, err)
		}
		return nil, fmt.Errorf("get feature: %w", err)
	}
	return f, nil
}

func (s *FeatureSQL) GetFeatureByIDInScope(scope feature.DefScope, id string) (*feature.Feature, error) {
	row := s.db.QueryRow(`
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, is_default, is_list, color, properties
FROM features
WHERE scope = ? AND dataset_id = ? AND id = ?
LIMIT 1
`, scope.Type, scope.DatasetID, id)
	f, err := scanFeature(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature not found: scope=%v id=%s: %w", scope, id, err)
		}
		return nil, fmt.Errorf("get feature: %w", err)
	}
	return f, nil
}

func (s *FeatureSQL) Create(f *feature.Feature) error {
	if f == nil {
		return errors.New("create feature: nil feature")
	}
	if f.ID == "" {
		return errors.New("create feature: missing id")
	}
	propsJSON, err := json.Marshal(f.Properties)
	if err != nil {
		return fmt.Errorf("create feature: marshal properties: %w", err)
	}

	const q = `
INSERT INTO features (
  id, name, description, dataset_id, scope,
  is_default, is_list, color, properties,
  created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?,
  ?, ?, ?, ?,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
`
	_, err = s.db.Exec(
		q,
		f.ID,
		f.Name,
		f.Description,
		f.Scope.DatasetID,
		f.Scope.Type,
		f.IsDefault,
		f.IsList,
		f.Color,
		string(propsJSON),
	)
	if err != nil {
		return fmt.Errorf("create feature: %w", err)
	}

	// Refresh timestamps from DB
	created, err := s.GetFeatureByIDInScope(f.Scope, f.ID)
	if err != nil {
		return err
	}
	f.CreatedAt = created.CreatedAt
	f.UpdatedAt = created.UpdatedAt
	return nil
}

func (s *FeatureSQL) Update(f *feature.Feature) error {
	if f == nil {
		return errors.New("update feature: nil feature")
	}

	propsJSON, err := json.Marshal(f.Properties)
	if err != nil {
		return fmt.Errorf("update feature: marshal properties: %w", err)
	}

	q := `
UPDATE features
SET
  name = ?,
  description = ?,
  is_default = ?,
  is_list = ?,
  color = ?,
  properties = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE scope = ? AND id = ? AND dataset_id = ?`
	res, err := s.db.Exec(
		q,
		f.Name,
		f.Description,
		f.IsDefault,
		f.IsList,
		f.Color,
		string(propsJSON),
		f.Scope.Type,
		f.ID,
		f.Scope.DatasetID,
	)
	if err != nil {
		return fmt.Errorf("update feature: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update feature: rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("feature not found: dataset_id=%s id=%s", f.Scope.DatasetID, f.ID)
	}

	updated, err := s.GetFeatureByIDInScope(f.Scope, f.ID)
	if err != nil {
		return err
	}

	// Keep caller’s pointer updated with server-truth metadata.
	f.ID = updated.ID
	f.CreatedAt = updated.CreatedAt
	f.UpdatedAt = updated.UpdatedAt
	return nil
}

func (s *FeatureSQL) DeleteFeature(scope feature.DefScope, id string) error {
	res, err := s.db.Exec(`DELETE FROM features WHERE scope = ? AND dataset_id = ? AND id = ?`, scope.Type, scope.DatasetID, id)
	if err != nil {
		return fmt.Errorf("delete feature: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete feature: rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("feature not found: scope=%v id=%v", scope, id)
	}
	return nil
}

func scanFeature(scanner func(...any) error) (*feature.Feature, error) {
	var (
		id          string
		name        string
		desc        string
		createdAt   time.Time
		updatedAt   time.Time
		datasetID   sql.NullString
		scope       feature.ScopeType
		isDefault   bool
		isList      bool
		color       string
		properties  string
		propertiesV []string
	)

	if err := scanner(
		&id,
		&name,
		&desc,
		&createdAt,
		&updatedAt,
		&datasetID,
		&scope,
		&isDefault,
		&isList,
		&color,
		&properties,
	); err != nil {
		return nil, err
	}

	if properties == "" {
		properties = "[]"
	}
	if err := json.Unmarshal([]byte(properties), &propertiesV); err != nil {
		return nil, fmt.Errorf("scan feature: unmarshal properties: %w", err)
	}

	return &feature.Feature{
		Meta: common.Meta{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		Scope: feature.DefScope{
			Type:      scope,
			DatasetID: datasetID.String,
		},
		IsDefault:      isDefault,
		IsList:         isList,
		Color:          color,
		Properties:     propertiesV,
		Revisions:      nil,
		LatestRevision: nil,
	}, nil
}
