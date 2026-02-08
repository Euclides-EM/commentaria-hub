package featureplat

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// FeatureSQL is the SQL store for features (table: features).
type FeatureSQL struct {
	db *sql.DB
}

// NewFeatureSQL returns a new FeatureSQL store using the given DB.
func NewFeatureSQL(db *sql.DB) *FeatureSQL {
	return &FeatureSQL{db: db}
}

func (s *FeatureSQL) List(collectionID string) ([]*featureplat.Feature, error) {
	rows, err := s.db.Query(`
		SELECT collection_id, id, created_at, updated_at, name, description, is_root, is_default, color
		FROM features
		WHERE collection_id = ?
		ORDER BY updated_at DESC
	`, collectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*featureplat.Feature
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

func (s *FeatureSQL) GetByID(collectionID, id string) (*featureplat.Feature, error) {
	var isRoot, isDefault int
	var f featureplat.Feature
	err := s.db.QueryRow(`
		SELECT collection_id, id, created_at, updated_at, name, description, is_root, is_default, color
		FROM features
		WHERE collection_id = ? AND id = ?
		LIMIT 1
	`, collectionID, id).Scan(
		&f.CollectionID,
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
			return nil, fmt.Errorf("feature with id %s not found in collection %s", id, collectionID)
		}
		return nil, err
	}
	f.IsRoot = isRoot != 0
	f.IsDefault = isDefault != 0
	return &f, nil
}

func (s *FeatureSQL) Create(f *featureplat.Feature) error {
	if f == nil {
		return fmt.Errorf("feature is nil")
	}
	if f.CollectionID == "" {
		return fmt.Errorf("feature collection_id is empty")
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
		INSERT INTO features (collection_id, id, created_at, updated_at, name, description, is_root, is_default, color)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.CollectionID, f.ID, f.CreatedAt, f.UpdatedAt, f.Name, f.Description, isRoot, isDefault, f.Color)
	return err
}

func (s *FeatureSQL) Update(collectionID, id string, f *featureplat.Feature) error {
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
		WHERE collection_id = ? AND id = ?
	`, f.Name, f.Description, isDefault, f.Color, f.UpdatedAt, collectionID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature with id %s not found in collection %s", id, collectionID)
	}
	return nil
}

func (s *FeatureSQL) Delete(collectionID, id string) error {
	res, err := s.db.Exec(`DELETE FROM features WHERE collection_id = ? AND id = ?`, collectionID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature with id %s not found in collection %s", id, collectionID)
	}
	return nil
}

// scanFeature scans one row into a feature. Scanner is typically rows.Scan.
func scanFeature(scanner func(...any) error) (*featureplat.Feature, error) {
	f := &featureplat.Feature{}
	var isRoot, isDefault int
	err := scanner(
		&f.CollectionID,
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
