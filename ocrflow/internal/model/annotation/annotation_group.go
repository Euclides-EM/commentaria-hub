package annotation

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Group struct {
	common.Meta `json:",inline"`
	Annotations []*Reference `json:"annotations"`
}
