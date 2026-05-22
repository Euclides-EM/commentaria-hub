package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
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

func extractDatasetFeatureID(r *http.Request) (string, string, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return "", "", err
	}
	featureI, err := extractFeatureID(r)
	if err != nil {
		return "", "", err
	}
	return dataSetId, featureI, nil
}

func extractFeatureID(r *http.Request) (string, error) {
	featureId := r.PathValue("featureId")
	if featureId == "" {
		return "", fmt.Errorf("missing feature ID")
	}
	return featureId, nil
}

func extractRevisionID(r *http.Request) (string, error) {
	revisionId := r.PathValue("revisionId")
	if revisionId == "" {
		return "", fmt.Errorf("missing revision ID")
	}
	return revisionId, nil
}

func extractDatasetFeatureRevisionID(r *http.Request) (string, string, string, error) {
	dataSetId, featureId, err := extractDatasetFeatureID(r)
	if err != nil {
		return "", "", "", err
	}
	revisionId, err := extractRevisionID(r)
	if err != nil {
		return "", "", "", err
	}
	return dataSetId, featureId, revisionId, nil
}

func extractFeatureRevisionID(r *http.Request) (string, string, error) {
	featureId, err := extractFeatureID(r)
	if err != nil {
		return "", "", err
	}
	revisionId, err := extractRevisionID(r)
	if err != nil {
		return "", "", err
	}
	return featureId, revisionId, nil
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

func extractGroupId(request *http.Request) (string, error) {
	groupId := request.PathValue("groupId")
	if groupId == "" {
		return "", fmt.Errorf("missing group ID")
	}
	return groupId, nil
}

func extractDefScope(r *http.Request) (feature.DefScope, error) {
	scopeType := feature.ScopeType(r.URL.Query().Get("scope"))
	if scopeType == "" {
		return feature.DefScope{}, fmt.Errorf("missing scope type")
	}
	if scopeType == feature.ScopeTypeEditions {
		return feature.NewEditionDefScope(), nil
	}
	if scopeType == feature.ScopeTypeDataset {
		datasetId := r.URL.Query().Get("dataset")
		return feature.NewDatasetDefScope(datasetId), nil
	}
	return feature.DefScope{}, fmt.Errorf("invalid scope")
}

func extractExecScope(r *http.Request) (feature.ExecScope, error) {
	scopeType := feature.ScopeType(r.URL.Query().Get("scope"))
	switch scopeType {
	case feature.ScopeTypeEditions:
		return feature.NewEditionExecScope(), nil
	case feature.ScopeTypeDataset:
		datasetID := r.URL.Query().Get("dataset")
		annotationID := r.URL.Query().Get("annotation")
		if datasetID == "" || annotationID == "" {
			return feature.ExecScope{}, fmt.Errorf("dataset and annotation are required for dataset feature results")
		}
		return feature.NewDatasetExecScope(datasetID, annotationID), nil
	case "":
		return feature.ExecScope{}, fmt.Errorf("missing scope type")
	default:
		return feature.ExecScope{}, fmt.Errorf("invalid scope")
	}
}

func DecodeBody(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("failed to decode request body to type %T: %w", dst, err)
	}
	return nil
}
