package app

import (
	"database/sql"
	"fmt"
	"net/http"

	api "github.com/MiaMish/elements-dh/ocrflow/internal/api/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
)

type FeaturePlatform struct {
	Env    *config.FeatureAppEnvConfig
	DB     *sql.DB
	Router http.Handler
}

func NewFeaturePlatform() (*FeaturePlatform, error) {
	env, err := config.InitFeatureAppEnvConfig()
	if err != nil {
		return nil, fmt.Errorf("init env: %w", err)
	}

	sqlDB, err := db.InitDB(env.DBPath, env.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	healthSvc := ocrflow.NewHealthService(sqlDB)

	deps := &api.Dependencies{
		Env:                 env,
		HealthSvc:           healthSvc,
		FeatureSvc:          featureplat.NewFeature(),
		FeatureRevisionSvc:  featureplat.NewRevision(),
		FeatureResultSvc:    featureplat.NewResult(),
		FeatureExecutionSvc: featureplat.NewExecution(),
		TEISvc:              featureplat.NewTEI(),
	}

	router := api.NewFeatureAppRouter(deps)

	return &FeaturePlatform{
		Env:    env,
		DB:     sqlDB,
		Router: router,
	}, nil
}

func (a *FeaturePlatform) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
