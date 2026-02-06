package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	ofservice "github.com/MiaMish/elements-dh/ocrflow/internal/service/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
)

type OCRFlowApp struct {
	Env    *config.EnvConfig
	DB     *sql.DB
	Router http.Handler
}

func NewOCRFlowApp() (*OCRFlowApp, error) {
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

	healthSvc := ofservice.NewHealthService(sqlDB)
	modelSvc := ofservice.NewModelService(modelStore, fileSystemManager)
	ruleApplier := ofservice.NewAnnotationRuleApplier(modelSvc, fileSystemManager, env.RoboflowAPIKey)
	editionSvc := ofservice.NewEditionService(editionStore, facsimileStore)
	datasetSvc := ofservice.NewDatasetService(editionSvc, datasetStore, fileSystemManager, ghDownloader)
	annotationSvc := ofservice.NewAnnotationsService(datasetSvc, ruleApplier, fileSystemManager, annotationStore)
	metadataDetailsSvc := ofservice.NewMetadataDetails()
	annotationUploader := ofservice.NewAnnotationsUploader(
		annotationSvc,
		datasetSvc,
		fileSystemManager,
		env.RoboflowAPIKey,
		env.PythonExecutable,
		env.EscriptoriumUsername,
		env.EscriptoriumPassword,
		env.EscriptoriumBasePath,
	)
	annotationTEI := ofservice.NewAnnotationTEI(annotationSvc, fileSystemManager)

	metaStoreManager := ofservice.NewMetaStoreManager(
		datasetSvc,
		annotationSvc,
		modelSvc,
		fileSystemManager,
	)
	trainSvc := ofservice.NewTrainService(annotationSvc, modelSvc, fileSystemManager, env.TrainingDir)
	assetGen := ofservice.NewAssetGen(datasetSvc, annotationTEI, annotationSvc, fileSystemManager)
	annotationSearch := ofservice.NewAnnotationSearch(annotationSvc, fileSystemManager)
	deps := &ocrflow.Dependencies{
		Env:                 env,
		HealthSvc:           healthSvc,
		EditionSvc:          editionSvc,
		DatasetSvc:          datasetSvc,
		AnnotationSvc:       annotationSvc,
		ModelSvc:            modelSvc,
		TrainSvc:            trainSvc,
		MetadataDetailsSvc:  metadataDetailsSvc,
		MetaStoreManager:    metaStoreManager,
		AnnotationsUploader: annotationUploader,
		AnnotationTEI:       annotationTEI,
		AssetGen:            assetGen,
		AnnotationSearch:    annotationSearch,
	}

	router := ocrflow.NewRouter(deps)

	return &OCRFlowApp{
		Env:    env,
		DB:     sqlDB,
		Router: router,
	}, nil
}

func (a *OCRFlowApp) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
