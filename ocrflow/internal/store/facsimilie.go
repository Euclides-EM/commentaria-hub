package store

import (
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/samber/lo"
)

const FacsimileIDPrefix = "fac"

type FacsimileSQL struct {
	BaseSQL
}

func (s *FacsimileSQL) ListFacsimiles(editionIDs []string) ([]*model.Facsimile, error) {
	q := `SELECT id, edition_id, url, main_text_pages, created_at, updated_at, name, description FROM facsimiles`
	if len(editionIDs) > 0 {
		q += ` WHERE edition_id IN (%s)`
		q = fmt.Sprintf(q, strings.Join(slices.Repeat([]string{"?"}, len(editionIDs)), ", "))
	}
	rows, err := s.db.Query(q, lo.ToAnySlice(editionIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var facsimiles []*model.Facsimile
	for rows.Next() {
		f := &model.Facsimile{}
		if err = rows.Scan(
			&f.ID,
			&f.EditionID,
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
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return facsimiles, nil
}

func (s *FacsimileSQL) GetFacsimileByID(facsimileID string) (*model.Facsimile, error) {
	f := &model.Facsimile{}

	err := s.db.QueryRow(`
		SELECT id, edition_id, url, main_text_pages, created_at, updated_at, name, description
		FROM facsimiles
		WHERE id = ?
	`, facsimileID).Scan(
		&f.ID,
		&f.EditionID,
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

func (s *FacsimileSQL) InsertFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	_, err := s.db.Exec(`
		INSERT INTO facsimiles (id, edition_id, url, main_text_pages, created_at, updated_at, name, description, url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, f.ID, f.EditionID, f.ScanURL, f.MainTextPages, f.CreatedAt, f.UpdatedAt, f.Name, f.Description, f.ScanURL)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (s *FacsimileSQL) UpdateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	f.UpdatedAt = time.Now()
	_, err := s.db.Exec(`
		UPDATE facsimiles
		SET edition_id = ?, url = ?, main_text_pages = ?, updated_at = ?, name = ?, description = ?
		WHERE id = ?
	`, f.EditionID, f.ScanURL, f.MainTextPages, f.UpdatedAt, f.Name, f.Description, f.ID)
	if err != nil {
		return nil, err
	}
	return s.GetFacsimileByID(f.ID)
}

func (s *FacsimileSQL) DeleteFacsimile(id string) error {
	_, err := s.db.Exec(`
		DELETE FROM facsimiles
		WHERE id = ?
	`, id)
	return err
}

func NewFacsimileSql(db *sql.DB) *FacsimileSQL {
	return &FacsimileSQL{
		BaseSQL{db: db},
	}
}
