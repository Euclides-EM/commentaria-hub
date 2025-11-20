package httpapi

import (
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/service"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
)

type Dependencies struct {
	Env            *config.EnvConfig
	HealthService  *service.Health
	EditionService *service.Edition
}

func NewRouter(deps *Dependencies) http.Handler {
	mux := http.NewServeMux()

	h := NewHandlers(deps)

	mux.HandleFunc("/health", httpwrapper.Get(h.Health))
	mux.HandleFunc("/editions", httpwrapper.Get(h.ListEditions))
	mux.HandleFunc("/editions/{key}/facsimiles/{id}/download", httpwrapper.Get(h.ListEditions))
	// mux.HandleFunc("/editions/", h.GetEdition) // simple path based on prefix

	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	return mux
}
