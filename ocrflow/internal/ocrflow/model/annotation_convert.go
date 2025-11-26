package model

type SegmontoGranularity string

const (
	SegmontoGranularityRegion  SegmontoGranularity = "region"
	SegmontoGranularitySubtype SegmontoGranularity = "subtype"
	SegmontoGranularityFull    SegmontoGranularity = "full"
)

type AnnotationConvert struct {
	From                AnnotationFormat    `json:"from" example:"alto"`
	To                  AnnotationFormat    `json:"to" example:"yolo"`
	Shuffle             float64             `json:"shuffle" example:"0"`
	SegmontoGranularity SegmontoGranularity `json:"segmonto_granularity" example:"full"`
}
