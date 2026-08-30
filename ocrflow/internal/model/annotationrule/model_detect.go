package annotationrule

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type ModelDetect struct {
	Base       `json:",inline"`
	Model      string              `json:"model" example:"1615FineTunedCapricciosaM_0312"`
	ModelType  common.OCRModelType `json:"model_type" example:"string,readonly"`
	UseGPUFarm bool                `json:"use_gpu_farm"`
}

func (t *ModelDetect) GetType() Type {
	return TypeModelDetect
}

func (t *ModelDetect) SetDefaultValues() {
	t.Model = "1615FineTunedCapricciosaM_0312"
}

func NewModelDetect(model string) *ModelDetect {
	return newModelDetect(model, common.OCRModelTypeSegment)
}

func NewOCRModelDetect(model string) *ModelDetect {
	return newModelDetect(model, common.OCRModelTypeOCR)
}

func newModelDetect(model string, modelType common.OCRModelType) *ModelDetect {
	return &ModelDetect{
		Base:       Base{Type: TypeModelDetect, ApplicableStages: modelDetectApplicableStages(modelType)},
		Model:      model,
		ModelType:  modelType,
		UseGPUFarm: true,
	}
}

func (t *ModelDetect) ApplicablePipelineStages() []PipelineStage {
	return modelDetectApplicableStages(t.ModelType)
}

func (t *ModelDetect) EnsuredPipelineStage() PipelineStage {
	return modelDetectEnsuredStage(t.ModelType)
}

func modelDetectApplicableStages(modelType common.OCRModelType) []PipelineStage {
	if modelType == common.OCRModelTypeOCR {
		return []PipelineStage{PipelineStageTextLineSegmentation}
	}
	return GetApplicableStages(TypeModelDetect)
}

func modelDetectEnsuredStage(modelType common.OCRModelType) PipelineStage {
	if modelType == common.OCRModelTypeOCR {
		return PipelineStageOCR
	}
	return minEnsuredStageByType[TypeModelDetect]
}
