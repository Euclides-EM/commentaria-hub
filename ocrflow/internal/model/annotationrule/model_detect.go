package annotationrule

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type ModelDetect struct {
	Base      `json:",inline"`
	Model     string              `json:"model" example:"1615FineTunedCapricciosaM_0312"`
	ModelType common.OCRModelType `json:"model_type" example:"string,readonly"`
}

func (t *ModelDetect) GetType() Type {
	return TypeModelDetect
}

func (t *ModelDetect) SetDefaultValues() {
	t.Model = "1615FineTunedCapricciosaM_0312"
}

func NewModelDetect(model string) *ModelDetect {
	return &ModelDetect{
		Base:      Base{Type: TypeModelDetect, ApplicableStages: GetApplicableStages(TypeModelDetect)},
		Model:     model,
		ModelType: common.OCRModelTypeSegment,
	}
}
