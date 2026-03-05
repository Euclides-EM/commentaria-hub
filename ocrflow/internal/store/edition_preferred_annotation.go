package store

import (
	"database/sql"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
)

type EditionPreferredAnnotationSql struct {
	db *sql.DB
}

func NewEditionPreferredAnnotationSql(db *sql.DB) *EditionPreferredAnnotationSql {
	return &EditionPreferredAnnotationSql{db: db}
}

func (s *EditionPreferredAnnotationSql) ListEditionPreferredTranscription(editions []string) (map[string]*annotation.Reference, error) {

}

func (s *EditionPreferredAnnotationSql) UpsertEditionPreferredTranscription(editionID string, preferredAnnotation *annotation.Reference) error {

}
