package featureplat

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	fpservice "github.com/MiaMish/elements-dh/ocrflow/internal/service/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service/ocrflow"
)

type Dependencies struct {
	Env                 *config.FeatureAppEnvConfig
	HealthSvc           *ocrflow.Health
	FeatureSvc          *fpservice.Feature
	FeatureRevisionSvc  *fpservice.Revision
	FeatureExecutionSvc *fpservice.Execution
	FeatureResultSvc    *fpservice.Result
	TEISvc              *fpservice.TEI
}
