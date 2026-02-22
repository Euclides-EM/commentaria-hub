package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api"
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/diagramcrops"
	"github.com/MiaMish/elements-dh/ocrflow/internal/migrations"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/ghwrapper"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
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

	fileSystemManager := filesys.NewFileSystemManager(env.DataDir(), env.TrainingDir(), env.ModelsDir(), env.DiagramsDir())
	editionStore := store.NewEditionCSV(env.ItemsMetadataStoreDir())
	geoStore := store.NewGeoCSV(env.ItemsMetadataStoreDir())
	sqlDB, err := db.InitDB(env.DBPath(), migrations.Migrations, "ocrflow")
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}
	facsimileStore := store.NewFacsimileSql(sqlDB)
	datasetStore := store.NewDatasetSQL(sqlDB, fileSystemManager)
	annotationStore := store.NewAnnotationSQL(sqlDB)
	modelStore := store.NewModelSQL(sqlDB)
	featureRevisionStore := store.NewFeatureRevisionSQL(sqlDB)
	featureExecutionStore := store.NewFeatureExecutionSQL(sqlDB)
	featureStore := store.NewFeatureSQL(sqlDB)
	featureResultStore := store.NewFeatureResultSQL(sqlDB)
	tpsTranscriptionsStore := store.NewTPSTranscriptions()
	diagramCropsStore := store.NewDiagramCropsStore(fileSystemManager, env.FacsimilesGithubRepoUrl)
	datasetImageStore := store.NewDatasetImageStore(fileSystemManager)

	ghDownloader := ghwrapper.NewWrapper(env.GithubToken, env.GithubDownloaderTimeout)

	healthSvc := service.NewHealthService(sqlDB)
	geoSvc := service.NewGeoService(geoStore)
	modelSvc := service.NewModelService(modelStore, fileSystemManager)
	ruleApplier := service.NewAnnotationRuleApplier(modelSvc, fileSystemManager, env.RoboflowAPIKey)
	editionSvc := service.NewEditionService(editionStore, facsimileStore)
	facsimileSvc := service.NewFacsimileService(facsimileStore, ghDownloader, fmt.Sprintf("%s/blob/main/docs", env.FacsimilesGithubRepoUrl))
	datasetSvc := service.NewDatasetService(editionSvc, facsimileSvc, modelSvc, datasetStore, fileSystemManager, ghDownloader)
	annotationSvc := service.NewAnnotationsService(datasetSvc, ruleApplier, fileSystemManager, annotationStore)
	metadataDetailsSvc := service.NewMetadataDetails()
	diagramCropsSvc := service.NewDiagramCropsService(diagramCropsStore)
	featureRevisionSvc := service.NewRevision(featureRevisionStore)
	featureSvc := service.NewFeature(featureStore, featureRevisionStore)
	featureResultSvc := service.NewResult(featureResultStore)
	// func NewExecution(featureRevisionsSvc *Revision, featuresSvc *Feature, featureResultsSvc *Result, annotationSvc *Annotation, store *fpstore.FeatureExecutionSQL, filesysManager *filesys.Manager, datasetImg *DatasetImg, llmClient *llm.Client) *Execution {

	featureExecutionSvc := service.NewExecution(featureRevisionSvc, featureSvc, featureResultSvc, annotationSvc, featureExecutionStore, fileSystemManager, service.NewDatasetImg(datasetSvc, fileSystemManager, datasetImageStore, tpsTranscriptionsStore), llm.NewClient(env.OpenAIAPIKey))
	annotationUploader := service.NewAnnotationsUploader(
		annotationSvc,
		datasetSvc,
		fileSystemManager,
		env.RoboflowAPIKey,
		env.PythonExecutable,
		env.EscriptoriumUsername,
		env.EscriptoriumPassword,
		env.EscriptoriumBasePath,
		env.GithubToken,
		env.CommentariaPath,
	)
	titlePageTEI := service.NewTitlePageTEI(featureResultSvc, tpsTranscriptionsStore)
	annotationTEI := service.NewAnnotationTEI(annotationSvc, fileSystemManager, titlePageTEI)

	metaStoreManager := service.NewMetaStoreManager(
		datasetSvc,
		annotationSvc,
		modelSvc,
		fileSystemManager,
	)
	trainSvc := service.NewTrainService(annotationSvc, modelSvc, fileSystemManager, env.TrainingDir())
	annotationSearch := service.NewAnnotationSearch(annotationSvc, fileSystemManager)

	log.Printf("warming geo cache...")
	if err := geoStore.WarmCache(); err != nil {
		log.Fatalf("geo cache warm failed: %v", err)
	}
	log.Printf("finished warming geo cache")

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

	log.Printf("generating diagram crops metadata...")
	if err := diagramcrops.Generate(env, diagramcrops.Options{}); err != nil {
		log.Printf("warning: failed to generate diagram crops metadata: %v", err)
	}
	log.Printf("finished generating diagram crops metadata")

	deps := &api.Dependencies{
		Env:                 env,
		HealthSvc:           healthSvc,
		EditionSvc:          editionSvc,
		GeoSvc:              geoSvc,
		FacsimileSvc:        facsimileSvc,
		DatasetSvc:          datasetSvc,
		DatasetImgSvc:       service.NewDatasetImg(datasetSvc, fileSystemManager, datasetImageStore, tpsTranscriptionsStore),
		AnnotationSvc:       annotationSvc,
		ModelSvc:            modelSvc,
		TrainSvc:            trainSvc,
		MetadataDetailsSvc:  metadataDetailsSvc,
		MetaStoreManager:    metaStoreManager,
		AnnotationsUploader: annotationUploader,
		AnnotationTEI:       annotationTEI,
		AnnotationSearch:    annotationSearch,
		FeatureSvc:          featureSvc,
		FeatureRevisionSvc:  featureRevisionSvc,
		FeatureResultSvc:    featureResultSvc,
		FeatureExecutionSvc: featureExecutionSvc,
		DiagramCropsSvc:     diagramCropsSvc,
		USTC:                service.NewUSTC(),
		IntegrationJobSvc:   service.NewIntegrationJob(store.NewIntegrationJobStore(cache.NewCache()), annotationUploader),
		VCSMgt:              service.NewVCSMgt(env.ItemsMetadataStoreDir(), fileSystemManager.DatasetImagesDirByID("tps")),
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
