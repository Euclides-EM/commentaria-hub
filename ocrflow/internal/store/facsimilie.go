package store

import (
	"database/sql"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
)

const FacsimileIDPrefix = "fac"

type FacsimileSQL struct {
	BaseSQL
}

func (s *FacsimileSQL) ListFacsimilesByEditionID(editionID string) ([]*ocrflow.Facsimile, error) {
	rows, err := s.db.Query(`
		SELECT id, url, main_text_pages, created_at, updated_at, name, description
		FROM facsimiles
		WHERE edition_id = ?
	`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facsimiles []*ocrflow.Facsimile

	for rows.Next() {
		f := &ocrflow.Facsimile{}
		if err := rows.Scan(
			&f.ID,
			&f.ScanURL,
			&f.MainTextPages,
			&f.CreatedAt,
			&f.UpdatedAt,
			&f.Name,
			&f.Description,
		); err != nil {
			return nil, err
		}
		facsimiles = append(facsimiles, f)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return facsimiles, nil
}

func (s *FacsimileSQL) GetFacsimileByID(editionKey string, facsimileID string) (*ocrflow.Facsimile, error) {
	f := &ocrflow.Facsimile{}

	err := s.db.QueryRow(`
		SELECT id, url, main_text_pages, created_at, updated_at, name, description
		FROM facsimiles
		WHERE edition_id = ? AND id = ?
	`, editionKey, facsimileID).Scan(
		&f.ID,
		&f.ScanURL,
		&f.MainTextPages,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.Name,
		&f.Description,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}

	return f, nil
}

func (s *FacsimileSQL) InsertFacsimile(editionId string, f *ocrflow.Facsimile) (*ocrflow.Facsimile, error) {
	_, err := s.db.Exec(`
		INSERT INTO facsimiles (id, edition_id, url, main_text_pages, created_at, updated_at, name, description, url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.ID, editionId, f.ScanURL, f.MainTextPages, f.CreatedAt, f.UpdatedAt, f.Name, f.Description, f.ScanURL)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func NewFacsimileSql(db *sql.DB) *FacsimileSQL {
	return &FacsimileSQL{
		BaseSQL{db: db},
	}
}
