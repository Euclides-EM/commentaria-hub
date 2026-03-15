package api

import (
	"net/http"

	"github.com/samber/lo"
)

func CORSMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	ao := lo.SliceToMap(allowedOrigins, func(origin string) (string, bool) {
		return origin, true
	})
	isAllowedOrigin := func(origin string) bool {
		return ao["*"] || ao[origin]
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if r.Method == http.MethodOptions {
			if isAllowedOrigin(origin) {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			http.Error(w, "CORS origin not allowed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
