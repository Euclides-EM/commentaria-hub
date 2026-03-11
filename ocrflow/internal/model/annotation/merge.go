package annotation

type MergeRequest struct {
	AnnotationsToMerge []*Reference `json:"annotations_to_merge"`
}
