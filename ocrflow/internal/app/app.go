package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api"
	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/diagramcrops"
	"github.com/MiaMish/elements-dh/ocrflow/internal/migrations"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/titlepage"
	"github.com/MiaMish/elements-dh/ocrflow/internal/service"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/db"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
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
	futils.InitTemp(env.TmpDir())
	if err = store.InitMountProjToStore(env.RootDir, env.StoreDir); err != nil {
		return nil, fmt.Errorf("init mount proj to store: %w", err)
	}
	var sqlDB *sql.DB
	bckSvc := service.NewBackup(env.DataDir(), env.ModelsDir(), env.ItemsMetadataStoreDir(), env.DBPath(), env.BackupDir(), env.RestoreDir(), env.RcloneRemoteName, env.BackupGDriveFolderID, env.BackupMaxToStore, func() error {
		log.Printf("shutting down app for backup/restore...")
		if sqlDB != nil {
			log.Printf("closing db connection for backup/restore...")
			if err := sqlDB.Close(); err != nil {
				return fmt.Errorf("close db: %w", err)
			}
			log.Printf("finished closing db connection for backup/restore")
		}
		// Kill the app process to ensure a clean state for backup/restore.
		// The app will be restarted by the process manager (e.g. systemd) after backup/restore is complete.
		log.Printf("exiting process for backup/restore...")
		os.Exit(0)
		return nil
	})

	log.Printf("starting app for backup/restore if needed...")
	if err := bckSvc.RestoreLatestBackupIfRelevant(); err != nil {
		return nil, fmt.Errorf("handle restore if needed: %w", err)
	}
	log.Printf("finished app for backup/restore if needed")

	fileSystemManager := filesys.NewFileSystemManager(env.DataDir(), env.TrainingDir(), env.ModelsDir(), env.DiagramsDir())
	geoStore := store.NewGeoCSV(env.ItemsMetadataStoreDir())
	sqlDB, err = db.InitDB(env.DBPath(), migrations.Migrations, "ocrflow", env.OptionalMigrations())
	if err != nil {
		return nil, fmt.Errorf("init db: %w", err)
	}
	// So backups include all data when SQLite uses WAL: flush WAL into main DB before copying.
	bckSvc.SetCheckpointFunc(func() error {
		_, err := sqlDB.Exec("PRAGMA wal_checkpoint(FULL)")
		return err
	})
	editionPreferredTranscriptionStore := store.NewEditionPreferredAnnotationSql(sqlDB)
	editionStore := store.NewEditionCSV(env.ItemsMetadataStoreDir(), editionPreferredTranscriptionStore.OnDeleteEdition)
	facsimileStore := store.NewFacsimileSql(sqlDB, env.ItemsMetadataStoreDir())
	datasetStore := store.NewDatasetSQL(sqlDB, fileSystemManager)
	annotationStore := store.NewAnnotationSQL(sqlDB)
	annotationGroupStore := store.NewAnnotationGroupSQL(sqlDB)
	modelStore := store.NewModelSQL(sqlDB)
	featureRevisionStore := store.NewFeatureRevisionSQL(sqlDB)
	featureExecutionStore := store.NewFeatureExecutionStore(cache.NewCache())
	featureStore := store.NewFeatureSQL(sqlDB)
	featureResultStore := store.NewFeatureResultSQL(sqlDB)
	diagramCropsStore := store.NewDiagramCropsStore(fileSystemManager, env.FacsimilesDiagramsURL)
	datasetImageStore := store.NewDatasetImageStore(fileSystemManager)

	vcsMgtSvc := service.NewVCSMgt(
		env.RootDir,
		filepath.Join(env.RootDir, "store", "items_metadata"),
		filepath.Join(env.RootDir, "store", "data", titlepage.DatasetID, "imgs"),
	)
	healthSvc := service.NewHealthService(sqlDB, vcsMgtSvc)
	logsSvc := service.NewLogsService(env.LogsSystemdUnit, env.LogsTailDefaultLines, env.LogsTailMaxLines)
	geoSvc := service.NewGeoService(geoStore)
	modelSvc := service.NewModelService(modelStore, fileSystemManager)
	ruleApplier := service.NewAnnotationRuleApplier(modelSvc, fileSystemManager, env.RoboflowAPIKey)
	editionSvc := service.NewEditionService(editionStore, facsimileStore)
	facsimileSvc := service.NewFacsimileService(
		facsimileStore,
		env.FacsimilesPDFDir,
		env.FacsimilesDiagramsPath,
		env.DiagramsDir(),
		env.ItemsMetadataStoreDir(),
		env.FacsimilesRemoteAPIURL,
		env.GithubToken,
		env.RcloneRemoteName,
		env.FacsimilesGDriveFolderID,
	)
	datasetSvc := service.NewDatasetService(
		editionSvc,
		facsimileSvc,
		modelSvc,
		datasetStore,
		fileSystemManager,
		env.GithubToken,
		env.DatasetCreateMaxParallel,
		env.DatasetCreateQueueWait,
	)
	datasetImgSvc := service.NewDatasetImg(datasetSvc, fileSystemManager, datasetImageStore, editionSvc)
	featureProperty := service.NewFeatureProperty()
	featureSvc := service.NewFeature(featureStore, featureRevisionStore, featureProperty)
	featureResultSvc := service.NewResult(featureResultStore, featureSvc, featureProperty)
	annotationSvc := service.NewAnnotationsService(datasetSvc, datasetImgSvc, ruleApplier, featureResultSvc, fileSystemManager, annotationStore)
	annotationGroupSvc := service.NewAnnotationGroupService(annotationSvc, annotationGroupStore)
	editionTranscriptionSvc := service.NewEditionTranscription(editionPreferredTranscriptionStore, editionSvc, datasetSvc, annotationSvc)
	metadataDetailsSvc := service.NewMetadataDetails()
	diagramCropsSvc := service.NewDiagramCropsService(diagramCropsStore)
	featureRevisionSvc := service.NewRevision(featureRevisionStore, featureProperty)
	annotationTEI := service.NewAnnotationTEI(annotationSvc, datasetSvc, fileSystemManager, datasetImgSvc, featureResultSvc, featureSvc, editionSvc)
	titlePageProvisionSvc := service.NewTitlePageProvision(annotationSvc, datasetSvc, editionSvc)
	langResolver := service.NewLanguagesResolver(editionSvc, datasetSvc)

	featureExecutionSvc := service.NewExecution(featureRevisionSvc, featureSvc, featureResultSvc, annotationSvc, annotationTEI, editionSvc, langResolver, featureProperty, featureExecutionStore, fileSystemManager, service.NewDatasetImg(datasetSvc, fileSystemManager, datasetImageStore, editionSvc), llm.NewClient(env.OpenAIAPIKey, env.OllamaBaseURL))
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
	editionTEI := service.NewEditionTEI(fileSystemManager, editionSvc)
	metaStoreManager := service.NewMetaStoreManager(
		datasetSvc,
		annotationSvc,
		modelSvc,
		fileSystemManager,
	)
	trainSvc := service.NewTrainService(annotationSvc, modelSvc, fileSystemManager, env.TrainingDir())
	annotationSearch := service.NewAnnotationSearch(annotationSvc, fileSystemManager, featureResultSvc, annotationTEI, datasetImgSvc)

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

	log.Printf("updating facsimiles from configured source...")
	if err := facsimileSvc.UpdateFromConfiguredSource(); err != nil {
		log.Printf("warning: failed to update facsimiles from configured source: %v", err)
	}
	log.Printf("finished updating facsimiles from configured source")

	log.Printf("generating diagram crops metadata...")
	if err := diagramcrops.Generate(env, diagramcrops.Options{}); err != nil {
		log.Printf("warning: failed to generate diagram crops metadata: %v", err)
	}
	log.Printf("finished generating diagram crops metadata")

	log.Printf("update title page annotations by metadata info...")
	if err := titlePageProvisionSvc.UpdateTitlePageAnnotationsByMetadataInfo(); err != nil {
		log.Printf("warning: failed to update title page annotations by metadata info: %v", err)
	}
	log.Printf("finished updating title page annotations by metadata info")

	deps := &api.Dependencies{
		Env:                     env,
		HealthSvc:               healthSvc,
		LogsSvc:                 logsSvc,
		EditionSvc:              editionSvc,
		GeoSvc:                  geoSvc,
		FacsimileSvc:            facsimileSvc,
		DatasetSvc:              datasetSvc,
		DatasetImgSvc:           datasetImgSvc,
		AnnotationSvc:           annotationSvc,
		AnnotationGroupSvc:      annotationGroupSvc,
		ModelSvc:                modelSvc,
		TrainSvc:                trainSvc,
		MetadataDetailsSvc:      metadataDetailsSvc,
		MetaStoreManager:        metaStoreManager,
		AnnotationsUploader:     annotationUploader,
		AnnotationTEI:           annotationTEI,
		EditionTEI:              editionTEI,
		EditionTranscriptionSvc: editionTranscriptionSvc,
		AnnotationSearch:        annotationSearch,
		FeatureSvc:              featureSvc,
		FeatureRevisionSvc:      featureRevisionSvc,
		FeatureResultSvc:        featureResultSvc,
		FeatureExecutionSvc:     featureExecutionSvc,
		FeaturePropertySvc:      service.NewFeatureProperty(),
		DiagramCropsSvc:         diagramCropsSvc,
		USTC:                    service.NewUSTC(),
		JobSvc:                  service.NewJob(store.NewJobStore(cache.NewCache()), annotationUploader, facsimileSvc, bckSvc),
		VCSMgt:                  vcsMgtSvc,
		BackupSvc:               bckSvc,
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
	futils.CleanTemp()
	return nil
}
