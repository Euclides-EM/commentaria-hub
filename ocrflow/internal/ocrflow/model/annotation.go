package model

type Annotation struct {
	Meta                     `json:",inline"`
	Pages                    string                  `json:"pages"`
	AltoDir                  string                  `json:"alto_dir" readonly:"true"`
	YoloDir                  string                  `json:"yolo_dir" readonly:"true"`
	RoboflowDir              string                  `json:"roboflow_dir" readonly:"true"`
	Dataset                  Reference               `json:"dataset" readonly:"true"`
	SegmentationModel        Reference               `json:"segmentation_model"`
	OCRModel                 Reference               `json:"ocr_model"`
	SegmentationAppliedRules []*AnnotationApplyRules `json:"segmentation_applied_rules" readonly:"true"`
}

func (a *Annotation) DatasetID() string {
	return a.Dataset.ID
}

func (a *Annotation) SegmentationModelID() string {
	return a.SegmentationModel.ID
}

func (a *Annotation) OCRModelID() string {
	return a.OCRModel.ID
}

func (a *Annotation) DeepCopy() *Annotation {
	if a == nil {
		return nil
	}
	return &Annotation{
		Meta:              a.Meta.DeepCopy(),
		Pages:             a.Pages,
		AltoDir:           a.AltoDir,
		YoloDir:           a.YoloDir,
		RoboflowDir:       a.RoboflowDir,
		Dataset:           a.Dataset.DeepCopy(),
		SegmentationModel: a.SegmentationModel.DeepCopy(),
		OCRModel:          a.OCRModel.DeepCopy(),
		SegmentationAppliedRules: func() []*AnnotationApplyRules {
			if a.SegmentationAppliedRules == nil {
				return nil
			}
			copied := make([]*AnnotationApplyRules, len(a.SegmentationAppliedRules))
			for i, v := range a.SegmentationAppliedRules {
				if v != nil {
					copied[i] = v.DeepCopy()
				}
			}
			return copied
		}(),
	}
}
