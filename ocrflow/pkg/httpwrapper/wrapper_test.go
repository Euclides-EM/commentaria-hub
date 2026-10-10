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
