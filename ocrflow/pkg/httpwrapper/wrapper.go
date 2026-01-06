package httpwrapper

import (
	"encoding/json"
	"net/http"
	"os"
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
	get    func(w http.ResponseWriter, r *http.Request)
	post   func(w http.ResponseWriter, r *http.Request)
	put    func(w http.ResponseWriter, r *http.Request)
	delete func(w http.ResponseWriter, r *http.Request)
}

func Get(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Get(f)
}

func GetXML(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetXML(f)
}

func GetPNG(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetPNG(f)
}

func GetZip(f func(*http.Request) (zipPath string, deleteAfterServe bool, err error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetZip(f)
}

func Create(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Create(f)
}

func CreateFile(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.CreateFile(f)
}

func Update(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Update(f)
}

func Delete(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.Delete(f)
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

func (wb *wrapperBuilder) GetXML(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return wb
}

func (wb *wrapperBuilder) GetPNG(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "image/png")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return wb
}

func (wb *wrapperBuilder) GetZip(f func(r *http.Request) (zipPath string, deleteAfterServe bool, err error)) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		zipPath, deleteAfterServe, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"data.zip\"")
		http.ServeFile(w, r, zipPath)
		if deleteAfterServe {
			_ = os.Remove(zipPath)
		}
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
		if resp == nil {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		writeJSON(w, http.StatusCreated, resp)
	}
	return wb
}

func (wb *wrapperBuilder) CreateFile(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb.post = func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
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

func (wb *wrapperBuilder) Delete(f func(*http.Request) (any, error)) *wrapperBuilder {
	wb.delete = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if resp != nil {
			writeJSON(w, http.StatusOK, resp)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	return wb
}

func (wb *wrapperBuilder) Build() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !wb.authorized(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
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
		case http.MethodDelete:
			if wb.delete != nil {
				wb.delete(w, r)
				return
			}
		}
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}
