package annotation

import (
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type Annotation struct {
	common.Meta        `json:",inline"`
	Pages              string                         `json:"pages"`
	Segmented          bool                           `json:"segmented" readonly:"true"`
	GroundTruth        bool                           `json:"ground_truth"`
	Ocred              bool                           `json:"ocred" readonly:"true"`
	LinesDetected      bool                           `json:"lines_detected" readonly:"true"`
	Hidden             bool                           `json:"hidden"`
	DatasetID          string                         `json:"dataset_id" readonly:"true"`
	AppliedRules       annotationrule.AnnotationRules `json:"applied_rules" readonly:"true"`
	OriginAnnotationID string                         `json:"origin_annotation_id,omitempty" readonly:"true"`
	MergedAnnotations  []MergedReference              `json:"merged_annotations,omitempty" readonly:"true"`
	PipelineStage      annotationrule.PipelineStage   `json:"pipeline_stage,omitempty" readonly:"true"`

	TranscriptionFallback *TranscriptionFallback `json:"transcription_fallback,omitempty" readonly:"true"`
}

type TranscriptionFallback struct {
	Format  TranscriptionFormat `json:"format"`
	Level   TranscriptionLevel  `json:"level"`
	Partial bool                `json:"partial"`
}

func NewTranscriptionFallback(format TranscriptionFormat, level TranscriptionLevel, partial bool) *TranscriptionFallback {
	return &TranscriptionFallback{
		Format:  format,
		Level:   level,
		Partial: partial,
	}
}

type TranscriptionFormat string

const (
	TranscriptionFormatText     TranscriptionFormat = "text"
	TranscriptionFormatMarkdown TranscriptionFormat = "markdown"
	TranscriptionFormatALTO     TranscriptionFormat = "alto"
)

type TranscriptionLevel string

const (
	TranscriptionLevelDataset    TranscriptionLevel = "dataset"
	TranscriptionLevelAnnotation TranscriptionLevel = "annotation"
)

type MergedReference struct {
	Reference `json:",inline"`
	MergedAt  time.Time `json:"merged_at"`
}
