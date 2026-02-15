package api

import (
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/httpwrapper"
)

// VersionControlPull godoc
// @Summary Pull latest changes from the repository
// @Description Pulls the latest changes from the repository. If the current branch is not main, it will check out main before pulling. Requires GitHub token in the Authorization header.
// @Tags Version Control
// @Accept json
// @Produce json
// @Success 200 {object} model.VCSStatus
// @Security 	 BearerAuth
// @Router  /version_control/pull [post]
func (h *Handlers) VersionControlPull(r *http.Request) (any, error) {
	token, _ := r.Context().Value(httpwrapper.GitHubTokenKey).(string)
	if token == "" {
		return nil, fmt.Errorf("authorization required for repo operations")
	}
	return h.deps.VCSMgt.Pull(token)
}

// VersionControlPush godoc
// @Summary Push local changes to the repository
// @Description Pushes local changes to the repository. This will first pull the latest changes to ensure the local branch is up to date, then push any local commits. Requires GitHub token in the Authorization header.
// @Tags Version Control
// @Accept json
// @Produce json
// @Success 200 {object} model.VCSStatus
// @Security 	 BearerAuth
// @Router  /version_control/push [post]
func (h *Handlers) VersionControlPush(r *http.Request) (any, error) {
	token, _ := r.Context().Value(httpwrapper.GitHubTokenKey).(string)
	return h.deps.VCSMgt.Push(token)
}
