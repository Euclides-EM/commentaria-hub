package httpwrapper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetStopsAfterHandlerError(t *testing.T) {
	handler := Get(func(r *http.Request) (any, error) {
		return nil, errors.New("boom")
	}).Build()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	if body := rec.Body.String(); !strings.Contains(body, "boom") {
		t.Fatalf("body = %q, expected error message", body)
	}
}

func TestGetDownloadServesRequestedNameAndCleansUp(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "temporary-name.zip")
	if err := os.WriteFile(filePath, []byte("backup"), 0o644); err != nil {
		t.Fatal(err)
	}
	cleaned := false
	handler := GetDownload(func(r *http.Request) (string, string, func(), error) {
		return filePath, "euclides_backup.zip", func() { cleaned = true }, nil
	}, "application/zip").Build()

	req := httptest.NewRequest(http.MethodGet, "/backups/example", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if disposition := rec.Header().Get("Content-Disposition"); disposition != `attachment; filename="euclides_backup.zip"` {
		t.Fatalf("Content-Disposition = %q", disposition)
	}
	if rec.Body.String() != "backup" {
		t.Fatalf("body = %q, want backup", rec.Body.String())
	}
	if !cleaned {
		t.Fatal("cleanup was not called after serving download")
	}
}

type badRequestError struct{}

func (badRequestError) Error() string       { return "annotation ann_io6n1b cannot be represented as TEI" }
func (badRequestError) HTTPStatusCode() int { return http.StatusBadRequest }

func TestGetXMLUsesStatusCodeProvidedByError(t *testing.T) {
	handler := GetXML(func(r *http.Request) ([]byte, error) {
		return nil, badRequestError{}
	}).Build()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if body := rec.Body.String(); body != "annotation ann_io6n1b cannot be represented as TEI\n" {
		t.Fatalf("body = %q, want annotation error", body)
	}
}
