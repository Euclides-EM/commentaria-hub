package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/httpapi"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
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
	ruleApplier := service.NewAnnotationRuleApplier(env.DataDir, env.RoboflowAPIKey, modelSvc)
	editionSvc := service.NewEditionService()
	datasetSvc := service.NewDatasetService(ghDownloader, editionSvc, env.DataDir)
	annotationSvc := service.NewAnnotationsService(
		datasetSvc,
		ruleApplier,
	)
	annotationUploader := service.NewAnnotationsUploader(
		annotationSvc,
		datasetSvc,
		env.RoboflowAPIKey,
		env.PythonExecutable,
		env.EscriptoriumUsername,
		env.EscriptoriumPassword,
		env.EscriptoriumBasePath,
	)
	anntoationTEI := service.NewAnnotationTEI(annotationSvc, datasetSvc)

	metaStoreManager := service.NewMetaStoreManager(
		datasetSvc,
		annotationSvc,
		modelSvc,
		env.ModelsDir,
		env.DataDir,
	)
	trainSvc := service.NewTrainService(annotationSvc, modelSvc, env.TrainingDir)

	deps := &httpapi.Dependencies{
		Env:                 env,
		HealthSvc:           heathSvc,
		EditionSvc:          editionSvc,
		DatasetSvc:          datasetSvc,
		AnnotationSvc:       annotationSvc,
		ModelSvc:            modelSvc,
		TrainSvc:            trainSvc,
		MetaStoreManager:    metaStoreManager,
		AnnotationsUploader: annotationUploader,
		AnnotationTEI:       anntoationTEI,
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
