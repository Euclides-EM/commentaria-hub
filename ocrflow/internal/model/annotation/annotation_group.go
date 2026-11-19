package annotation

import "github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"

type Group struct {
	common.Meta `json:",inline"`
	Annotations []*Reference `json:"annotations"`
}
