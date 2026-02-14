package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureRevisionSQL is the SQL store for feature revisions (table: feature_revisions, feature_revision_features).
type FeatureRevisionSQL struct {
	db *sql.DB
}

// NewFeatureRevisionSQL returns a new FeatureRevisionSQL store using the given DB.
func NewFeatureRevisionSQL(db *sql.DB) *FeatureRevisionSQL {
	return &FeatureRevisionSQL{db: db}
}

func (s *FeatureRevisionSQL) ListByFeatureID(datasetID, featureID string) ([]*feature.Revision, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type
		FROM feature_revisions
		WHERE dataset_id = ? AND feature_id = ?
		ORDER BY updated_at DESC
	`, datasetID, featureID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*feature.Revision
	for rows.Next() {
		rev, err := scanFeatureRevision(rows.Scan)
		if err != nil {
			return nil, err
		}
		// Load Features (references) if needed
		features, err := s.listRevisionFeatures(rev.DatasetID, rev.ID)
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

func (s *FeatureRevisionSQL) GetByID(datasetID, featureID, revisionID string) (*feature.Revision, error) {
	rev := &feature.Revision{}
	err := s.db.QueryRow(`
		SELECT dataset_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type
		FROM feature_revisions
		WHERE dataset_id = ? AND id = ? AND feature_id = ?
		LIMIT 1
	`, datasetID, revisionID, featureID).Scan(
		&rev.DatasetID,
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
	features, err := s.listRevisionFeatures(rev.DatasetID, rev.ID)
	if err != nil {
		return nil, err
	}
	rev.Features = features
	return rev, nil
}

func (s *FeatureRevisionSQL) Create(datasetID, featureID string, rev *feature.Revision) error {
	if rev == nil {
		return fmt.Errorf("feature revision is nil")
	}
	if rev.ID == "" {
		return fmt.Errorf("feature revision id is empty")
	}
	rev.DatasetID = datasetID
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
		INSERT INTO feature_revisions (dataset_id, id, feature_id, created_at, updated_at, name, description, prompt, regex, execution_strategy, note, type)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rev.DatasetID, rev.ID, featureID, rev.CreatedAt, rev.UpdatedAt, rev.Name, rev.Description, rev.Prompt, rev.Regex, rev.ExecutionStrategy, rev.Note, rev.Type)
	if err != nil {
		return err
	}

	if err := s.insertRevisionFeaturesTx(tx, datasetID, rev.ID, rev.Features); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureRevisionSQL) Update(datasetID, featureID, revisionID string, rev *feature.Revision) error {
	if rev == nil {
		return fmt.Errorf("feature revision is nil")
	}
	rev.DatasetID = datasetID
	rev.UpdatedAt = time.Now()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.Exec(`
		UPDATE feature_revisions
		SET name = ?, description = ?, prompt = ?, regex = ?, execution_strategy = ?, note = ?, type = ?, updated_at = ?
		WHERE dataset_id = ? AND id = ? AND feature_id = ?
	`, rev.Name, rev.Description, rev.Prompt, rev.Regex, rev.ExecutionStrategy, rev.Note, rev.Type, rev.UpdatedAt, datasetID, revisionID, featureID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureID, revisionID)
	}

	// Update Features references
	if _, err := tx.Exec(`DELETE FROM feature_revision_features WHERE dataset_id = ? AND feature_revision_id = ?`, datasetID, revisionID); err != nil {
		return err
	}
	if err := s.insertRevisionFeaturesTx(tx, datasetID, revisionID, rev.Features); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureRevisionSQL) Delete(datasetID, featureID, revisionID string) error {
	res, err := s.db.Exec(`DELETE FROM feature_revisions WHERE dataset_id = ? AND id = ? AND feature_id = ?`, datasetID, revisionID, featureID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureID, revisionID)
	}
	return nil
}

func scanFeatureRevision(scanner func(...any) error) (*feature.Revision, error) {
	rev := &feature.Revision{}
	err := scanner(
		&rev.DatasetID,
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

func (s *FeatureRevisionSQL) listRevisionFeatures(datasetID, revisionID string) ([]common.Reference, error) {
	rows, err := s.db.Query(`
		SELECT feature_id
		FROM feature_revision_features
		WHERE dataset_id = ? AND feature_revision_id = ?
		ORDER BY sort_order ASC
	`, datasetID, revisionID)
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

func (s *FeatureRevisionSQL) insertRevisionFeaturesTx(tx *sql.Tx, datasetID, revisionID string, features []common.Reference) error {
	for i, ref := range features {
		_, err := tx.Exec(`
			INSERT INTO feature_revision_features (dataset_id, feature_revision_id, feature_id, sort_order)
			VALUES (?, ?, ?, ?)
		`, datasetID, revisionID, ref.ID, i)
		if err != nil {
			return err
		}
	}
	return nil
}
