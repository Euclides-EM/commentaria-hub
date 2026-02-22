package feature

import (
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Feature struct {
	common.Meta
	DatasetID string `json:"dataset_id"`
	// IsRoot is immutable.
	IsRoot    bool `json:"is_root"`
	IsDefault bool `json:"is_default"`
	// IsList indicates whether this feature can have multiple values (e.g. a list of named entities) or just a single value (e.g. an annotation). It is immutable.
	IsList bool `json:"is_list"`
	// Color is an optional UI color hint for this feature, e.g. "#FF0000" for red.
	Color string `json:"color"`
	// Type is immutable and determines the type of this feature, e.g. annotation or NER.
	Type Type `json:"type"`
	// Features is relevant only if this feature is root; it lists the child features that are part of this feature.
	Features []common.Reference `json:"features,omitempty"`
	// LatestRevision is the most recent revision of this feature. It is read-only and only included if expand=latest_revision is specified in the request.
	LatestRevision *Revision `json:"latest_revision,omitempty" readonly:"true"`
	// Revisions is the list of all revisions of this feature, ordered by created_at descending. It is read-only and only included if expand=revisions is specified in the request.
	Revisions []*Revision `json:"revisions,omitempty" readonly:"true"`
}

type ExpandOptions string

const (
	ExpandLatestRevision ExpandOptions = "latest_revision"
	ExpandRevisions      ExpandOptions = "revisions"
)

func ToFeatureExpandOptions(values []string) []ExpandOptions {
	var opts []ExpandOptions
	for _, v := range values {
		for _, candidate := range strings.Split(v, ",") {
			switch ExpandOptions(strings.TrimSpace(candidate)) {
			case ExpandLatestRevision:
				opts = append(opts, ExpandLatestRevision)
			case ExpandRevisions:
				opts = append(opts, ExpandRevisions)
			}
		}
	}
	return opts
}
