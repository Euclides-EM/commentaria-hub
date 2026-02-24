package feature

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type ResultValueProfile struct {
	common.Meta `json:",inline"`
	Attributes  map[string]*ResultValueProfileAttribute `json:"attributes,omitempty"`
	Regex       string                                  `json:"regex,omitempty"`
}

type ResultValueProfileAttribute struct {
	StringValue string   `json:"string_value,omitempty"`
	ListValue   []string `json:"list_value,omitempty"`
}
