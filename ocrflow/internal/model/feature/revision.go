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
	Type              Type              `json:"type"`
	// Features is relevant only if the parent feature is root, in which case it lists the features that are part of this revision.
	Features []common.Reference `json:"features,omitempty" readonly:"true"`
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
