package store

import (
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/annotationrule"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
)

type AnnotationSQL struct {
	BaseSQL
}

func NewAnnotationSQL(db *sql.DB) *AnnotationSQL {
	return &AnnotationSQL{
		BaseSQL: BaseSQL{db: db},
	}
}

func (s *AnnotationSQL) GetAnnotation(datasetID, id string) (*model.Annotation, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id
		FROM annotations
		WHERE dataset_id = ? AND id = ?
		LIMIT 1
	`, datasetID, id)

	a := &model.Annotation{}
	err := row.Scan(
		&a.ID,
		&a.Name,
		&a.Description,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.Pages,
		&a.Segmented,
		&a.GroundTruth,
		&a.Ocred,
		&a.DatasetID,
		&a.OriginAnnotationID,
	)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	rules, err := s.listAppliedRules(a.ID)
	if err != nil {
		return nil, err
	}
	a.AppliedRules = rules

	return a, nil
}

func (s *AnnotationSQL) ListAnnotationsByDatasetID(datasetID string) ([]*model.Annotation, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id
		FROM annotations
		WHERE dataset_id = ?
		ORDER BY created_at ASC
	`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var annotations []*model.Annotation
	for rows.Next() {
		a := &model.Annotation{}
		if err := rows.Scan(
			&a.ID,
			&a.Name,
			&a.Description,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.Pages,
			&a.Segmented,
			&a.GroundTruth,
			&a.Ocred,
			&a.DatasetID,
			&a.OriginAnnotationID,
		); err != nil {
			return nil, err
		}

		rules, err := s.listAppliedRules(a.ID)
		if err != nil {
			return nil, err
		}
		a.AppliedRules = rules

		annotations = append(annotations, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

func (s *AnnotationSQL) listAppliedRules(annotationID string) ([]annotationrule.AnnotationRule, error) {
	rows, err := s.db.Query(`
		SELECT r.rule_definition
		FROM annotation_applied_rules ar
		JOIN annotation_rules r ON r.id = ar.rule_id
		WHERE ar.annotation_id = ?
		ORDER BY r.created_at ASC
	`, annotationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []annotationrule.AnnotationRule
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}

		rule, err := annotationrule.UnmarshalRuleJSON([]byte(raw))
		if err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return rules, nil
}

func (s *AnnotationSQL) UpdateAnnotation(a *model.Annotation) error {
	if a == nil {
		return fmt.Errorf("annotation is nil")
	}
	if a.ID == "" {
		return fmt.Errorf("annotation id is empty")
	}
	if a.DatasetID == "" {
		return fmt.Errorf("annotation dataset_id is empty")
	}

	now := time.Now()
	a.UpdatedAt = now
	if a.CreatedAt.IsZero() {
		a.CreatedAt = now
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 0) capture currently linked rules so we can garbage collect orphans later
	var oldRuleIDs []string
	{
		rows, err := tx.Query(`
			SELECT rule_id
			FROM annotation_applied_rules
			WHERE annotation_id = ?
		`, a.ID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var rid string
			if err := rows.Scan(&rid); err != nil {
				return err
			}
			oldRuleIDs = append(oldRuleIDs, rid)
		}
		if err := rows.Err(); err != nil {
			return err
		}
	}

	// 1) update annotations row
	res, err := tx.Exec(`
		UPDATE annotations
		SET name = ?, description = ?, updated_at = ?, pages = ?, segmented = ?, ground_truth = ?, ocred = ?, origin_annotation_id = ?
		WHERE dataset_id = ? AND id = ?
	`, a.Name, a.Description, a.UpdatedAt, a.Pages, a.Segmented, a.GroundTruth, a.Ocred, a.OriginAnnotationID, a.DatasetID, a.ID)
	if err != nil {
		return err
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if aff == 0 {
		return sql.ErrNoRows
	}

	// 2) delete all applied-rule links for this annotation
	if _, err := tx.Exec(`
		DELETE FROM annotation_applied_rules
		WHERE annotation_id = ?
	`, a.ID); err != nil {
		return err
	}

	// 3) re-add applied rules (and upsert rules table)
	for _, rule := range a.AppliedRules {
		if rule == nil {
			continue
		}

		b, err := json.Marshal(rule)
		if err != nil {
			return fmt.Errorf("marshal rule: %w", err)
		}

		sum := sha1.Sum(b)
		ruleID := "rule_" + hex.EncodeToString(sum[:])

		ruleName := "rule"
		{
			var base annotationrule.Base
			if err := json.Unmarshal(b, &base); err == nil && base.Type != "" {
				ruleName = string(base.Type)
			}
		}

		if _, err := tx.Exec(`
			INSERT OR IGNORE INTO annotation_rules (
				id, name, description, created_at, updated_at, rule_definition
			) VALUES (?, ?, ?, ?, ?, ?)
		`, ruleID, ruleName, "", now, now, string(b)); err != nil {
			return err
		}

		if _, err := tx.Exec(`
			INSERT INTO annotation_applied_rules (annotation_id, rule_id)
			VALUES (?, ?)
		`, a.ID, ruleID); err != nil {
			return err
		}
	}

	// 4) garbage-collect old rules that are now unreferenced by any annotation
	// Only checks oldRuleIDs so we do not scan the entire table.
	if len(oldRuleIDs) > 0 {
		placeholders := make([]string, 0, len(oldRuleIDs))
		args := make([]any, 0, len(oldRuleIDs))
		for _, rid := range oldRuleIDs {
			placeholders = append(placeholders, "?")
			args = append(args, rid)
		}

		q := fmt.Sprintf(`
			DELETE FROM annotation_rules
			WHERE id IN (%s)
			  AND NOT EXISTS (
			    SELECT 1
			    FROM annotation_applied_rules ar
			    WHERE ar.rule_id = annotation_rules.id
			  )
		`, strings.Join(placeholders, ","))

		if _, err := tx.Exec(q, args...); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *AnnotationSQL) InsertAnnotation(a *model.Annotation) error {
	if a == nil {
		return fmt.Errorf("annotation is nil")
	}
	if a.DatasetID == "" {
		return fmt.Errorf("annotation dataset_id is empty")
	}

	if a.ID == "" {
		a.ID = idgen.GenerateID()
	}

	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 1) annotations
	if _, err := tx.Exec(`
		INSERT INTO annotations (
			id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, dataset_id, origin_annotation_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.Name, a.Description, a.CreatedAt, a.UpdatedAt, a.Pages, a.Segmented, a.GroundTruth, a.Ocred, a.DatasetID, a.OriginAnnotationID); err != nil {
		return err
	}

	// 2) annotation_rules + annotation_applied_rules
	for _, rule := range a.AppliedRules {
		if rule == nil {
			continue
		}

		b, err := json.Marshal(rule)
		if err != nil {
			return fmt.Errorf("marshal rule: %w", err)
		}

		// Deterministic ID so identical rule definitions converge to the same DB row.
		sum := sha1.Sum(b)
		ruleID := "rule_" + hex.EncodeToString(sum[:])

		// Optional: set a readable name from the "type" field.
		ruleName := "rule"
		{
			var base annotationrule.Base
			if err := json.Unmarshal(b, &base); err == nil && base.Type != "" {
				ruleName = string(base.Type)
			}
		}

		// Insert rule row if missing
		if _, err := tx.Exec(`
			INSERT OR IGNORE INTO annotation_rules (
				id, name, description, created_at, updated_at, rule_definition
			) VALUES (?, ?, ?, ?, ?, ?)
		`, ruleID, ruleName, "", now, now, string(b)); err != nil {
			return err
		}

		// Link rule to annotation
		if _, err := tx.Exec(`
			INSERT INTO annotation_applied_rules (annotation_id, rule_id)
			VALUES (?, ?)
		`, a.ID, ruleID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *AnnotationSQL) DeleteAnnotation(datasetID string, annotationID string) error {
	_, err := s.db.Exec(`
		DELETE FROM annotations
		WHERE dataset_id = ? AND id = ?
	`, datasetID, annotationID)
	return err
}
