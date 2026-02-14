package store

import (
	"database/sql"
	"fmt"
	"time"

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

func (s *FeatureSQL) List(datasetID string) ([]*feature.Feature, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color
		FROM features
		WHERE dataset_id = ?
		ORDER BY updated_at DESC
	`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*feature.Feature
	for rows.Next() {
		f, err := scanFeature(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FeatureSQL) GetByID(datasetID, id string) (*feature.Feature, error) {
	var isRoot, isDefault int
	var f feature.Feature
	err := s.db.QueryRow(`
		SELECT dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color
		FROM features
		WHERE dataset_id = ? AND id = ?
		LIMIT 1
	`, datasetID, id).Scan(
		&f.DatasetID,
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Description,
		&isRoot,
		&isDefault,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feature with id %s not found in dataset %s", id, datasetID)
		}
		return nil, err
	}
	f.IsRoot = isRoot != 0
	f.IsDefault = isDefault != 0
	return &f, nil
}

func (s *FeatureSQL) Create(f *feature.Feature) error {
	if f == nil {
		return fmt.Errorf("feature is nil")
	}
	if f.DatasetID == "" {
		return fmt.Errorf("feature dataset_id is empty")
	}
	if f.ID == "" {
		return fmt.Errorf("feature id is empty")
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = f.CreatedAt
	}
	isRoot := 0
	if f.IsRoot {
		isRoot = 1
	}
	isDefault := 0
	if f.IsDefault {
		isDefault = 1
	}
	_, err := s.db.Exec(`
		INSERT INTO features (dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.DatasetID, f.ID, f.CreatedAt, f.UpdatedAt, f.Name, f.Description, isRoot, isDefault, f.Color)
	return err
}

func (s *FeatureSQL) Update(datasetID, id string, f *feature.Feature) error {
	if f == nil {
		return fmt.Errorf("feature is nil")
	}
	f.UpdatedAt = time.Now()
	isDefault := 0
	if f.IsDefault {
		isDefault = 1
	}
	res, err := s.db.Exec(`
		UPDATE features
		SET name = ?, description = ?, is_default = ?, color = ?, updated_at = ?
		WHERE dataset_id = ? AND id = ?
	`, f.Name, f.Description, isDefault, f.Color, f.UpdatedAt, datasetID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature with id %s not found in dataset %s", id, datasetID)
	}
	return nil
}

func (s *FeatureSQL) Delete(datasetID, id string) error {
	res, err := s.db.Exec(`DELETE FROM features WHERE dataset_id = ? AND id = ?`, datasetID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature with id %s not found in dataset %s", id, datasetID)
	}
	return nil
}

// scanFeature scans one row into a feature. Scanner is typically rows.Scan.
func scanFeature(scanner func(...any) error) (*feature.Feature, error) {
	f := &feature.Feature{}
	var isRoot, isDefault int
	err := scanner(
		&f.DatasetID,
		&f.ID,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Description,
		&isRoot,
		&isDefault,
		&f.Color,
	)
	if err != nil {
		return nil, err
	}
	f.IsRoot = isRoot != 0
	f.IsDefault = isDefault != 0
	return f, nil
}
