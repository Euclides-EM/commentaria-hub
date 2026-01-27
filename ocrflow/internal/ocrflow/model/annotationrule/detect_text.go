package annotationrule

import "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model/common"

type DetectText struct {
	Base      `json:",inline"`
	Model     string              `json:"model" example:"1615FineTunedCapricciosaM_0312"`
	ModelType common.OCRModelType `json:"model_type" example:"text"`
}

func (t *DetectText) GetType() Type {
	return TypeDetectText
}

func (t *DetectText) SetDefaultValues() {
	t.Model = "1615FineTunedGallicorpor_0301"
}

func NewDetectText(model string) *DetectText {
	return &DetectText{
		Base:      Base{Type: TypeDetectText, ApplicableStages: GetApplicableStages(TypeDetectText)},
		Model:     model,
		ModelType: common.OCRModelTypeOCR,
	}
}
