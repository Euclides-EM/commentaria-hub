package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
)

const DatasetIDPrefix = "ds"

type DatasetSQL struct {
	BaseSQL
	FilesysManager *filesys.Manager
}

func NewDatasetSQL(db *sql.DB, filesysManager *filesys.Manager) *DatasetSQL {
	return &DatasetSQL{
		BaseSQL:        BaseSQL{db: db},
		FilesysManager: filesysManager,
	}
}

func (s *DatasetSQL) ListDatasets() ([]*model.Dataset, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed, denoised, pages, status, creation_error
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
			&d.Denoised,
			&d.Pages,
			&d.Status,
			&d.CreationError,
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
		SELECT id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed, denoised, pages, status, creation_error
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
		&d.Denoised,
		&d.Pages,
		&d.Status,
		&d.CreationError,
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
	status := d.Status
	if status == "" {
		status = model.DatasetStatusReady
	}
	_, err := s.db.Exec(`
		INSERT INTO datasets (id, name, description, created_at, updated_at, edition_id, facsimile_id, dpi, deskewed, denoised, pages, status, creation_error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		d.Denoised,
		d.Pages,
		status,
		d.CreationError,
	)
	return err
}

// UpdateDatasetCreationStatus sets status and optional creation_error (e.g. after async creation completes).
func (s *DatasetSQL) UpdateDatasetCreationStatus(id, status, creationError string) error {
	_, err := s.db.Exec(`
		UPDATE datasets SET status = ?, creation_error = ?, updated_at = ?
		WHERE id = ?
	`, status, creationError, time.Now(), id)
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
		SET name = ?, description = ?, updated_at = ?, edition_id = ?, facsimile_id = ?, dpi = ?, deskewed = ?, denoised = ?, pages = ?, status = ?, creation_error = ?
		WHERE id = ?
	`,
		ds.Name,
		ds.Description,
		ds.UpdatedAt,
		ds.EditionID,
		ds.FacsimileID,
		ds.DPI,
		ds.Deskewed,
		ds.Denoised,
		ds.Pages,
		ds.Status,
		ds.CreationError,
		ds.ID,
	)
	return err
}
