package httpwrapper

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

}

func Get[T any](f func(*http.Request) (T, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func Upsert(f func(*http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
