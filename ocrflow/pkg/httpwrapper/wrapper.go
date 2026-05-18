package httpwrapper

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return
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

func GetSQL(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetSQL(f)
}

func GetPNG(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetPNG(f)
}

func GetZip(f func(*http.Request) (zipPath string, deleteAfterServe bool, err error)) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetZip(f)
}

func GetFile(f func(*http.Request) (filePath string, downloadName string, err error), contentType string) *wrapperBuilder {
	wb := &wrapperBuilder{}
	return wb.GetFile(f, contentType)
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
			return
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
			return
		}
		w.Header().Set("Content-Type", "application/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
	return wb
}

func (wb *wrapperBuilder) GetSQL(f func(*http.Request) ([]byte, error)) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
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
			return
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
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(zipPath)+"\"")
		http.ServeFile(w, r, zipPath)
		if deleteAfterServe {
			_ = os.Remove(zipPath)
		}
	}
	return wb
}

func (wb *wrapperBuilder) GetFile(f func(r *http.Request) (filePath string, downloadName string, err error), contentType string) *wrapperBuilder {
	wb.get = func(w http.ResponseWriter, r *http.Request) {
		filePath, downloadName, err := f(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		if downloadName == "" {
			downloadName = filepath.Base(filePath)
		}
		w.Header().Set("Content-Disposition", "attachment; filename=\""+downloadName+"\"")
		http.ServeFile(w, r, filePath)
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
		origin := r.Header.Get("Origin")
		httpAddr := os.Getenv("HTTP_ADDR")
		if httpAddr == "" {
			httpAddr = ":8085"
		}
		serverPort := strings.TrimPrefix(httpAddr, ":")
		allowedOrigins := []string{
			"http://localhost:5180",
			"http://localhost:5181",
			"http://localhost:5173",
			"http://localhost:" + serverPort,
		}

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				break
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		authedReq, isAuthorized := authorized(r)
		if !isAuthorized {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		r = authedReq
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
