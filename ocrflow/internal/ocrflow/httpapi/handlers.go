package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/docs" // swagger docs
	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
)

type Handlers struct {
	deps *Dependencies
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{deps: deps}
}

func extractDatasetAndAnnotationIDs(r *http.Request) (string, string, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return "", "", err
	}
	annotationID := r.PathValue("id")
	if annotationID == "" {
		return "", "", fmt.Errorf("missing annotation ID")
	}
	return datasetID, annotationID, nil
}

func extractDatasetID(r *http.Request) (string, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return "", fmt.Errorf("missing dataset ID")
	}
	return datasetID, nil
}

func decodeBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("failed to decode request body to type %T: %w", dst, err)
	}
	return nil
}

type AuthValidateResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
}

// ValidateAuth godoc
// @Summary Validate authentication token
// @Description Validates the provided Bearer token and returns user information
// @Tags Authentication
// @Produce json
// @Security BearerAuth
// @Success 200 {object} AuthValidateResponse
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/validate [post]
func (h *Handlers) ValidateAuth(r *http.Request) (any, error) {
	userInfo := r.Context().Value(httpwrapper.GitHubUserKey)
	if userInfo == nil {
		return nil, fmt.Errorf("no github user info found in context")
	}

	user, ok := userInfo.(*httpwrapper.GitHubUser)
	if !ok {
		return nil, fmt.Errorf("invalid user info type in context")
	}

	return AuthValidateResponse{
		Email:    user.Email,
		Username: user.Login,
	}, nil
}
