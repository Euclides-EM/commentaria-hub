package featureplat

import (
	"fmt"
	"net/http"
)

func extractCollectionID(r *http.Request) (string, error) {
	collectionId := r.PathValue("collectionId")
	if collectionId == "" {
		return "", fmt.Errorf("missing collection ID")
	}
	return collectionId, nil
}

func extractFeatureID(r *http.Request) (string, string, error) {
	collectionId, err := extractCollectionID(r)
	if err != nil {
		return "", "", err
	}
	featureId := r.PathValue("featureId")
	if featureId == "" {
		return "", "", fmt.Errorf("missing feature ID")
	}
	return collectionId, featureId, nil
}

func extractFeatureRevisionID(r *http.Request) (string, string, string, error) {
	collectionId, featureId, err := extractFeatureID(r)
	if err != nil {
		return "", "", "", err
	}
	revisionId := r.PathValue("revisionId")
	if revisionId == "" {
		return "", "", "", fmt.Errorf("missing revision ID")
	}
	return collectionId, featureId, revisionId, nil
}

func extractExecutionID(r *http.Request) (string, error) {
	executionId := r.PathValue("executionId")
	if executionId == "" {
		return "", fmt.Errorf("missing execution ID")
	}
	return executionId, nil
}
