package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureRevisionSQL is the SQL store for feature revisions (table: feature_revisions).
type FeatureRevisionSQL struct {
	db *sql.DB
}

// NewFeatureRevisionSQL returns a new FeatureRevisionSQL store using the given DB.
func NewFeatureRevisionSQL(db *sql.DB) *FeatureRevisionSQL {
	return &FeatureRevisionSQL{db: db}
}

func (s *FeatureRevisionSQL) ListByFeatureID(datasetID, featureID string) ([]*feature.Revision, error) {
	if datasetID == "" || featureID == "" {
		return nil, errors.New("list feature revisions: missing dataset_id or feature_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, feature_id, prompt, categorizer
FROM feature_revisions
WHERE dataset_id = ? AND feature_id = ?
ORDER BY created_at DESC
`
	rows, err := s.db.Query(q, datasetID, featureID)
	if err != nil {
		return nil, fmt.Errorf("list feature revisions: %w", err)
	}
	defer rows.Close()

	var out []*feature.Revision
	for rows.Next() {
		rev, err := scanFeatureRevision(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("list feature revisions scan: %w", err)
		}
		out = append(out, rev)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list feature revisions rows: %w", err)
	}
	return out, nil
}

func (s *FeatureRevisionSQL) GetByID(datasetID, featureID, revisionID string) (*feature.Revision, error) {
	if datasetID == "" || featureID == "" || revisionID == "" {
		return nil, errors.New("get feature revision: missing dataset_id, feature_id, or revision_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, feature_id, prompt, categorizer
FROM feature_revisions
WHERE dataset_id = ? AND feature_id = ? AND id = ?
LIMIT 1
`
	row := s.db.QueryRow(q, datasetID, featureID, revisionID)
	rev, err := scanFeatureRevision(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature revision not found: dataset_id=%s feature_id=%s revision_id=%s: %w", datasetID, featureID, revisionID, err)
		}
		return nil, fmt.Errorf("get feature revision: %w", err)
	}
	return rev, nil
}

func (s *FeatureRevisionSQL) Create(datasetID, featureID string, rev *feature.Revision) error {
	if rev == nil {
		return errors.New("create feature revision: nil revision")
	}
	if datasetID == "" || featureID == "" {
		return errors.New("create feature revision: missing dataset_id or feature_id")
	}
	if rev.ID == "" {
		return errors.New("create feature revision: missing id")
	}

	// Ensure DB truth for these
	rev.DatasetID = datasetID
	rev.FeatureID = featureID

	const q = `
INSERT INTO feature_revisions (
  id, name, description, dataset_id, feature_id,
  prompt, categorizer,
  created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?,
  ?, ?,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
`
	_, err := s.db.Exec(
		q,
		rev.ID,
		rev.Name,
		rev.Description,
		rev.DatasetID,
		rev.FeatureID,
		rev.Prompt,
		rev.Categorizer,
	)
	if err != nil {
		return fmt.Errorf("create feature revision: %w", err)
	}

	// Refresh timestamps
	created, err := s.GetByID(datasetID, featureID, rev.ID)
	if err == nil && created != nil {
		rev.CreatedAt = created.CreatedAt
		rev.UpdatedAt = created.UpdatedAt
	}
	return nil
}

func (s *FeatureRevisionSQL) Update(datasetID, featureID, revisionID string, rev *feature.Revision) error {
	if rev == nil {
		return errors.New("update feature revision: nil revision")
	}
	if datasetID == "" || featureID == "" || revisionID == "" {
		return errors.New("update feature revision: missing dataset_id, feature_id, or revision_id")
	}

	const q = `
UPDATE feature_revisions
SET
  name = ?,
  description = ?,
  prompt = ?,
  categorizer = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE dataset_id = ? AND feature_id = ? AND id = ?
`
	res, err := s.db.Exec(
		q,
		rev.Name,
		rev.Description,
		rev.Prompt,
		rev.Categorizer,
		datasetID,
		featureID,
		revisionID,
	)
	if err != nil {
		return fmt.Errorf("update feature revision: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update feature revision: rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("feature revision not found: dataset_id=%s feature_id=%s revision_id=%s", datasetID, featureID, revisionID)
	}

	// Keep caller pointer consistent
	rev.ID = revisionID
	rev.DatasetID = datasetID
	rev.FeatureID = featureID

	updated, err := s.GetByID(datasetID, featureID, revisionID)
	if err == nil && updated != nil {
		rev.CreatedAt = updated.CreatedAt
		rev.UpdatedAt = updated.UpdatedAt
	}
	return nil
}

func (s *FeatureRevisionSQL) Delete(datasetID, featureID, revisionID string) error {
	if datasetID == "" || featureID == "" || revisionID == "" {
		return errors.New("delete feature revision: missing dataset_id, feature_id, or revision_id")
	}

	const q = `
DELETE FROM feature_revisions
WHERE dataset_id = ? AND feature_id = ? AND id = ?
`
	res, err := s.db.Exec(q, datasetID, featureID, revisionID)
	if err != nil {
		return fmt.Errorf("delete feature revision: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete feature revision: rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("feature revision not found: dataset_id=%s feature_id=%s revision_id=%s", datasetID, featureID, revisionID)
	}
	return nil
}

func scanFeatureRevision(scanner func(...any) error) (*feature.Revision, error) {
	var (
		id          string
		name        string
		desc        string
		createdAt   time.Time
		updatedAt   time.Time
		datasetID   string
		featureID   string
		prompt      string
		categorizer string
	)

	if err := scanner(
		&id,
		&name,
		&desc,
		&createdAt,
		&updatedAt,
		&datasetID,
		&featureID,
		&prompt,
		&categorizer,
	); err != nil {
		return nil, err
	}

	return &feature.Revision{
		Meta: common.Meta{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		DatasetID:   datasetID,
		FeatureID:   featureID,
		Prompt:      prompt,
		Categorizer: categorizer,
	}, nil
}
