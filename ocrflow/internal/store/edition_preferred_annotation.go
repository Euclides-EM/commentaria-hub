package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

type EditionPreferredAnnotationSql struct {
	db *sql.DB
}

func NewEditionPreferredAnnotationSql(db *sql.DB) *EditionPreferredAnnotationSql {
	return &EditionPreferredAnnotationSql{db: db}
}

func (s *EditionPreferredAnnotationSql) ListEditionPreferredTranscription(editions []string) (map[string]*annotation.Reference, error) {
	out := make(map[string]*annotation.Reference)
	if len(editions) == 0 {
		return out, nil
	}

	placeholders := make([]string, 0, len(editions))
	args := make([]any, 0, len(editions))
	for _, e := range editions {
		placeholders = append(placeholders, "?")
		args = append(args, e)
	}
	query := fmt.Sprintf(`
		SELECT edition_id, dataset_id, annotation_id
		FROM editions_preferred_annotation
		WHERE edition_id IN (%s)
		ORDER BY edition_id, dataset_id
	`, strings.Join(placeholders, ","))
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var editionID, datasetID, annotationID string
	for rows.Next() {
		if err := rows.Scan(&editionID, &datasetID, &annotationID); err != nil {
			return nil, err
		}
		if _, ok := out[editionID]; !ok {
			out[editionID] = &annotation.Reference{DatasetID: datasetID, ID: annotationID}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *EditionPreferredAnnotationSql) UpsertEditionPreferredTranscription(editionID string, preferredAnnotation *annotation.Reference) error {
	if editionID == "" {
		return fmt.Errorf("edition_id is required")
	}
	if preferredAnnotation == nil || preferredAnnotation.DatasetID == "" || preferredAnnotation.ID == "" {
		return fmt.Errorf("preferred annotation with dataset_id and id is required")
	}

	_, err := s.db.Exec(`
		DELETE FROM editions_preferred_annotation WHERE edition_id = ?
	`, editionID)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO editions_preferred_annotation (edition_id, dataset_id, annotation_id)
		VALUES (?, ?, ?)
	`, editionID, preferredAnnotation.DatasetID, preferredAnnotation.ID)
	return err
}

func (s *EditionPreferredAnnotationSql) OnDeleteEdition(editionID string) {
	_, err := s.db.Exec(`
		DELETE FROM editions_preferred_annotation WHERE edition_id = ?
	`, editionID)
	if err != nil {
		fmt.Printf("failed to delete preferred annotation for edition %s: %v", editionID, err)
	}
}
