package app

import (
	"database/sql"
	"fmt"
	"log"
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

	fileSystemManager := filesys.NewFileSystemManager(env.DataDir, env.TrainingDir, env.ModelsDir)
	editionStore := store.NewEditionCSV(env.ItemsMetadataStoreDir)

	sqlDB, err := db.InitDB(env.DBPath, env.MigrationsDir)
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}
	facsimileStore := store.NewFacsimileSql(sqlDB)
	datasetStore := store.NewDatasetSQL(sqlDB, fileSystemManager)
	annotationStore := store.NewAnnotationSQL(sqlDB)
	modelStore := store.NewModelSQL(sqlDB)
	featureRevisionStore := store.NewFeatureRevisionSQL(sqlDB)
	featureExecutionStore := store.NewFeatureExecutionSQL(sqlDB)
	featureResultStore := store.NewFeatureResultSQL(sqlDB)
	tpsTranscriptionsStore := store.NewTPSTranscriptions()

	ghDownloader := ghwrapper.NewWrapper(env.GithubToken, env.GithubDownloaderTimeout)

	healthSvc := service.NewHealthService(sqlDB)
	modelSvc := service.NewModelService(modelStore, fileSystemManager)
	ruleApplier := service.NewAnnotationRuleApplier(modelSvc, fileSystemManager, env.RoboflowAPIKey)
	editionSvc := service.NewEditionService(editionStore, facsimileStore)
	facsimileSvc := service.NewFacsimileService(facsimileStore, ghDownloader, env.FacsimilesGithubRepoUrl)
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
	featureResultSvc := service.NewResult(featureResultStore)
	titlePageTEI := service.NewTitlePageTEI(featureResultSvc, tpsTranscriptionsStore)
	annotationTEI := service.NewAnnotationTEI(annotationSvc, fileSystemManager, titlePageTEI)

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

	log.Printf("warming edition cache...")
	if err := editionStore.WarmCache(); err != nil {
		log.Fatalf("edition cache warm failed: %v", err)
	}
	log.Printf("finished warming edition cache")

	log.Printf("updating facsimiles from github...")
	if err := facsimileSvc.UpdateFromGithubRepo(); err != nil {
		log.Printf("warning: failed to update facsimiles from github: %v", err)
	}
	log.Printf("finished updating facsimiles from github")

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
		USTC:                service.NewUSTC(),
		VCSMgt:              service.NewVCSMgt(env.ItemsMetadataStoreDir, fileSystemManager.DatasetImagesDirByID("tps")),
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
