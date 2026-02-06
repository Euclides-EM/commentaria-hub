package annotationrule

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Segment struct {
	Base      `json:",inline"`
	Model     string              `json:"model" example:"1615FineTunedCapricciosaM_0312"`
	ModelType common.OCRModelType `json:"model_type" example:"string,readonly"`
}

func (t *Segment) GetType() Type {
	return TypeSegment
}

func (t *Segment) SetDefaultValues() {
	t.Model = "1615FineTunedCapricciosaM_0312"
}

func NewSegment(model string) *Segment {
	return &Segment{
		Base:      Base{Type: TypeSegment, ApplicableStages: GetApplicableStages(TypeSegment)},
		Model:     model,
		ModelType: common.OCRModelTypeSegment,
	}
}
