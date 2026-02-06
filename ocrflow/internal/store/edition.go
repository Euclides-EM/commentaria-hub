package store

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
)

type EditionSQL struct {
	BaseSQL
}

func (s *EditionSQL) ListEditions() ([]*ocrflow.Edition, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM editions
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var editions []*ocrflow.Edition

	for rows.Next() {
		e := &ocrflow.Edition{}
		if err := rows.Scan(
			&e.ID,
			&e.Name,
			&e.Description,
			&e.CreatedAt,
			&e.UpdatedAt,
		); err != nil {
			return nil, err
		}
		editions = append(editions, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return editions, nil
}

func (s *EditionSQL) GetEditionByID(id string) (*ocrflow.Edition, error) {
	e := &ocrflow.Edition{}

	err := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM editions
		WHERE id = ?
	`, id).Scan(
		&e.ID,
		&e.Name,
		&e.Description,
		&e.CreatedAt,
		&e.UpdatedAt,
	)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return e, nil
}

func (s *EditionSQL) InsertEdition(edition *ocrflow.Edition) error {
	now := time.Now().UTC()

	edition.CreatedAt = now
	edition.UpdatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO editions (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`,
		edition.ID,
		edition.Name,
		edition.Description,
		edition.CreatedAt,
		edition.UpdatedAt,
	)

	return err
}

func (s *EditionSQL) UpdateEdition(edition *ocrflow.Edition) error {
	edition.UpdatedAt = time.Now().UTC()

	_, err := s.db.Exec(`
		UPDATE editions
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`,
		edition.Name,
		edition.Description,
		edition.UpdatedAt,
		edition.ID,
	)

	return err
}

func NewEditionSQL(db *sql.DB) *EditionSQL {
	return &EditionSQL{
		BaseSQL{db: db},
	}
}
