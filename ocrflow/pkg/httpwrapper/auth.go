package httpwrapper

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/samber/lo"
)

var allowList = map[string]struct{}{
	"reallyliri": {},
	"miamish":    {},
}

type githubUser struct {
	Login string `json:"login"`
}

func (wb *wrapperBuilder) authorized(r *http.Request) bool {
	if lo.Contains([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, r.Method) {
		return true
	}

	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return false
	}
	token = strings.TrimSpace(token)

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var user githubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return false
	}

	if user.Login == "" {
		return false
	}

	_, allowed := allowList[strings.ToLower(user.Login)]
	return allowed
}
