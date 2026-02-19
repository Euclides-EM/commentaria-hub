package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
)

// FeatureExecutionSQL is the SQL store for feature executions (table: feature_executions, feature_execution_apply).
type FeatureExecutionSQL struct {
	db *sql.DB
}

// NewFeatureExecutionSQL returns a new FeatureExecutionSQL store using the given DB.
func NewFeatureExecutionSQL(db *sql.DB) *FeatureExecutionSQL {
	return &FeatureExecutionSQL{db: db}
}

func (s *FeatureExecutionSQL) List(datasetID string, featureIDs []string, statuses []feature.ExecutionStatus) ([]*feature.Execution, error) {
	query := `
		SELECT id, created_at, updated_at, name, description, dataset_id, annotation_id, keys, policy_skip_if, status
		FROM feature_executions
		WHERE 1=1
	`
	var args []any

	if len(statuses) > 0 {
		query += ` AND status IN (`
		for i, status := range statuses {
			if i > 0 {
				query += `, `
			}
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

	var out []*feature.Execution
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
					if item.Feature == fid && item.DatasetID == datasetID {
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

func (s *FeatureExecutionSQL) GetByID(id string) (*feature.Execution, error) {
	exec := &feature.Execution{}
	var keysJSON string
	var policySkipIf string
	err := s.db.QueryRow(`
		SELECT id, created_at, updated_at, name, description, dataset_id, annotation_id, keys, policy_skip_if, status
		FROM feature_executions
		WHERE id = ?
		LIMIT 1
	`, id).Scan(
		&exec.ID,
		&exec.CreatedAt,
		&exec.UpdatedAt,
		&exec.Name,
		&exec.Description,
		&exec.DatasetID,
		&exec.AnnotationID,
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

	if policySkipIf != "" && policySkipIf != "[]" {
		var skipIf []feature.ExecutionSkipIf
		if err := json.Unmarshal([]byte(policySkipIf), &skipIf); err != nil {
			return nil, fmt.Errorf("failed to parse skipIf: %w", err)
		}
		skipIf, err := parsePolicySkipIf(policySkipIf)
		if err != nil {
			return nil, err
		}
		exec.Policy = &feature.ExecutionPolicy{
			SkipIf: skipIf,
		}
	}

	apply, err := s.listExecutionApply(exec.ID)
	if err != nil {
		return nil, err
	}
	exec.Apply = apply

	return exec, nil
}

func (s *FeatureExecutionSQL) Create(exec *feature.Execution) error {
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
		policySkipIfBytes, err := json.Marshal(exec.Policy.SkipIf)
		if err != nil {
			return fmt.Errorf("failed to marshal policy skip_if: %w", err)
		}
		policySkipIf = string(policySkipIfBytes)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(`
		INSERT INTO feature_executions (id, created_at, updated_at, name, description, dataset_id, annotation_id, keys, policy_skip_if, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, exec.ID, exec.CreatedAt, exec.UpdatedAt, exec.Name, exec.Description, exec.DatasetID, exec.AnnotationID, keysJSON, policySkipIf, exec.Status)
	if err != nil {
		return err
	}

	if err := s.insertExecutionApplyTx(tx, exec.ID, exec.DatasetID, exec.AnnotationID, exec.Apply); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *FeatureExecutionSQL) UpdateStatus(id string, status feature.ExecutionStatus) error {
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

func scanFeatureExecution(scanner func(...any) error) (*feature.Execution, error) {
	exec := &feature.Execution{}
	var keysJSON string
	var policySkipIf string
	err := scanner(
		&exec.ID,
		&exec.CreatedAt,
		&exec.UpdatedAt,
		&exec.Name,
		&exec.Description,
		&exec.DatasetID,
		&exec.AnnotationID,
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

	if policySkipIf != "" && policySkipIf != "[]" {
		var skipIf []feature.ExecutionSkipIf
		if err := json.Unmarshal([]byte(policySkipIf), &skipIf); err != nil {
			return nil, fmt.Errorf("failed to parse skipIf: %w", err)
		}
		skipIf, err := parsePolicySkipIf(policySkipIf)
		if err != nil {
			return nil, err
		}
		exec.Policy = &feature.ExecutionPolicy{
			SkipIf: skipIf,
		}
	}

	return exec, nil
}

func parsePolicySkipIf(raw string) ([]feature.ExecutionSkipIf, error) {
	if raw == "" {
		return nil, nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if !strings.HasPrefix(raw, "[") {
		return []feature.ExecutionSkipIf{feature.ExecutionSkipIf(raw)}, nil
	}
	var skipIf []feature.ExecutionSkipIf
	if err := json.Unmarshal([]byte(raw), &skipIf); err != nil {
		return nil, fmt.Errorf("failed to parse policy skip_if JSON: %w", err)
	}
	return skipIf, nil
}

func (s *FeatureExecutionSQL) listExecutionApply(executionID string) ([]feature.ExecutionApplyItem, error) {
	rows, err := s.db.Query(`
		SELECT dataset_id, annotation_id, feature_id, revision_id
		FROM feature_execution_apply
		WHERE execution_id = ?
		ORDER BY sort_order ASC
	`, executionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apply []feature.ExecutionApplyItem
	for rows.Next() {
		var item feature.ExecutionApplyItem
		if err := rows.Scan(&item.DatasetID, &item.AnnotationId, &item.Feature, &item.Revision); err != nil {
			return nil, err
		}
		apply = append(apply, item)
	}
	return apply, rows.Err()
}

func (s *FeatureExecutionSQL) insertExecutionApplyTx(tx *sql.Tx, executionID, datasetID, annotationID string, apply []feature.ExecutionApplyItem) error {
	for i, item := range apply {
		_, err := tx.Exec(`
			INSERT INTO feature_execution_apply (execution_id, dataset_id, annotation_id, feature_id, revision_id, sort_order)
			VALUES (?, ?, ?, ?, ?, ?)
		`, executionID, datasetID, annotationID, item.Feature, item.Revision, i)
		if err != nil {
			return err
		}
	}
	return nil
}
