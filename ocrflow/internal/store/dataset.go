package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

const DatasetIDPrefix = "ds"

type DatasetSQL struct {
	BaseSQL
}

func NewDatasetSQL(db *sql.DB) *DatasetSQL {
	return &DatasetSQL{
		BaseSQL: BaseSQL{db: db},
	}
}

func (s *DatasetSQL) ListDatasets() ([]*model.Dataset, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed 
		FROM datasets
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasets []*model.Dataset
	for rows.Next() {
		d := &model.Dataset{}
		if err := rows.Scan(
			&d.ID,
			&d.Name,
			&d.Description,
			&d.CreatedAt,
			&d.UpdatedAt,
			&d.EditionID,
			&d.FacsimileID,
			&d.DPI,
			&d.Deskewed,
		); err != nil {
			return nil, err
		}
		datasets = append(datasets, d)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return datasets, nil
}

func (s *DatasetSQL) GetDataset(id string) (*model.Dataset, error) {
	d := &model.Dataset{}
	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed
		FROM datasets
		WHERE id = ?
	`, id).Scan(
		&d.ID,
		&d.Name,
		&d.Description,
		&d.CreatedAt,
		&d.UpdatedAt,
		&d.EditionID,
		&d.FacsimileID,
		&d.DPI,
		&d.Deskewed,
	)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DatasetSQL) InsertDataset(d *model.Dataset) error {
	_, err := s.db.Exec(`
		INSERT INTO datasets (id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		d.ID,
		d.Name,
		d.Description,
		d.CreatedAt,
		d.UpdatedAt,
		d.EditionID,
		d.FacsimileID,
		d.DPI,
		d.Deskewed,
	)
	return err
}

func (s *DatasetSQL) DeleteDataset(id string) error {
	_, err := s.db.Exec(`
		DELETE FROM datasets
		WHERE id = ?
	`, id)
	return err
}

func (s *DatasetSQL) UpdateDataset(ds *model.Dataset) error {
	ds.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE datasets
		SET name = ?, description = ?, updated_at = ?, edition_id = ?, facsimile_id = ?, dpi = ?, deskewed = ?
		WHERE id = ?
	`,
		ds.Name,
		ds.Description,
		ds.UpdatedAt,
		ds.EditionID,
		ds.FacsimileID,
		ds.DPI,
		ds.Deskewed,
		ds.ID,
	)
	return err
}
