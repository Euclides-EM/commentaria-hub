package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/samber/lo"
)

const ModelIDPrefix = "mod"

type ModelSQL struct {
	BaseSQL
}

func NewModelSQL(db *sql.DB) *ModelSQL {
	return &ModelSQL{
		BaseSQL{db: db},
	}
}

func (s *ModelSQL) ListModels() ([]*model.Model, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, type, location, algorithm_family, local_path, base_model_id
		FROM models
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*model.Model
	for rows.Next() {
		m := &model.Model{}
		var algoFamily string
		var baseModelID sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Description,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.Type,
			&m.Location,
			&algoFamily,
			&m.LocalPath,
			&baseModelID,
		); err != nil {
			return nil, err
		}

		// algorithm_family is stored as TEXT DEFAULT '' NOT NULL
		m.AlgorithmFamily = model.OCRModelAlgorithmFamily(algoFamily)

		cats, err := s.listModelCategories(m.ID)
		if err != nil {
			return nil, err
		}
		m.Categories = cats

		if baseModelID.Valid {
			m.BaseModelID = baseModelID.String
		}
		refs, err := s.listModelBaseAnnotations(m.ID)
		if err != nil {
			return nil, err
		}
		m.BaseAnnotations = refs

		models = append(models, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return models, nil
}

func (s *ModelSQL) GetModelByID(id string) (*model.Model, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, type, location, algorithm_family, local_path, base_model_id
		FROM models
		WHERE id = ?
		LIMIT 1
	`, id)

	m := &model.Model{}
	var algoFamily string
	var baseModelID sql.NullString
	if err := row.Scan(
		&m.ID,
		&m.Name,
		&m.Description,
		&m.CreatedAt,
		&m.UpdatedAt,
		&m.Type,
		&m.Location,
		&algoFamily,
		&m.LocalPath,
		&baseModelID,
	); err != nil {
		return nil, err
	}
	m.AlgorithmFamily = model.OCRModelAlgorithmFamily(algoFamily)

	cats, err := s.listModelCategories(m.ID)
	if err != nil {
		return nil, err
	}
	m.Categories = cats

	if baseModelID.Valid {
		m.BaseModelID = baseModelID.String
	}
	refs, err := s.listModelBaseAnnotations(m.ID)
	if err != nil {
		return nil, err
	}
	m.BaseAnnotations = refs

	return m, nil
}

func (s *ModelSQL) InsertModel(m *model.Model) error {
	if m == nil {
		return fmt.Errorf("model is nil")
	}
	if m.ID == "" {
		return fmt.Errorf("model id is empty")
	}
	if m.Type == "" {
		return fmt.Errorf("model type is empty")
	}
	if m.Location == "" {
		return fmt.Errorf("model location is empty")
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = m.CreatedAt
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
		INSERT INTO models (
			id, name, description, created_at, updated_at, type, location, algorithm_family, local_path, base_model_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, m.ID, m.Name, m.Description, m.CreatedAt, m.UpdatedAt, m.Type, m.Location, string(m.AlgorithmFamily), m.LocalPath, lo.Ternary[any](m.BaseModelID == "", nil, m.BaseModelID)); err != nil {
		return err
	}

	if err := s.insertModelCategoriesTx(tx, m.ID, m.Categories); err != nil {
		return err
	}

	if err := s.insertModelBaseAnnotationsTx(tx, m.ID, m.BaseAnnotations); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ModelSQL) UpdateModel(m *model.Model) error {
	if m == nil {
		return fmt.Errorf("model is nil")
	}
	if m.ID == "" {
		return fmt.Errorf("model id is empty")
	}
	if m.Type == "" {
		return fmt.Errorf("model type is empty")
	}
	if m.Location == "" {
		return fmt.Errorf("model location is empty")
	}

	m.UpdatedAt = time.Now()
	if m.CreatedAt.IsZero() {
		m.CreatedAt = m.UpdatedAt
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE models
		SET name = ?, description = ?, updated_at = ?, type = ?, location = ?, algorithm_family = ?, local_path = ?, base_model_id = ?
		WHERE id = ?
	`, m.Name, m.Description, m.UpdatedAt, m.Type, m.Location, string(m.AlgorithmFamily), m.LocalPath, lo.Ternary[any](m.BaseModelID == "", nil, m.BaseModelID), m.ID)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}

	// Delete categories, re-add
	if _, err := tx.Exec(`
		DELETE FROM model_categories
		WHERE model_id = ?
	`, m.ID); err != nil {
		return err
	}
	if err := s.insertModelCategoriesTx(tx, m.ID, m.Categories); err != nil {
		return err
	}

	// Delete base annotations, re-add
	if _, err := tx.Exec(`
		DELETE FROM models_base_annotations
		WHERE model_id = ?
	`, m.ID); err != nil {
		return err
	}
	if err := s.insertModelBaseAnnotationsTx(tx, m.ID, m.BaseAnnotations); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *ModelSQL) DeleteModel(id string) error {
	if id == "" {
		return fmt.Errorf("model id is empty")
	}

	res, err := s.db.Exec(`
		DELETE FROM models
		WHERE id = ?
	`, id)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *ModelSQL) listModelCategories(modelID string) ([]string, error) {
	rows, err := s.db.Query(`
		SELECT category
		FROM model_categories
		WHERE model_id = ?
		ORDER BY category ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cats, nil
}

func (s *ModelSQL) insertModelCategoriesTx(tx *sql.Tx, modelID string, categories []string) error {
	for _, c := range categories {
		if c == "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO model_categories (model_id, category)
			VALUES (?, ?)
		`, modelID, c); err != nil {
			return err
		}
	}
	return nil
}

func (s *ModelSQL) listModelBaseAnnotations(modelID string) ([]*model.AnnotationReference, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, annotation_id
		FROM models_base_annotations
		WHERE model_id = ?
		ORDER BY dataset_id ASC, annotation_id ASC
	`, modelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []*model.AnnotationReference
	for rows.Next() {
		r := &model.AnnotationReference{}
		if err := rows.Scan(&r.DatasetID, &r.ID); err != nil {
			return nil, err
		}
		refs = append(refs, r)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return refs, nil
}

func (s *ModelSQL) insertModelBaseAnnotationsTx(tx *sql.Tx, modelID string, refs []*model.AnnotationReference) error {
	for _, r := range refs {
		if r == nil {
			continue
		}
		if r.DatasetID == "" || r.ID == "" {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO models_base_annotations (model_id, dataset_id, annotation_id)
			VALUES (?, ?, ?)
		`, modelID, r.DatasetID, r.ID); err != nil {
			return err
		}
	}
	return nil
}
