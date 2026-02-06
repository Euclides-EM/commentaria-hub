package featureplat

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/MiaMish/elements-dh/ocrflow/internal/docs/featureplat" // swagger docs for featureplat API
)

func NewFeatureAppRouter(deps *Dependencies) http.Handler {
	root := http.NewServeMux()

	api := http.NewServeMux()
	h := NewFeatureAppHandlers(deps)

	api.HandleFunc("/health", httpwrapper.Get(h.Health).Build())

	api.HandleFunc("/auth/validate", httpwrapper.Create(h.ValidateAuth).Build())

	api.HandleFunc("/collections/{collectionId}/features", httpwrapper.Get(h.ListFeatures).Create(h.CreateFeatures).Build())
	api.HandleFunc("/collections/{collectionId}/features/{featureId}", httpwrapper.Delete(h.DeleteFeature).Get(h.GetFeature).Update(h.UpdateFeature).Build())

	api.HandleFunc("/collections/{collectionId}/features/{featureId}/revisions", httpwrapper.Get(h.ListFeatureRevisions).Create(h.CreateFeatureRevision).Build())
	api.HandleFunc("/collections/{collectionId}/features/{featureId}/revisions/{revisionId}", httpwrapper.Get(h.GetFeatureRevision).Build())

	api.HandleFunc("/executions", httpwrapper.Get(h.ListExecutions).Create(h.CreateExecution).Build())
	api.HandleFunc("/executions/{executionId}", httpwrapper.Get(h.GetExecution).Build())
	api.HandleFunc("/executions/{executionId}/cancel", httpwrapper.Update(h.CancelExecution).Build())

	api.HandleFunc("/collections/{collectionId}/results", httpwrapper.Get(h.ListResults).Create(h.CreateResult).Build())

	// GET /collections/{id}/tei?key=Paris_1667&features=feature1,feature2
	api.HandleFunc("/collections/{collectionId}/tei", httpwrapper.GetXML(h.GetTEI).Build())

	// mount API under /api/v1
	root.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	// ---------- Swagger ----------
	root.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.InstanceName("featureplat"),
		httpSwagger.PersistAuthorization(true),
	))

	handler := common.CORSMiddleware(root)
	return handler
}
