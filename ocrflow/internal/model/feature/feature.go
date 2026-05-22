package feature

import (
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
)

type Feature struct {
	common.Meta
	Scope      DefScope `json:"scope"`
	IsDefault  bool     `json:"is_default"`
	IsList      bool   `json:"is_list"`
	IsBoolean   bool   `json:"is_boolean"`
	FeatureName string `json:"feature_name,omitempty"`
	Color       string `json:"color"`
	Properties []string `json:"properties,omitempty"`

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

func ToExpandOptions(values []string) []ExpandOptions {
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
