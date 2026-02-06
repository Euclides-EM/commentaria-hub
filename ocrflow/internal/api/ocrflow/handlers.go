package ocrflow

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/api/common"
)

type Handlers struct {
	deps *Dependencies
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{deps: deps}
}

// ValidateAuth godoc
// @Summary Validate authentication token
// @Description Validates the provided Bearer token and returns user information
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} common.AuthValidateResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/validate [post]
func (h *Handlers) ValidateAuth(r *http.Request) (any, error) {
	return common.ValidateAuth(r)
}
