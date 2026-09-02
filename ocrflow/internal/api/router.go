package api

import (
	"net/http"
	"path/filepath"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/config"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/service"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/httpwrapper"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Dependencies struct {
	Env                     *config.EnvConfig
	HealthSvc               *service.Health
	LogsSvc                 *service.Logs
	EditionSvc              *service.Edition
	ReprintSvc              *service.Reprint
	FacsimileSvc            *service.Facsimile
	ShelfmarkSvc            *service.Shelfmark
	DatasetSvc              *service.Dataset
	DatasetImgSvc           *service.DatasetImg
	AnnotationSvc           *service.Annotation
	AnnotationRuleExecution *service.AnnotationRuleExecution
	AnnotationGroupSvc      *service.AnnotationGroup
	ModelSvc                *service.Model
	MetadataDetailsSvc      *service.MetadataDetails
	MetaStoreManager        *service.MetaStoreManager
	AnnotationsUploader     *service.AnnotationsUploader
	AnnotationTEI           *service.AnnotationTEI
	EditionTEI              *service.EditionTEI
	EditionTranscriptionSvc *service.EditionTranscription
	AnnotationSearch        *service.AnnotationSearch
	FeatureSvc              *service.Feature
	FeatureRevisionSvc      *service.Revision
	FeatureExecutionSvc     *service.Execution
	FeatureResultSvc        *service.Result
	FeaturePropertySvc      *service.FeatureProperty
	DiagramCropsSvc         *service.DiagramCrops
	USTC                    *service.USTC
	JobSvc                  *service.Job
	ModelTrainingSvc        *service.ModelTraining
	GeoSvc                  *service.Geo
	VCSMgt                  *service.VCSMgt
	BackupSvc               *service.Backup
}

