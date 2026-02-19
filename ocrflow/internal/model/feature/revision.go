package feature

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Revision struct {
	common.Meta
	DatasetID string `json:"dataset_id"`
	FeatureID string `json:"feature_id"`
	// Prompt is relevant only if the execution strategy is ExecutionStrategy.
	Prompt string `json:"prompt"`
	// Regex is relevant only if the execution strategy is ExecutionStrategyRegex.
	Regex             string            `json:"regex"`
	ExecutionStrategy ExecutionStrategy `json:"execution_strategy"`
	Note              string            `json:"note,omitempty"`
}

type Type string

const (
	TypeAnnotation Type = "annotation"
	TypeNER        Type = "ner"
)

type ExecutionStrategy string

const (
	ExecutionStrategyPrompt ExecutionStrategy = "prompt"
	ExecutionStrategyRegex  ExecutionStrategy = "regex"
)
