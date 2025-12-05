package httpapi

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
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
}

func NewRouter(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	h := NewHandlers(deps)

	mux.HandleFunc("/health", httpwrapper.Get(h.Health).Build())

	mux.HandleFunc("/editions", httpwrapper.Get(h.ListEditions).Build())

	mux.HandleFunc("/datasets", httpwrapper.Get(h.ListDatasets).Create(h.CreateDataset).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations", httpwrapper.Get(h.ListAnnotations).Create(h.CreateAnnotation).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/fromzip", httpwrapper.CreateFile(h.GetAnnotationZipFile).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/fromurl", httpwrapper.CreateFile(h.GetAnnotationURL).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/{id}/convert", httpwrapper.Update(h.ConvertAnnotations).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/{id}/apply", httpwrapper.Update(h.ApplyRules).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/{id}/upload/roboflow", httpwrapper.Update(h.UploadToRoboflow).Build())
	mux.HandleFunc("/datasets/{dataSetId}/annotations/{id}/upload/escriptorium", httpwrapper.Update(h.UploadToEscriptorium).Build())

	mux.HandleFunc("/models", httpwrapper.Get(h.ListModels).Build())
	mux.HandleFunc("/train", httpwrapper.Create(h.TrainModel).Build())

	mux.HandleFunc("/store/cleanup/local", httpwrapper.Delete(h.CleanupLocalStore).Build())

	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return mux
}
