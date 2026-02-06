package ocrflow

import (
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
