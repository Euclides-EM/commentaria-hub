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
	Color          string             `json:"color"`                                     // optional UI color (hex)
	LatestRevision *FeatureRevision   `json:"latest_revision,omitempty" readonly:"true"` // ONLY if expand=latest_revision
	Revisions      []*FeatureRevision `json:"revisions,omitempty" readonly:"true"`       // ONLY if expand=revisions
}

type FeatureExpandOptions string

const (
	FeatureExpandLatestRevision FeatureExpandOptions = "latest_revision"
	FeatureExpandRevisions      FeatureExpandOptions = "revisions"
)

func ToFeatureExpandOptions(s string) []FeatureExpandOptions {
	var opts []FeatureExpandOptions
	for _, candidate := range strings.Split(s, ",") {
		switch FeatureExpandOptions(candidate) {
		case FeatureExpandLatestRevision:
			opts = append(opts, FeatureExpandLatestRevision)
		case FeatureExpandRevisions:
			opts = append(opts, FeatureExpandRevisions)
		}
	}
	return opts
}
