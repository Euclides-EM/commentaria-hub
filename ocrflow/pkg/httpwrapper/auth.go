package httpwrapper

import (
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

var cache *ttlcache.Cache[string, bool]

func init() {
	cache = ttlcache.New[string, bool](
		ttlcache.WithTTL[string, bool](30 * time.Minute),
	)

	go cache.Start()
}

type githubUser struct {
	Login string `json:"login"`
}

func authorized(r *http.Request) bool {
	if lo.Contains([]string{http.MethodGet, http.MethodHead, http.MethodOptions}, r.Method) {
		return true
	}

	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" {
		return false
	}
	token = strings.TrimSpace(token)

	if cache.Get(token) != nil && cache.Get(token).Value() {
		return true
	}

	allowed := authInGit(token)
	cache.Set(token, allowed, ttlcache.DefaultTTL)
	return allowed
}

func authInGit(token string) bool {
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
