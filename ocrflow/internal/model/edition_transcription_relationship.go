package model

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
)

type EditionTranscription struct {
	EditionID           string                                   `json:"edition_id"`
	Datasets            []string                                 `json:"datasets"`
	PreferredAnnotation *EditionTranscriptionPreferredAnnotation `json:"preferred_annotation,omitempty"`
}

type EditionTranscriptionPreferredAnnotation struct {
	annotation.Reference `json:",inline"`
	Source               EditionTranscriptionPreferredAnnotationSource `json:"source"`
}

type EditionTranscriptionPreferredAnnotationSource string

const (
	EditionTranscriptionPreferredAnnotationSourceManual     EditionTranscriptionPreferredAnnotationSource = "manual"
	EditionTranscriptionPreferredAnnotationSourceCalculated EditionTranscriptionPreferredAnnotationSource = "calculated"
)
