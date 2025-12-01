package model

type TrainingStatus string

const (
	TrainingStatusRunning   TrainingStatus = "running"
	TrainingStatusCompleted TrainingStatus = "completed"
	TrainingStatusFailed    TrainingStatus = "failed"
)

type Training struct {
	Meta           `json:",inline"`
	OriginModel    Reference             `json:"origin_model"`
	AnnotationSets []AnnotationReference `json:"annotation_sets"`
	Status         TrainingStatus        `json:"status"`
	Name           string                `json:"name"`
}

func (t *Training) DeepCopy() *Training {
	if t == nil {
		return nil
	}
	return &Training{
		Meta:        t.Meta.DeepCopy(),
		OriginModel: t.OriginModel.DeepCopy(),
		AnnotationSets: func(src []AnnotationReference) []AnnotationReference {
			if src == nil {
				return nil
			}
			dst := make([]AnnotationReference, len(src))
			for i, v := range src {
				dst[i] = v.DeepCopy()
			}
			return dst
		}(t.AnnotationSets),
	}
}
