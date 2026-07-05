package api

import (
	"fmt"
	"net/http"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/httpwrapper"
)

func ValidateAuth(r *http.Request) (any, error) {
	userInfo := r.Context().Value(httpwrapper.GitHubUserKey)
	if userInfo == nil {
		return nil, fmt.Errorf("no github user info found in context")
	}

	user, ok := userInfo.(*httpwrapper.GitHubUser)
	if !ok {
		return nil, fmt.Errorf("invalid user info type in context")
	}

	return common.AuthValidateResponse{
		Email:    user.Email,
		Username: user.Login,
	}, nil
}
