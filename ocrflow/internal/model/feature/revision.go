package feature

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type Revision struct {
	common.Meta
	Scope       DefScope   `json:"scope"`
	FeatureID   string     `json:"feature_id"`
	Prompt      string     `json:"prompt"`
	Categorizer string     `json:"categorizer"`
	AIProvider  AIProvider `json:"ai_provider"`
	AIModel     string     `json:"ai_model"`
}
