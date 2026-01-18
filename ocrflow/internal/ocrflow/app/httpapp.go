package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/httpapi"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
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
	editionStore := store.NewEditionSQL(sqlDB)
	facsimileStore := store.NewFacsimileSql(sqlDB)
	datasetStore := store.NewDatasetSQL(sqlDB)
	annotationStore := store.NewAnnotationSQL(sqlDB)
	modelStore := store.NewModelSQL(sqlDB)

	fileSystemManager := filesys.NewFileSystemManager(env.DataDir, env.TrainingDir, env.ModelsDir)

	ghDownloader := ghwrapper.NewDownloader(env.GithubToken, env.GithubDownloaderTimeout)

	healthSvc := service.NewHealthService(sqlDB)
	modelSvc := service.NewModelService(modelStore, fileSystemManager)
	ruleApplier := service.NewAnnotationRuleApplier(modelSvc, fileSystemManager, env.RoboflowAPIKey)
	editionSvc := service.NewEditionService(editionStore, facsimileStore)
	datasetSvc := service.NewDatasetService(editionSvc, datasetStore, fileSystemManager, ghDownloader)
	annotationSvc := service.NewAnnotationsService(datasetSvc, ruleApplier, fileSystemManager, annotationStore)
	annotationUploader := service.NewAnnotationsUploader(
		annotationSvc,
		datasetSvc,
		fileSystemManager,
		env.RoboflowAPIKey,
		env.PythonExecutable,
		env.EscriptoriumUsername,
		env.EscriptoriumPassword,
		env.EscriptoriumBasePath,
	)
	annotationTEI := service.NewAnnotationTEI(annotationSvc, fileSystemManager)

	metaStoreManager := service.NewMetaStoreManager(
		datasetSvc,
		annotationSvc,
		modelSvc,
		fileSystemManager,
	)
	trainSvc := service.NewTrainService(annotationSvc, modelSvc, fileSystemManager, env.TrainingDir)
	assetGen := service.NewAssetGen(datasetSvc, annotationTEI, annotationSvc, fileSystemManager)
	annotationSearch := service.NewAnnotationSearch(annotationSvc, fileSystemManager)
	deps := &httpapi.Dependencies{
		Env:                 env,
		HealthSvc:           healthSvc,
		EditionSvc:          editionSvc,
		DatasetSvc:          datasetSvc,
		AnnotationSvc:       annotationSvc,
		ModelSvc:            modelSvc,
		TrainSvc:            trainSvc,
		MetaStoreManager:    metaStoreManager,
		AnnotationsUploader: annotationUploader,
		AnnotationTEI:       annotationTEI,
		AssetGen:            assetGen,
		AnnotationSearch:    annotationSearch,
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
