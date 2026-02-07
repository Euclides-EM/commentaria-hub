package featureplat

import (
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Feature struct {
	common.Meta
	CollectionID   string             `json:"collection_id"`                             // collection scope
	IsRoot         bool               `json:"is_root"`                                   // immutable
	IsDefault      bool               `json:"is_default"`                                // whether this feature should be used by default
	LatestRevision *FeatureRevision   `json:"latest_revision,omitempty" readonly:"true"` // ONLY if expand=latest_revision
	Revisions      []*FeatureRevision `json:"revisions,omitempty" readonly:"true"`       // ONLY if expand=revisions
}

type FeatureExpandOptions string

const (
	FeatureExpandLatestRevision FeatureExpandOptions = "latest_revision"
	FeatureExpandRevisions      FeatureExpandOptions = "revisions"
)

func ToFeatureExpandOptions(values []string) []FeatureExpandOptions {
	var opts []FeatureExpandOptions
	for _, v := range values {
		for _, candidate := range strings.Split(v, ",") {
			switch FeatureExpandOptions(strings.TrimSpace(candidate)) {
			case FeatureExpandLatestRevision:
				opts = append(opts, FeatureExpandLatestRevision)
			case FeatureExpandRevisions:
				opts = append(opts, FeatureExpandRevisions)
			}
		}
	}
	return opts
}
