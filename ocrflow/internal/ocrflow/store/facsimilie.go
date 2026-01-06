package store

import (
	"database/sql"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
)

type FacsimileSQL struct {
	BaseSQL
}

func (s *FacsimileSQL) ListFacsimilesByEditionID(editionID string) ([]*model.Facsimile, error) {
	rows, err := s.db.Query(`
		SELECT id, url, created_at, updated_at, name, description
		FROM facsimiles
		WHERE edition_id = ?
	`, editionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var facsimiles []*model.Facsimile

	for rows.Next() {
		f := &model.Facsimile{}
		if err := rows.Scan(
			&f.ID,
			&f.ScanURL,
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

func (s *FacsimileSQL) GetFacsimileByID(editionKey string, facsimileID string) (*model.Facsimile, error) {
	f := &model.Facsimile{}

	err := s.db.QueryRow(`
		SELECT id, url, created_at, updated_at, name, description
		FROM facsimiles
		WHERE edition_id = ? AND id = ?
	`, editionKey, facsimileID).Scan(
		&f.ID,
		&f.ScanURL,
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

func NewFacsimileSql(db *sql.DB) *FacsimileSQL {
	return &FacsimileSQL{
		BaseSQL{db: db},
	}
}
