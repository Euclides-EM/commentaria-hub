package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type FeatureRevision struct {
	common.Meta
	CollectionID      string                   `json:"collection_id"`                      // collection scope
	Prompt            string                   `json:"prompt"`                             // only if execution_strategy=prompt
	Regex             string                   `json:"regex"`                              // only if execution_strategy=regex
	ExecutionStrategy FeatureExecutionStrategy `json:"execution_strategy"`                 // "prompt" or "regex"
	Note              string                   `json:"note,omitempty"`                     // optional note about the revision
	Type              FeatureType              `json:"type"`                               // "annotation" or "ner"
	Features          []Reference              `json:"features,omitempty" readonly:"true"` // list of feature IDs that are part of this revision; only if the parent feature is root
}

type FeatureType string

const (
	FeatureTypeAnnotation FeatureType = "annotation"
	FeatureTypeNER        FeatureType = "ner"
)

type FeatureExecutionStrategy string

const (
	FeatureExecutionStrategyPrompt FeatureExecutionStrategy = "prompt"
	FeatureExecutionStrategyRegex  FeatureExecutionStrategy = "regex"
)
