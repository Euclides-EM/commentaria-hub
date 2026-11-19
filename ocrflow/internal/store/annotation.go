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

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
)

const AnnotationIDPrefix = "ann"

type AnnotationSQL struct {
	BaseSQL
}

func NewAnnotationSQL(db *sql.DB) *AnnotationSQL {
	return &AnnotationSQL{
		BaseSQL: BaseSQL{db: db},
	}
}

// rowScanner is satisfied by *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanAnnotation(scanner rowScanner, a *annotation.Annotation) error {
	return scanner.Scan(
		&a.ID,
		&a.Name,
		&a.Description,
		&a.CreatedAt,
		&a.UpdatedAt,
		&a.Pages,
		&a.Segmented,
		&a.GroundTruth,
		&a.Ocred,
		&a.LinesDetected,
		&a.Hidden,
		&a.DatasetID,
		&a.OriginAnnotationID,
	)
}

func (s *AnnotationSQL) listMergedAnnotations(annotationID string) ([]annotation.MergedReference, error) {
	rows, err := s.db.Query(`
		SELECT merged_dataset_id, merged_annotation_id, merged_at
		FROM annotation_merged_annotations
		WHERE annotation_id = ?
		ORDER BY merged_at ASC
	`, annotationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var refs []annotation.MergedReference
	for rows.Next() {
		var m annotation.MergedReference
		if err := rows.Scan(&m.DatasetID, &m.ID, &m.MergedAt); err != nil {
			return nil, err
		}
		refs = append(refs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return refs, nil
}

func (s *AnnotationSQL) enrichAnnotation(a *annotation.Annotation) error {
	rules, err := s.listAppliedRules(a.ID)
	if err != nil {
		return err
	}
	a.AppliedRules = rules
	merged, err := s.listMergedAnnotations(a.ID)
	if err != nil {
		return err
	}
	a.MergedAnnotations = merged
	a.PipelineStage = calculatePipelineStage(a)
	return nil
}

func (s *AnnotationSQL) GetAnnotation(datasetID, id string) (*annotation.Annotation, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, lines_detected, hidden, dataset_id, origin_annotation_id
		FROM annotations
		WHERE dataset_id = ? AND id = ?
		LIMIT 1
	`, datasetID, id)

	a := &annotation.Annotation{}
	if err := scanAnnotation(row, a); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if err := s.enrichAnnotation(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *AnnotationSQL) ListAnnotationsByAnnotationReferences(annRefs []*annotation.Reference) ([]*annotation.Annotation, error) {
	res := make([]*annotation.Annotation, 0)
	if len(annRefs) == 0 {
		return res, nil
	}

	placeholders := make([]string, 0, len(annRefs))
	args := make([]any, 0, len(annRefs)*2)
	for _, ref := range annRefs {
		placeholders = append(placeholders, "(?, ?)")
		args = append(args, ref.DatasetID, ref.ID)
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, lines_detected, hidden, dataset_id, origin_annotation_id
		FROM annotations
		WHERE (dataset_id, id) IN (%s)
	`, strings.Join(placeholders, ","))
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		a := &annotation.Annotation{}
		if err := scanAnnotation(rows, a); err != nil {
			return nil, err
		}
		if err := s.enrichAnnotation(a); err != nil {
			return nil, err
		}
		res = append(res, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *AnnotationSQL) ListAnnotationsByDatasetID(datasetID string) ([]*annotation.Annotation, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, lines_detected, hidden, dataset_id, origin_annotation_id
		FROM annotations
		WHERE dataset_id = ?
		ORDER BY created_at ASC
	`, datasetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var annotations []*annotation.Annotation
	for rows.Next() {
		a := &annotation.Annotation{}
		if err := scanAnnotation(rows, a); err != nil {
			return nil, err
		}
		if err := s.enrichAnnotation(a); err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return annotations, nil
}

func (s *AnnotationSQL) ListAppliedRulesByAnnotationIDs() (map[string]map[string][]annotationrule.AnnotationRule, error) {
	rows, err := s.db.Query(`
		SELECT a.dataset_id, ar.annotation_id, r.rule_definition
		FROM annotation_applied_rules ar
		JOIN annotation_rules r ON r.id = ar.rule_id
		JOIN annotations a ON a.id = ar.annotation_id
		ORDER BY ar.annotation_id ASC, ar.applied_index ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]map[string][]annotationrule.AnnotationRule)
	for rows.Next() {
		var datasetID string
		var annotationID string
		var raw string
		if err := rows.Scan(&datasetID, &annotationID, &raw); err != nil {
			return nil, err
		}

		rule, err := annotationrule.UnmarshalRuleJSON([]byte(raw))
		if err != nil {
			return nil, err
		}
		if _, ok := result[datasetID]; !ok {
			result[datasetID] = make(map[string][]annotationrule.AnnotationRule)
		}
		result[datasetID][annotationID] = append(result[datasetID][annotationID], rule)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
func (s *AnnotationSQL) listAppliedRules(annotationID string) ([]annotationrule.AnnotationRule, error) {
	rows, err := s.db.Query(`
		SELECT r.rule_definition
		FROM annotation_applied_rules ar
		JOIN annotation_rules r ON r.id = ar.rule_id
		WHERE ar.annotation_id = ?
		ORDER BY ar.applied_index ASC
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

func (s *AnnotationSQL) UpdateAnnotation(a *annotation.Annotation) error {
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
			SELECT DISTINCT rule_id
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
		SET name = ?, description = ?, updated_at = ?, pages = ?, segmented = ?, ground_truth = ?, ocred = ?, lines_detected = ?, hidden = ?,  origin_annotation_id = ?
		WHERE dataset_id = ? AND id = ?
	`, a.Name, a.Description, a.UpdatedAt, a.Pages, a.Segmented, a.GroundTruth, a.Ocred, a.LinesDetected, a.Hidden, a.OriginAnnotationID, a.DatasetID, a.ID)
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

	// 3) re-add applied rules (and upsert rules table), preserving order
	for i, rule := range a.AppliedRules {
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
			INSERT INTO annotation_applied_rules (annotation_id, rule_id, applied_index)
			VALUES (?, ?, ?)
		`, a.ID, ruleID, i); err != nil {
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

	// 5) replace merged-annotation links
	if _, err := tx.Exec(`
		DELETE FROM annotation_merged_annotations
		WHERE annotation_id = ?
	`, a.ID); err != nil {
		return err
	}
	for _, m := range a.MergedAnnotations {
		if _, err := tx.Exec(`
			INSERT INTO annotation_merged_annotations (annotation_id, merged_dataset_id, merged_annotation_id, merged_at)
			VALUES (?, ?, ?, ?)
		`, a.ID, m.DatasetID, m.ID, m.MergedAt); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (s *AnnotationSQL) InsertAnnotation(a *annotation.Annotation) error {
	if a == nil {
		return fmt.Errorf("annotation is nil")
	}
	if a.DatasetID == "" {
		return fmt.Errorf("annotation dataset_id is empty")
	}

	if a.ID == "" {
		a.ID = idgen.GenerateID(AnnotationIDPrefix)
	}

	now := time.Now()
	a.CreatedAt = now
	a.UpdatedAt = now

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// 1) annotations
	if _, err := tx.Exec(`
		INSERT INTO annotations (
			id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, lines_detected, hidden, dataset_id, origin_annotation_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.Name, a.Description, a.CreatedAt, a.UpdatedAt, a.Pages, a.Segmented, a.GroundTruth, a.Ocred, a.LinesDetected, a.Hidden, a.DatasetID, a.OriginAnnotationID); err != nil {
		return err
	}

	// 2) annotation_rules + annotation_applied_rules (preserving order)
	for i, rule := range a.AppliedRules {
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
			INSERT INTO annotation_applied_rules (annotation_id, rule_id, applied_index)
			VALUES (?, ?, ?)
		`, a.ID, ruleID, i); err != nil {
			return err
		}
	}

	// 3) merged annotations
	for _, m := range a.MergedAnnotations {
		if _, err := tx.Exec(`
			INSERT INTO annotation_merged_annotations (annotation_id, merged_dataset_id, merged_annotation_id, merged_at)
			VALUES (?, ?, ?, ?)
		`, a.ID, m.DatasetID, m.ID, m.MergedAt); err != nil {
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

func (s *AnnotationSQL) ListAnnotationsByDatasetIDs(ds []string) ([]*annotation.Annotation, error) {
	if len(ds) == 0 {
		return []*annotation.Annotation{}, nil
	}

	placeholders := make([]string, 0, len(ds))
	args := make([]any, 0, len(ds))
	for _, datasetID := range ds {
		placeholders = append(placeholders, "?")
		args = append(args, datasetID)
	}

	query := fmt.Sprintf(`
		SELECT id, name, description, created_at, updated_at, pages, segmented, ground_truth, ocred, lines_detected, hidden, dataset_id, origin_annotation_id
		FROM annotations
		WHERE dataset_id IN (%s)
		ORDER BY created_at ASC
	`, strings.Join(placeholders, ","))
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var annotations []*annotation.Annotation
	for rows.Next() {
		a := &annotation.Annotation{}
		if err := scanAnnotation(rows, a); err != nil {
			return nil, err
		}
		annotations = append(annotations, a)
	}
	return annotations, nil
}

func calculatePipelineStage(a *annotation.Annotation) annotationrule.PipelineStage {
	s := annotationrule.PipelineStageRaw
	if a.Ocred {
		s = annotationrule.PipelineStageOCR
	} else if a.LinesDetected {
		s = annotationrule.PipelineStageTextLineSegmentation
	} else if a.Segmented {
		s = annotationrule.PipelineStageZoneSegmentation
	}
	for _, rule := range a.AppliedRules {
		if rule == nil {
			continue
		}
		minEnsured := rule.EnsuredPipelineStage()
		if minEnsured.After(s) {
			s = minEnsured
		}
	}
	return s
}
