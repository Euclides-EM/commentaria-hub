package store

import (
	"database/sql"
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

func (s *FeatureSQL) List(datasetID string) ([]*feature.Feature, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color, type
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
		features, err := s.listFeatureFeatures(datasetID, f.ID)
		if err != nil {
			return nil, err
		}
		f.Features = features
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FeatureSQL) GetByID(datasetID, id string) (*feature.Feature, error) {
	f, err := s.getByIDRow(datasetID, id)
	if err != nil {
		return nil, err
	}
	features, err := s.listFeatureFeatures(datasetID, f.ID)
	if err != nil {
		return nil, err
	}
	f.Features = features
	return f, nil
}

func (s *FeatureSQL) getByIDRow(datasetID, id string) (*feature.Feature, error) {
	var isRoot, isDefault int
	var f feature.Feature
	err := s.db.QueryRow(`
		SELECT dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color, type
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
		&f.Color,
		&f.Type,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	_, err = tx.Exec(`
		INSERT INTO features (dataset_id, id, created_at, updated_at, name, description, is_root, is_default, color, type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.DatasetID, f.ID, f.CreatedAt, f.UpdatedAt, f.Name, f.Description, isRoot, isDefault, f.Color, f.Type)
	if err != nil {
		return err
	}
	if err := s.insertFeatureFeaturesTx(tx, f.DatasetID, f.ID, f.Features); err != nil {
		return err
	}
	return tx.Commit()
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
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	res, err := tx.Exec(`
		UPDATE features
		SET name = ?, description = ?, is_default = ?, color = ?, type = ?, updated_at = ?
		WHERE dataset_id = ? AND id = ?
	`, f.Name, f.Description, isDefault, f.Color, f.Type, f.UpdatedAt, datasetID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature with id %s not found in dataset %s", id, datasetID)
	}
	if _, err := tx.Exec(`DELETE FROM feature_features WHERE dataset_id = ? AND feature_id = ?`, datasetID, id); err != nil {
		return err
	}
	if err := s.insertFeatureFeaturesTx(tx, datasetID, id, f.Features); err != nil {
		return err
	}
	return tx.Commit()
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
		&f.Type,
	)
	if err != nil {
		return nil, err
	}
	f.IsRoot = isRoot != 0
	f.IsDefault = isDefault != 0
	return f, nil
}

func (s *FeatureSQL) listFeatureFeatures(datasetID, featureID string) ([]common.Reference, error) {
	rows, err := s.db.Query(`
		SELECT child_feature_id
		FROM feature_features
		WHERE dataset_id = ? AND feature_id = ?
		ORDER BY sort_order ASC
	`, datasetID, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []common.Reference
	for rows.Next() {
		var ref common.Reference
		if err := rows.Scan(&ref.ID); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func (s *FeatureSQL) insertFeatureFeaturesTx(tx *sql.Tx, datasetID, featureID string, features []common.Reference) error {
	for i, ref := range features {
		_, err := tx.Exec(`
			INSERT INTO feature_features (dataset_id, feature_id, child_feature_id, sort_order)
			VALUES (?, ?, ?, ?)
		`, datasetID, featureID, ref.ID, i)
		if err != nil {
			return err
		}
	}
	return nil
}