func NewRouter(deps *Dependencies) http.Handler {
	root := http.NewServeMux()

	api := http.NewServeMux()
	h := NewHandlers(deps)

	api.HandleFunc("/health", httpwrapper.Get(h.Health).Build())
	api.HandleFunc("/logs", httpwrapper.Get(h.ListLogs).Build())

	api.HandleFunc("/auth/validate", httpwrapper.Create(h.ValidateAuth).Build())
	api.HandleFunc("/version_control/pull", httpwrapper.Create(h.VersionControlPull).Build())
	api.HandleFunc("/version_control/push", httpwrapper.Create(h.VersionControlPush).Build())
	api.HandleFunc("/catalogs/ustc/lookup", httpwrapper.Create(h.USTCLookup).Build())
	api.HandleFunc("/cities", httpwrapper.Get(h.ListCities).Build())
	api.HandleFunc("/shelfmarks", httpwrapper.Get(h.ListAllShelfmarks).Build())

	api.HandleFunc("/backups", httpwrapper.Get(h.ListBackups).Create(h.CreateBackup).Build())
	api.HandleFunc("/backups/fromzip", httpwrapper.CreateFile(h.CreateBackupFromZip).Build())
	api.HandleFunc("/backups/{backupId}", httpwrapper.GetZip(h.DownloadBackup).Build())
	api.HandleFunc("/backups/{backupId}/sync", httpwrapper.Update(h.SyncBackupToDrive).Build())
	api.HandleFunc("/backups/{backupId}/restore", httpwrapper.Update(h.RestoreLatestBackup).Build())

	api.HandleFunc("/datasets", httpwrapper.Get(h.ListDatasets).Create(h.CreateDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}", httpwrapper.Delete(h.DeleteDataset).Update(h.UpdateDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/suggested_rules", httpwrapper.Get(h.ListSuggestedRulesForDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/suggested_reviews", httpwrapper.Get(h.ListSuggestedReviewForDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/images", httpwrapper.Get(h.GetDatasetImages).Delete(h.DeleteDatasetImages).Build())
	api.HandleFunc("/datasets/{dataSetId}/images/upload", httpwrapper.CreateFile(h.UploadDatasetImage).Build())
	api.HandleFunc("/datasets/{dataSetId}/images/{pageNumOrKey}", h.ServeDatasetImage)
	api.HandleFunc("/datasets/{dataSetId}/annotations", httpwrapper.Get(h.ListAnnotations).Create(h.CreateAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}", httpwrapper.Delete(h.DeleteAnnotation).Update(h.UpdateAnnotation).Get(h.GetAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/fromzip", httpwrapper.CreateFile(h.GetAnnotationZipFile).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/fromurl", httpwrapper.CreateFile(h.GetAnnotationURL).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/duplicate", httpwrapper.Create(h.DuplicateAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/merge", httpwrapper.Update(h.MergeAnnotation).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/index", httpwrapper.Get(h.GetAnnotationIndex).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/categories", httpwrapper.Get(h.ListAnnotationCategories).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply", httpwrapper.Update(h.ApplyRules).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/model_detect", httpwrapper.Update(h.ApplyRuleModelDetect).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/slice_pages", httpwrapper.Update(h.ApplyRuleSlicePages).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/stretch", httpwrapper.Update(h.ApplyRuleStretch).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/add_margin", httpwrapper.Update(h.ApplyRuleAddMargin).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/detect_lines", httpwrapper.Update(h.ApplyRuleDetectLines).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/remove_categories", httpwrapper.Update(h.ApplyRuleRemoveCategories).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/rename_categories", httpwrapper.Update(h.ApplyRuleRenameCategories).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/remove_overlap", httpwrapper.Update(h.ApplyRuleRemoveOverlap).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/resolve_overlap_with_priority", httpwrapper.Update(h.ApplyRuleResolveOverlapWithPriority).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/recategorize_by_alignment", httpwrapper.Update(h.ApplyRuleRecategorizeByAlignment).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/limit_category_zones", httpwrapper.Update(h.ApplyRuleLimitCategoryZones).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/reassign_text_lines_by_tolerance", httpwrapper.Update(h.ApplyRuleReassignTextLinesByTolerance).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/text_block_corrections", httpwrapper.Update(h.ApplyRuleTextBlockCorrections).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/llm_transcription_corrector", httpwrapper.Update(h.ApplyRuleLLMTranscriptionCorrector).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/detection_upload", httpwrapper.CreateFile(h.UploadAnnotationDetectionResult).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/review", httpwrapper.Create(h.CreateAnnotationReview).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/search", httpwrapper.Get(h.SearchAnnotation).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/tei/{pageNumOrKey}", httpwrapper.GetXML(h.GetAnnotationTEI).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/teis", httpwrapper.GetXML(h.GetAnnotationTEIs).Build())

	api.HandleFunc("/annotation_groups", httpwrapper.Get(h.ListAnnotationGroups).Create(h.CreateAnnotationGroup).Build())
	api.HandleFunc("/annotation_groups/{groupId}", httpwrapper.Get(h.GetAnnotationGroup).Update(h.UpdateAnnotationGroup).Delete(h.DeleteAnnotationGroup).Build())

	api.HandleFunc("/features/properties", httpwrapper.Get(h.ListFeatureProperties).Build())
	api.HandleFunc("/feature_executions", httpwrapper.Get(h.ListExecutions).Create(h.CreateExecution).Build())
	api.HandleFunc("/feature_executions/{executionId}", httpwrapper.Get(h.GetExecution).Build())
	api.HandleFunc("/feature_executions/{executionId}/cancel", httpwrapper.Update(h.CancelExecution).Build())
	api.HandleFunc("/features", httpwrapper.Get(h.ListFeatures).Create(h.CreateFeature).Delete(h.DeleteFeatures).Build())
	api.HandleFunc("/features/{featureId}", httpwrapper.Delete(h.DeleteFeature).Get(h.GetFeature).Update(h.UpdateEditionsFeature).Build())

	api.HandleFunc("/features/{featureId}/revisions", httpwrapper.Get(h.ListFeatureRevisions).Create(h.CreateFeatureRevision).Build())
	api.HandleFunc("/features/{featureId}/revisions/{revisionId}", httpwrapper.Get(h.GetFeatureRevision).Build())

	api.HandleFunc("/features_results", httpwrapper.Get(h.ListFeaturesResults).Create(h.CreateFeaturesResult).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/results", httpwrapper.Get(h.ListDatasetResults).Create(h.CreateDatasetResult).Build())

	api.HandleFunc("/editions", httpwrapper.Create(h.CreateEdition).Build())

	api.HandleFunc("/editions/search", httpwrapper.Create(h.ListEditions).Build())
	api.HandleFunc("/editions/reprints/detect", httpwrapper.Create(h.DetectEditionReprints).Build())
	api.HandleFunc("/editions/reprints/apply", httpwrapper.Create(h.ApplyEditionReprints).Build())

	api.HandleFunc("/editions/transcriptions", httpwrapper.Get(h.ListEditionTranscriptionsDetails).Build())

	api.HandleFunc("/editions/{editionId}/notes", httpwrapper.Create(h.CreateEditionNote).Build())
	api.HandleFunc("/editions/{editionId}/results", httpwrapper.Get(h.ListEditionResults).Build())
	api.HandleFunc("/editions/{editionId}/diagrams", httpwrapper.Get(h.GetEditionDiagramCrops).Build())
	api.HandleFunc("/editions/{editionId}/shelfmarks", httpwrapper.Get(h.ListShelfmarks).Create(h.UpsertShelfmark).Build())
	api.HandleFunc("/editions/{editionId}/shelfmarks/{shelfmarkId}", httpwrapper.Update(h.UpdateShelfmark).Delete(h.DeleteShelfmark).Build())
	api.HandleFunc("/editions/{editionId}/tei/{pageNum}", httpwrapper.GetXML(h.GetEditionTEI).Build())

	api.HandleFunc("/editions/{editionId}", httpwrapper.Get(h.GetEdition).Update(h.UpdateEdition).Delete(h.DeleteEdition).Build())
	api.HandleFunc("/editions/{editionId}/transcriptions", httpwrapper.Update(h.UpdateEditionTranscriptionsDetails).Build())

	api.HandleFunc("/facsimilies", httpwrapper.Get(h.ListFacsimiles).Create(h.CreateFacsimile).Build())
	api.HandleFunc("/facsimilies/import-from-drive", httpwrapper.Create(h.ImportFacsimilesFromDrive).Build())
	api.HandleFunc("/facsimilies/mapping-csv", httpwrapper.GetZip(h.DownloadFacsimileMappingCSV).CreateFile(h.UploadFacsimileMappingCSV).Build())
	api.HandleFunc("/facsimilies/{id}/diagrams", httpwrapper.Get(h.GetFacsimileDiagramCrops).Build())
	api.HandleFunc("/facsimilies/{id}/pdf", httpwrapper.GetFile(h.DownloadFacsimilePDF, "application/pdf").Build())
	api.HandleFunc("/facsimilies/{id}", httpwrapper.Get(h.GetFacsimile).Update(h.UpdateFacsimile).Build())

	api.HandleFunc("/integrations/platforms", httpwrapper.Get(h.ListIntegrationPlatforms).Build())
	api.HandleFunc("/jobs", httpwrapper.Get(h.ListJobs).Create(h.CreateJobs).Build())
	api.HandleFunc("/jobs/{jobId}", httpwrapper.Get(h.GetJob).Build())

	api.HandleFunc("/models", httpwrapper.Get(h.ListModels).Build())
	api.HandleFunc("/models_train", httpwrapper.Create(h.TrainModel).Build())
	api.HandleFunc("/models_upload", httpwrapper.CreateFile(h.UploadModel).Build())
	api.HandleFunc("/models/{id}", httpwrapper.Delete(h.DeleteModel).Update(h.UpdateModel).Build())

	api.HandleFunc("/annotation_rules", httpwrapper.Get(h.ListAnnotationRules).Build())
	api.HandleFunc("/pipeline_stages", httpwrapper.Get(h.ListPipelineStages).Build())

	api.HandleFunc("/store/cleanup/local", httpwrapper.Delete(h.CleanupLocalStore).Build())

	api.HandleFunc("/ollama-proxy/api/generate", httpwrapper.Create(h.CreateOllamaRequest).Build())
	// mount API under /api/v1
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	// ---------- Swagger ----------
	root.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
	))

	// ---------- Static docs/images ----------
	root.Handle("/store/data/",
		http.StripPrefix("/store/data/",
			http.FileServer(http.Dir(deps.Env.DataDir())),
		),
	)
	if filepath.IsAbs(deps.Env.FacsimilesDiagramsPath) {
		root.Handle("/facsimiles/diagrams/",
			http.StripPrefix("/facsimiles/diagrams/",
				http.FileServer(http.Dir(deps.Env.FacsimilesDiagramsPath)),
			),
		)
	}

	handler := CORSMiddleware(deps.Env.AllowedOriginsCORSList(), root)
	return handler
}
