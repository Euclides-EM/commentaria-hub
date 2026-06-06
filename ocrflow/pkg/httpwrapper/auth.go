package httpwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"github.com/samber/lo"
)

var allowList = map[string]struct{}{
	"reallyliri": {},
	"miamish":    {},
}

var cache *ttlcache.Cache[string, *GitHubUser]

func init() {
	cache = ttlcache.New[string, *GitHubUser](
		ttlcache.WithTTL[string, *GitHubUser](30 * time.Minute),
	)

	go cache.Start()
}

type GitHubUser struct {
	Login string `json:"login"`
	Email string `json:"email"`
}

type contextKey string

const GitHubTokenKey contextKey = "github_token"
const GitHubUserKey contextKey = "github_user"

func authorized(r *http.Request) (*http.Request, bool) {
	if lo.Contains([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, r.Method) && isPublicReadPath(r.URL.Path) {
		return r, true
	}
	if strings.HasSuffix(r.URL.Path, "/search") && r.Method == http.MethodPost {
		return r, true
	}

	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return r, false
	}
	token = strings.TrimSpace(token)

	if cacheItem := cache.Get(token); cacheItem != nil {
		user := cacheItem.Value()
		if user != nil {
			ctx := context.WithValue(r.Context(), GitHubTokenKey, token)
			ctx = context.WithValue(ctx, GitHubUserKey, user)
			return r.WithContext(ctx), true
		}
	}

	user, allowed := authInGithub(token)
	if allowed {
		cache.Set(token, user, ttlcache.DefaultTTL)
		ctx := context.WithValue(r.Context(), GitHubTokenKey, token)
		ctx = context.WithValue(ctx, GitHubUserKey, user)
		return r.WithContext(ctx), true
	}
	return r, false
}

func isPublicReadPath(path string) bool {
	if strings.HasSuffix(path, "/logs") {
		return false
	}
	if strings.HasPrefix(path, "/facsimilies/") && strings.HasSuffix(path, "/pdf") {
		return false
	}
	if strings.HasPrefix(path, "/editions/") && strings.HasSuffix(path, "/facsimile.pdf") {
		return false
	}
	return true
}

func authInGithub(token string) (*GitHubUser, bool) {
	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.github.com/user",
		nil,
	)
	if err != nil {
		return nil, false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, false
	}

	if user.Login == "" {
		return nil, false
	}

	_, allowed := allowList[strings.ToLower(user.Login)]
	return &user, allowed
}
