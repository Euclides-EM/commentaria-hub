package annotationrule

type PipelineStage string

const (
	PipelineStageRaw                  PipelineStage = "raw"
	PipelineStageZoneSegmentation     PipelineStage = "zone_segmentation"
	PipelineStageTextLineSegmentation PipelineStage = "text_line_segmentation"
	PipelineStageOCR                  PipelineStage = "ocr"
)

func (p PipelineStage) After(other PipelineStage) bool {
	stages := []PipelineStage{
		PipelineStageRaw,
		PipelineStageZoneSegmentation,
		PipelineStageTextLineSegmentation,
		PipelineStageOCR,
	}

	var pIndex, otherIndex int
	for i, stage := range stages {
		if stage == p {
			pIndex = i
		}
		if stage == other {
			otherIndex = i
		}
	}
	return pIndex > otherIndex
}

var AllPipelineStages = []PipelineStage{
	PipelineStageRaw,
	PipelineStageZoneSegmentation,
	PipelineStageTextLineSegmentation,
	PipelineStageOCR,
}
