package annotationrule

type MetadataDetails struct {
	Type    Type           `json:"type"`
	Default AnnotationRule `json:"default"`
}
