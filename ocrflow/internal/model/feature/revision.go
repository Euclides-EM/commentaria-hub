package feature

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Revision struct {
	common.Meta
	DatasetID   string `json:"dataset_id"`
	FeatureID   string `json:"feature_id"`
	Prompt      string `json:"prompt"`
	Categorizer string `json:"categorizer"`
}
