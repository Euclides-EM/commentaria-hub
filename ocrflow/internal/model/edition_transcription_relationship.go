package model

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
)

type EditionTranscription struct {
	EditionID           string                `json:"edition_id"`
	Datasets            []string              `json:"datasets"`
	PreferredAnnotation *annotation.Reference `json:"preferred_annotation,omitempty"`
}
