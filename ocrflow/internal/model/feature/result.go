package feature

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Result struct {
	common.Meta  `json:",inline"`
	DatasetID    string `json:"dataset_id"`
	AnnotationID string `json:"annotation_id"`
	Feature      string `json:"feature"`
	Key          string `json:"key"`
	// Source indicates the origin of the value, such as which OCR process or manual correction it came from. This is important for traceability and debugging.
	Source ResultSource `json:"source"`
	// Values is a list of all the values that were extracted for this feature. There may be multiple values if the feature appears multiple times in the document.
	Values []ResultValue `json:"values"`
	// Note allows for any additional comments or annotations about the value, such as uncertainties, corrections, or special cases. This can be useful for human reviewers or downstream processing.
	Note string `json:"note,omitempty"`
}

type ResultSource struct {
	Resp     string `json:"resp"`
	Id       string `json:"id,omitempty"`
	Revision string `json:"revision,omitempty"`
	Name     string `json:"name,omitempty"`
}

type ResultValue struct {
	// Surface is the exact text as it appears in the document, which may include line breaks, hyphenation, and other formatting. This is the "raw" value that was extracted from the document.
	Surface    string            `json:"surface"`
	Properties map[string]string `json:"properties,omitempty"`
}
