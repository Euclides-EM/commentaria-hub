package httpapi

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Dependencies struct {
	Env                 *config.EnvConfig
	HealthSvc           *service.Health
	EditionSvc          *service.Edition
	DatasetSvc          *service.Dataset
	AnnotationSvc       *service.Annotation
	ModelSvc            *service.Model
	TrainSvc            *service.Train
	MetaStoreManager    *service.MetaStoreManager
	AnnotationsUploader *service.AnnotationsUploader
	AnnotationTEI       *service.AnnotationTEI
	AssetGen            *service.AssetGen
}

func NewRouter(deps *Dependencies) http.Handler {
	root := http.NewServeMux()

	api := http.NewServeMux()
	h := NewHandlers(deps)

	api.HandleFunc("/health", httpwrapper.Get(h.Health).Build())

	api.HandleFunc("/editions", httpwrapper.Get(h.ListEditions).Build())

	api.HandleFunc("/datasets", httpwrapper.Get(h.ListDatasets).Create(h.CreateDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/suggested_rules", httpwrapper.Get(h.ListSuggestedRulesForDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/suggested_reviews", httpwrapper.Get(h.ListSuggestedReviewForDataset).Build())
	api.HandleFunc("/datasets/{dataSetId}/images/{pageNum}", httpwrapper.GetPNG(h.GetPageImage).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations", httpwrapper.Get(h.ListAnnotations).Create(h.CreateAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}", httpwrapper.Delete(h.DeleteAnnotation).Update(h.UpdateAnnotation).Get(h.GetAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/fromzip", httpwrapper.CreateFile(h.GetAnnotationZipFile).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/fromurl", httpwrapper.CreateFile(h.GetAnnotationURL).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/duplicate", httpwrapper.CreateFile(h.DuplicateAnnotation).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/upload/roboflow", httpwrapper.Update(h.UploadToRoboflow).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/upload/escriptorium", httpwrapper.Update(h.UploadToEscriptorium).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/index", httpwrapper.Get(h.GetAnnotationIndex).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/categories", httpwrapper.Get(h.ListAnnotationCategories).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/download_assets", httpwrapper.GetZip(h.DownloadAnnotationAssets).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply", httpwrapper.Update(h.ApplyRules).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/segment", httpwrapper.Update(h.ApplyRuleSegment).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/slice_pages", httpwrapper.Update(h.ApplyRuleSlicePages).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/stretch", httpwrapper.Update(h.ApplyRuleStretch).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/add_margin", httpwrapper.Update(h.ApplyRuleAddMargin).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/detect_lines", httpwrapper.Update(h.ApplyRuleDetectLines).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/remove_categories", httpwrapper.Update(h.ApplyRuleRemoveCategories).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/remove_overlap", httpwrapper.Update(h.ApplyRuleRemoveOverlap).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/reassign_text_lines_by_tolerance", httpwrapper.Update(h.ApplyRuleReassignTextLinesByTolerance).Build())
	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply/text_block_corrections", httpwrapper.Update(h.ApplyRuleTextBlockCorrections).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/review", httpwrapper.Create(h.CreateAnnotationReview).Build())

	api.HandleFunc("/datasets/{dataSetId}/annotations/{id}/tei/{pageNum}", httpwrapper.GetXML(h.GetAnnotationTEI).Build())

	api.HandleFunc("/models", httpwrapper.Get(h.ListModels).Build())
	api.HandleFunc("/train", httpwrapper.Create(h.TrainModel).Build())

	api.HandleFunc("/store/cleanup/local", httpwrapper.Delete(h.CleanupLocalStore).Build())

	// mount API under /api/v1
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	// ---------- Swagger ----------
	root.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.PersistAuthorization(true),
	))

	// ---------- Viewer UI ----------
	root.Handle("/ui/",
		http.StripPrefix("/ui/",
			http.FileServer(http.Dir(deps.Env.UI_DIR)),
		),
	)

	return root

}
