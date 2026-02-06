package featureplat

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// FeatureRevisionSQL is the SQL store for feature revisions (table: feature_revisions, feature_revision_features).
type FeatureRevisionSQL struct {
	db *sql.DB
}

// NewFeatureRevisionSQL returns a new FeatureRevisionSQL store using the given DB.
func NewFeatureRevisionSQL(db *sql.DB) *FeatureRevisionSQL {
	return &FeatureRevisionSQL{db: db}
}

func (s *FeatureRevisionSQL) ListByFeatureID(collectionID, featureID string) ([]*featureplat.FeatureRevision, error) {
	rows, err := s.db.Query(`
		SELECT collection_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type
		FROM feature_revisions
		WHERE collection_id = ? AND feature_id = ?
		ORDER BY updated_at DESC
	`, collectionID, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*featureplat.FeatureRevision
	for rows.Next() {
		rev, err := scanFeatureRevision(rows.Scan)
		if err != nil {
			return nil, err
		}
		// Load Features (references) if needed
		features, err := s.listRevisionFeatures(rev.CollectionID, rev.ID)
		if err != nil {
			return nil, err
		}
		rev.Features = features
		out = append(out, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FeatureRevisionSQL) GetByID(collectionID, featureID, revisionID string) (*featureplat.FeatureRevision, error) {
	rev := &featureplat.FeatureRevision{}
	err := s.db.QueryRow(`
		SELECT collection_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type
		FROM feature_revisions
		WHERE collection_id = ? AND id = ? AND feature_id = ?
		LIMIT 1
	`, collectionID, revisionID, featureID).Scan(
		&rev.CollectionID,
		&rev.ID,
		&rev.FeatureID,
		&rev.CreatedAt,
		&rev.UpdatedAt,
		&rev.Name,
		&rev.Description,
		&rev.Prompt,
		&rev.Regex,
		&rev.ExecutionStrategy,
		&rev.Note,
		&rev.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureID, revisionID)
		}
		return nil, err
	}
	features, err := s.listRevisionFeatures(rev.CollectionID, rev.ID)
	if err != nil {
		return nil, err
	}
	rev.Features = features
	return rev, nil
}

func (s *FeatureRevisionSQL) Create(collectionID, featureID string, rev *featureplat.FeatureRevision) error {
	if rev == nil {
		return fmt.Errorf("feature revision is nil")
	}
	if rev.ID == "" {
		return fmt.Errorf("feature revision id is empty")
	}
	rev.CollectionID = collectionID
	rev.FeatureID = featureID
	if rev.CreatedAt.IsZero() {
		rev.CreatedAt = time.Now()
	}
	if rev.UpdatedAt.IsZero() {
		rev.UpdatedAt = rev.CreatedAt
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO feature_revisions (collection_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rev.CollectionID, rev.ID, featureID, rev.CreatedAt, rev.UpdatedAt, rev.Name, rev.Description, rev.Prompt, rev.Regex, rev.ExecutionStrategy, rev.Note, rev.Type)
	if err != nil {
		return err
	}

	if err := s.insertRevisionFeaturesTx(tx, collectionID, rev.ID, rev.Features); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureRevisionSQL) Update(collectionID, featureID, revisionID string, rev *featureplat.FeatureRevision) error {
	if rev == nil {
		return fmt.Errorf("feature revision is nil")
	}
	rev.CollectionID = collectionID
	rev.UpdatedAt = time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE feature_revisions
		SET name = ?, description = ?, prompt = ?, regex = ?, execution_strategy = ?, note = ?, type = ?, updated_at = ?
		WHERE collection_id = ? AND id = ? AND feature_id = ?
	`, rev.Name, rev.Description, rev.Prompt, rev.Regex, rev.ExecutionStrategy, rev.Note, rev.Type, rev.UpdatedAt, collectionID, revisionID, featureID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureID, revisionID)
	}

	// Update Features references
	if _, err := tx.Exec(`DELETE FROM feature_revision_features WHERE collection_id = ? AND feature_revision_id = ?`, collectionID, revisionID); err != nil {
		return err
	}
	if err := s.insertRevisionFeaturesTx(tx, collectionID, revisionID, rev.Features); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureRevisionSQL) Delete(collectionID, featureID, revisionID string) error {
	res, err := s.db.Exec(`DELETE FROM feature_revisions WHERE collection_id = ? AND id = ? AND feature_id = ?`, collectionID, revisionID, featureID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureID, revisionID)
	}
	return nil
}

func scanFeatureRevision(scanner func(...any) error) (*featureplat.FeatureRevision, error) {
	rev := &featureplat.FeatureRevision{}
	err := scanner(
		&rev.CollectionID,
		&rev.ID,
		&rev.FeatureID,
		&rev.CreatedAt,
		&rev.UpdatedAt,
		&rev.Name,
		&rev.Description,
		&rev.Prompt,
		&rev.Regex,
		&rev.ExecutionStrategy,
		&rev.Note,
		&rev.Type,
	)
	if err != nil {
		return nil, err
	}
	return rev, nil
}

func (s *FeatureRevisionSQL) listRevisionFeatures(collectionID, revisionID string) ([]featureplat.Reference, error) {
	rows, err := s.db.Query(`
		SELECT feature_id
		FROM feature_revision_features
		WHERE collection_id = ? AND feature_revision_id = ?
		ORDER BY sort_order ASC
	`, collectionID, revisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []featureplat.Reference
	for rows.Next() {
		var ref featureplat.Reference
		if err := rows.Scan(&ref.ID); err != nil {
			return nil, err
		}
		refs = append(refs, ref)
	}
	return refs, rows.Err()
}

func (s *FeatureRevisionSQL) insertRevisionFeaturesTx(tx *sql.Tx, collectionID, revisionID string, features []featureplat.Reference) error {
	for i, ref := range features {
		_, err := tx.Exec(`
			INSERT INTO feature_revision_features (collection_id, feature_revision_id, feature_id, sort_order)
			VALUES (?, ?, ?, ?)
		`, collectionID, revisionID, ref.ID, i)
		if err != nil {
			return err
		}
	}
	return nil
}
