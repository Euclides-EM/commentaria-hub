package featureplat

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
)

// FeatureExecutionSQL is the SQL store for feature executions (table: feature_executions, feature_execution_apply).
type FeatureExecutionSQL struct {
	db *sql.DB
}

// NewFeatureExecutionSQL returns a new FeatureExecutionSQL store using the given DB.
func NewFeatureExecutionSQL(db *sql.DB) *FeatureExecutionSQL {
	return &FeatureExecutionSQL{db: db}
}

func (s *FeatureExecutionSQL) List(collectionId string, featureIDs []string, statuses []featureplat.FeatureExecutionStatus) ([]*featureplat.FeatureExecution, error) {
	query := `
		SELECT id, created_at, updated_at, name, description, keys, policy_skip_if, status
		FROM feature_executions
	`
	var args []any
	argIdx := 0

	if len(statuses) > 0 {
		query += ` AND status IN (`
		for i, status := range statuses {
			if i > 0 {
				query += `, `
			}
			argIdx++
			query += `?`
			args = append(args, status)
		}
		query += `)`
	}

	query += ` ORDER BY updated_at DESC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*featureplat.FeatureExecution
	for rows.Next() {
		exec, err := scanFeatureExecution(rows.Scan)
		if err != nil {
			return nil, err
		}
		// Load Apply items
		apply, err := s.listExecutionApply(exec.ID)
		if err != nil {
			return nil, err
		}
		exec.Apply = apply

		// Filter by featureIDs if provided (check if any Apply item's Feature is in featureIDs)
		if len(featureIDs) > 0 {
			hasMatchingFeature := false
			for _, item := range exec.Apply {
				for _, fid := range featureIDs {
					if item.Feature == fid && item.Collection == collectionId {
						hasMatchingFeature = true
						break
					}
				}
				if hasMatchingFeature {
					break
				}
			}
			if !hasMatchingFeature {
				continue
			}
		}

		out = append(out, exec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *FeatureExecutionSQL) GetByID(id string) (*featureplat.FeatureExecution, error) {
	exec := &featureplat.FeatureExecution{}
	var keysJSON string
	var policySkipIf string
	err := s.db.QueryRow(`
		SELECT id, created_at, updated_at, name, description, collection, keys, policy_skip_if, status
		FROM feature_executions
		WHERE id = ?
		LIMIT 1
	`, id).Scan(
		&exec.ID,
		&exec.CreatedAt,
		&exec.UpdatedAt,
		&exec.Name,
		&exec.Description,
		&exec.Collection,
		&keysJSON,
		&policySkipIf,
		&exec.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("feature execution not found")
		}
		return nil, err
	}

	if keysJSON != "" && keysJSON != "[]" {
		if err := json.Unmarshal([]byte(keysJSON), &exec.Keys); err != nil {
			return nil, fmt.Errorf("failed to parse keys JSON: %w", err)
		}
	}

	if policySkipIf != "" {
		exec.Policy = &featureplat.FeatureExecutionPolicy{
			SkipIf: featureplat.FeatureExecutionSkipIf(policySkipIf),
		}
	}

	apply, err := s.listExecutionApply(exec.ID)
	if err != nil {
		return nil, err
	}
	exec.Apply = apply

	return exec, nil
}

func (s *FeatureExecutionSQL) Create(exec *featureplat.FeatureExecution) error {
	if exec == nil {
		return fmt.Errorf("feature execution is nil")
	}
	if exec.ID == "" {
		return fmt.Errorf("feature execution id is empty")
	}
	if exec.CreatedAt.IsZero() {
		exec.CreatedAt = time.Now()
	}
	if exec.UpdatedAt.IsZero() {
		exec.UpdatedAt = exec.CreatedAt
	}

	keysJSON := "[]"
	if len(exec.Keys) > 0 {
		keysBytes, err := json.Marshal(exec.Keys)
		if err != nil {
			return fmt.Errorf("failed to marshal keys: %w", err)
		}
		keysJSON = string(keysBytes)
	}

	policySkipIf := ""
	if exec.Policy != nil {
		policySkipIf = string(exec.Policy.SkipIf)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO feature_executions (id, created_at, updated_at, name, description, collection, keys, policy_skip_if, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, exec.CreatedAt, exec.UpdatedAt, exec.Name, exec.Description, exec.Collection, keysJSON, policySkipIf, exec.Status)
	if err != nil {
		return err
	}

	if err := s.insertExecutionApplyTx(tx, exec.ID, exec.Collection, exec.Apply); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureExecutionSQL) UpdateStatus(id string, status featureplat.FeatureExecutionStatus) error {
	res, err := s.db.Exec(`
		UPDATE feature_executions
		SET status = ?, updated_at = ?
		WHERE id = ?
	`, status, time.Now(), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("feature execution not found")
	}
	return nil
}

func scanFeatureExecution(scanner func(...any) error) (*featureplat.FeatureExecution, error) {
	exec := &featureplat.FeatureExecution{}
	var keysJSON string
	var policySkipIf string
	err := scanner(
		&exec.ID,
		&exec.CreatedAt,
		&exec.UpdatedAt,
		&exec.Name,
		&exec.Description,
		&exec.Collection,
		&keysJSON,
		&policySkipIf,
		&exec.Status,
	)
	if err != nil {
		return nil, err
	}

	if keysJSON != "" && keysJSON != "[]" {
		if err := json.Unmarshal([]byte(keysJSON), &exec.Keys); err != nil {
			return nil, fmt.Errorf("failed to parse keys JSON: %w", err)
		}
	}

	if policySkipIf != "" {
		exec.Policy = &featureplat.FeatureExecutionPolicy{
			SkipIf: featureplat.FeatureExecutionSkipIf(policySkipIf),
		}
	}

	return exec, nil
}

func (s *FeatureExecutionSQL) listExecutionApply(executionID string) ([]featureplat.FeatureExecutionApplyItem, error) {
	rows, err := s.db.Query(`
		SELECT collection_id, feature_id, revision_id
		FROM feature_execution_apply
		WHERE execution_id = ?
		ORDER BY sort_order ASC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apply []featureplat.FeatureExecutionApplyItem
	for rows.Next() {
		var item featureplat.FeatureExecutionApplyItem
		if err := rows.Scan(&item.Collection, &item.Feature, &item.Revision); err != nil {
			return nil, err
		}
		apply = append(apply, item)
	}
	return apply, rows.Err()
}

func (s *FeatureExecutionSQL) insertExecutionApplyTx(tx *sql.Tx, executionID, collectionID string, apply []featureplat.FeatureExecutionApplyItem) error {
	for i, item := range apply {
		_, err := tx.Exec(`
			INSERT INTO feature_execution_apply (execution_id, collection_id, feature_id, revision_id, sort_order)
			VALUES (?, ?, ?, ?, ?)
		`, executionID, collectionID, item.Feature, item.Revision, i)
		if err != nil {
			return err
		}
	}
	return nil
}
