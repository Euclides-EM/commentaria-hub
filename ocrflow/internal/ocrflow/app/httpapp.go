package app

import (
	"database/sql"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/httpapi"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"net/http"
)

type App struct {
	Env    *config.EnvConfig
	DB     *sql.DB
	Router http.Handler
}

func NewHTTPApp() (*App, error) {
	env, err := config.InitEnv()
	if err != nil {
		return nil, fmt.Errorf("init env: %w", err)
	}

	sqlDB, err := db.InitDB(env.DBPath, env.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}

	ghDownloader := ghwrapper.NewDownloader(env.GithubToken, env.GithubDownloaderTimeout)

	heathSvc := service.NewHealthService(sqlDB)
	modelSvc := service.NewModelService(env.ModelsDir)
	editionSvc := service.NewEditionService()
	datasetSvc := service.NewDatasetService(ghDownloader, editionSvc, env.DataDir)
	annotationsSvc := service.NewAnnotationsService(env.PythonExecutable, datasetSvc, modelSvc, env.RoboflowAPIKey)

	deps := &httpapi.Dependencies{
		Env:            env,
		HealthSvc:      heathSvc,
		EditionSvc:     editionSvc,
		DatasetSvc:     datasetSvc,
		AnnotationsSvc: annotationsSvc,
		ModelSvc:       modelSvc,
	}

	router := httpapi.NewRouter(deps)

	return &App{
		Env:    env,
		DB:     sqlDB,
		Router: router,
	}, nil
}

func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
