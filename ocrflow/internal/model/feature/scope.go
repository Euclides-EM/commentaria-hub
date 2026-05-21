package feature

type ScopeType string

const (
	ScopeTypeDataset  ScopeType = "dataset"
	ScopeTypeEditions ScopeType = "editions"
)

type DefScope struct {
	Type ScopeType `json:"type"`
	// DatasetID is relevant only for ScopeTypeDataset
	DatasetID string `json:"dataset_id"`
}

func NewDatasetDefScope(datasetID string) DefScope {
	return DefScope{
		Type:      ScopeTypeDataset,
		DatasetID: datasetID,
	}
}

func NewEditionDefScope() DefScope {
	return DefScope{
		Type: ScopeTypeEditions,
	}
}

type ExecScope struct {
	DefScope     `json:",inline"`
	AnnotationID string `json:"annotation_id"`
}

func NewDatasetExecScope(datasetID, annotationID string) ExecScope {
	return ExecScope{
		DefScope: DefScope{
			Type:      ScopeTypeDataset,
			DatasetID: datasetID,
		},
		AnnotationID: annotationID,
	}
}

func NewEditionExecScope() ExecScope {
	return ExecScope{
		DefScope: DefScope{
			Type: ScopeTypeEditions,
		},
	}
}
