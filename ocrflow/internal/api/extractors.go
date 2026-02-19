package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func extractDatasetAndAnnotationIDs(r *http.Request) (string, string, error) {
	datasetID, err := extractDatasetID(r)
	if err != nil {
		return "", "", err
	}
	annotationID := r.PathValue("id")
	if annotationID == "" {
		return "", "", fmt.Errorf("missing annotation ID")
	}
	return datasetID, annotationID, nil
}

func extractDatasetID(r *http.Request) (string, error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return "", fmt.Errorf("missing dataset ID")
	}
	return datasetID, nil
}

func extractFeatureID(r *http.Request) (string, string, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return "", "", err
	}
	featureId := r.PathValue("featureId")
	if featureId == "" {
		return "", "", fmt.Errorf("missing feature ID")
	}
	return dataSetId, featureId, nil
}

func extractFeatureRevisionID(r *http.Request) (string, string, string, error) {
	dataSetId, featureId, err := extractFeatureID(r)
	if err != nil {
		return "", "", "", err
	}
	revisionId := r.PathValue("revisionId")
	if revisionId == "" {
		return "", "", "", fmt.Errorf("missing revision ID")
	}
	return dataSetId, featureId, revisionId, nil
}

func extractExecutionID(r *http.Request) (string, error) {
	executionId := r.PathValue("executionId")
	if executionId == "" {
		return "", fmt.Errorf("missing execution ID")
	}
	return executionId, nil
}

func extractEditionId(r *http.Request) (string, error) {
	editionId := r.PathValue("editionId")
	if editionId == "" {
		return "", fmt.Errorf("missing editionId")
	}
	return editionId, nil
}

func extractJobID(r *http.Request) (string, error) {
	jobId := r.PathValue("jobId")
	if jobId == "" {
		return "", fmt.Errorf("missing job ID")
	}
	return jobId, nil
}

func DecodeBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("failed to decode request body to type %T: %w", dst, err)
	}
	return nil
}
