package httpwrapper

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
