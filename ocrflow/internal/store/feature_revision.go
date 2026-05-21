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

func (s *FeatureRevisionSQL) ListByFeatureID(featureID string) ([]*feature.Revision, error) {
	if featureID == "" {
		return nil, errors.New("list feature revisions: missing feature_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, feature_id, prompt, categorizer
FROM feature_revisions
WHERE feature_id = ?
ORDER BY created_at DESC
`
	rows, err := s.db.Query(q, featureID)
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

func (s *FeatureRevisionSQL) ListByFeatureIDInScope(scope feature.DefScope, featureID string) ([]*feature.Revision, error) {
	if featureID == "" {
		return nil, errors.New("list feature revisions: missing feature_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, feature_id, prompt, categorizer
FROM feature_revisions
WHERE scope = ? AND dataset_id = ? AND feature_id = ?
ORDER BY created_at DESC
`
	rows, err := s.db.Query(q, scope.Type, scope.DatasetID, featureID)
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

func (s *FeatureRevisionSQL) GetByID(featureID, revisionID string) (*feature.Revision, error) {
	if featureID == "" || revisionID == "" {
		return nil, errors.New("get feature revision: missing feature_id or revision_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, feature_id, prompt, categorizer
FROM feature_revisions
WHERE feature_id = ? AND id = ?
LIMIT 1
`
	row := s.db.QueryRow(q, featureID, revisionID)
	rev, err := scanFeatureRevision(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature revision not found: feature_id=%s revision_id=%s: %w", featureID, revisionID, err)
		}
		return nil, fmt.Errorf("get feature revision: %w", err)
	}
	return rev, nil
}

func (s *FeatureRevisionSQL) GetByIDInScope(scope feature.DefScope, featureID, revisionID string) (*feature.Revision, error) {
	if featureID == "" || revisionID == "" {
		return nil, errors.New("get feature revision: missing feature_id or revision_id")
	}

	const q = `
SELECT
  id, name, description, created_at, updated_at,
  dataset_id, scope, feature_id, prompt, categorizer
FROM feature_revisions
WHERE scope = ? AND dataset_id = ? AND feature_id = ? AND id = ?
LIMIT 1
`
	row := s.db.QueryRow(q, scope.Type, scope.DatasetID, featureID, revisionID)
	rev, err := scanFeatureRevision(row.Scan)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("feature revision not found: scope=%v feature_id=%s revision_id=%s: %w", scope, featureID, revisionID, err)
		}
		return nil, fmt.Errorf("get feature revision: %w", err)
	}
	return rev, nil
}

func (s *FeatureRevisionSQL) Create(rev *feature.Revision) error {
	const q = `
INSERT INTO feature_revisions (
  id, name, description, dataset_id, scope, feature_id,
  prompt, categorizer,
  created_at, updated_at
) VALUES (
  ?, ?, ?, ?, ?, ?,
  ?, ?,
  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
)
`
	_, err := s.db.Exec(
		q,
		rev.ID,
		rev.Name,
		rev.Description,
		rev.Scope.DatasetID,
		rev.Scope.Type,
		rev.FeatureID,
		rev.Prompt,
		rev.Categorizer,
	)
	if err != nil {
		return fmt.Errorf("create feature revision: %w", err)
	}

	// Refresh timestamps
	created, err := s.GetByIDInScope(rev.Scope, rev.FeatureID, rev.ID)
	if err == nil && created != nil {
		rev.CreatedAt = created.CreatedAt
		rev.UpdatedAt = created.UpdatedAt
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
		datasetID   sql.NullString
		scopeType   feature.ScopeType
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
		&scopeType,
		&featureID,
		&prompt,
		&categorizer,
	); err != nil {
		return nil, err
	}

	var scp feature.DefScope
	switch scopeType {
	case feature.ScopeTypeEditions:
		scp = feature.NewEditionDefScope()
	case feature.ScopeTypeDataset:
		scp = feature.NewDatasetDefScope(datasetID.String)
	default:
		return nil, fmt.Errorf("invalid scope type: %s", scopeType)
	}
	return &feature.Revision{
		Meta: common.Meta{
			ID:          id,
			Name:        name,
			Description: desc,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
		Scope:       scp,
		FeatureID:   featureID,
		Prompt:      prompt,
		Categorizer: categorizer,
	}, nil
}
