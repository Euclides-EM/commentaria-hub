package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

const AnnotationGroupIDPrefix = "anng"

type AnnotationGroupSql struct {
	BaseSQL
}

func NewAnnotationGroupSQL(db *sql.DB) *AnnotationGroupSql {
	return &AnnotationGroupSql{
		BaseSQL: BaseSQL{db: db},
	}
}

func (s *AnnotationGroupSql) List() ([]*annotation.Group, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at
		FROM annotation_groups
		ORDER BY created_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*annotation.Group
	for rows.Next() {
		g := &annotation.Group{}
		if err := rows.Scan(
			&g.ID,
			&g.Name,
			&g.Description,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		ids, err := s.listAnnotationIDs(g.ID)
		if err != nil {
			return nil, err
		}
		g.Annotations = ids
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *AnnotationGroupSql) Get(id string) (*annotation.Group, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at
		FROM annotation_groups
		WHERE id = ?
		LIMIT 1
	`, id)

	g := &annotation.Group{}
	err := row.Scan(
		&g.ID,
		&g.Name,
		&g.Description,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	ids, err := s.listAnnotationIDs(g.ID)
	if err != nil {
		return nil, err
	}
	g.Annotations = ids
	return g, nil
}

func (s *AnnotationGroupSql) listAnnotationIDs(groupID string) ([]*annotation.Reference, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, annotation_id
		FROM annotation_group_annotations
		WHERE group_id = ?
		ORDER BY annotation_id
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var annRefs []*annotation.Reference
	for rows.Next() {
		var annRef annotation.Reference
		if err := rows.Scan(&annRef.DatasetID, &annRef.ID); err != nil {
			return nil, err
		}
		annRefs = append(annRefs, &annRef)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return annRefs, nil
}

func (s *AnnotationGroupSql) Create(group *annotation.Group) (*annotation.Group, error) {
	if group == nil {
		return nil, fmt.Errorf("annotation group is nil")
	}
	if group.ID == "" {
		group.ID = idgen.GenerateID(AnnotationGroupIDPrefix)
	}

	now := time.Now()
	if group.CreatedAt.IsZero() {
		group.CreatedAt = now
	}
	group.UpdatedAt = now

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(`
		INSERT INTO annotation_groups (id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, group.ID, group.Name, group.Description, group.CreatedAt, group.UpdatedAt); err != nil {
		return nil, err
	}

	for _, annRef := range group.Annotations {
		if annRef == nil {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO annotation_group_annotations (group_id, dataset_id, annotation_id)
			VALUES (?, ?, ?)
		`, group.ID, annRef.DatasetID, annRef.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *AnnotationGroupSql) Update(group *annotation.Group) (*annotation.Group, error) {
	if group == nil {
		return nil, fmt.Errorf("annotation group is nil")
	}
	if group.ID == "" {
		return nil, fmt.Errorf("annotation group id is empty")
	}

	group.UpdatedAt = time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE annotation_groups
		SET name = ?, description = ?, updated_at = ?
		WHERE id = ?
	`, group.Name, group.Description, group.UpdatedAt, group.ID)
	if err != nil {
		return nil, err
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if aff == 0 {
		return nil, sql.ErrNoRows
	}

	if _, err := tx.Exec(`
		DELETE FROM annotation_group_annotations
		WHERE group_id = ?
	`, group.ID); err != nil {
		return nil, err
	}

	for _, annRef := range group.Annotations {
		if annRef == nil {
			continue
		}
		if _, err := tx.Exec(`
			INSERT INTO annotation_group_annotations (group_id, dataset_id, annotation_id)
			VALUES (?, ?, ?)
		`, group.ID, annRef.DatasetID, annRef.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *AnnotationGroupSql) Delete(id string) error {
	_, err := s.db.Exec(`
		DELETE FROM annotation_groups
		WHERE id = ?
	`, id)
	return err
}
