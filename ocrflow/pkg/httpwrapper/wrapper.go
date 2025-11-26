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

type wrapperBuilder struct {
	get  func(w http.ResponseWriter, r *http.Request)
	post func(w http.ResponseWriter, r *http.Request)
	put  func(w http.ResponseWriter, r *http.Request)
}

func Get(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Get(f)
}

func Create(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Create(f)
}

func Update(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Update(f)
}

func (wb *wrapperBuilder) Get(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		writeJSON(w, http.StatusOK, resp)
	}
	return wb
}

func (wb *wrapperBuilder) Create(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb.post = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, resp)
	}
	return wb
}

func (wb *wrapperBuilder) Update(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb.put = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
	return wb
}

func (wb *wrapperBuilder) Build() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if wb.get != nil {
				wb.get(w, r)
				return
			}
		case http.MethodPost:
			if wb.post != nil {
				wb.post(w, r)
				return
			}
		case http.MethodPut:
			if wb.put != nil {
				wb.put(w, r)
				return
			}
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
