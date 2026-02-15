package app

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api"
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service"
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
	editionStore := store.NewEditionCSV(env.DocsPublicDir)
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
	facsimileSvc := service.NewFacsimileService(facsimileStore)
	datasetSvc := service.NewDatasetService(editionSvc, facsimileSvc, datasetStore, fileSystemManager, ghDownloader)
	annotationSvc := service.NewAnnotationsService(datasetSvc, ruleApplier, fileSystemManager, annotationStore)
	metadataDetailsSvc := service.NewMetadataDetails()
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
	featureStore := store.NewFeatureSQL(sqlDB)
	featureRevisionStore := store.NewFeatureRevisionSQL(sqlDB)
	featureExecutionStore := store.NewFeatureExecutionSQL(sqlDB)
	featureResultStore := store.NewFeatureResultSQL(sqlDB)
	tpsTranscriptionsStore := store.NewTPSTranscriptions()

	featureResultSvc := service.NewResult(featureResultStore)
	deps := &api.Dependencies{
		Env:                 env,
		HealthSvc:           healthSvc,
		EditionSvc:          editionSvc,
		FacsimileSvc:        facsimileSvc,
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
		FeatureSvc:          service.NewFeature(featureStore, featureRevisionStore),
		FeatureRevisionSvc:  service.NewRevision(featureRevisionStore),
		FeatureResultSvc:    featureResultSvc,
		FeatureExecutionSvc: service.NewExecution(featureExecutionStore),
		TEISvc:              service.NewTEI(featureResultSvc, tpsTranscriptionsStore),
		USTC:                service.NewUSTC(),
		VCSMgt:              service.NewVCSMgt(env.DocsPublicDir),
	}

	router := api.NewRouter(deps)

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
