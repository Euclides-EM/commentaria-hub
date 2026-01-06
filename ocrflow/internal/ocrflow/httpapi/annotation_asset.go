package httpapi

import (
	"fmt"
	"net/http"
)

// DownloadAnnotationAssets godoc
// @Summary Download annotation assets
// @Description Generate and download assets for a specific annotation within a dataset
// @Tags Annotations
// @Param dataSetId path string true "Dataset ID"
// @Param id path string true "Annotation ID"
// @Produce application/zip
// @Success 200 {file} file "ZIP file containing the annotation assets"
// @Router /datasets/{dataSetId}/annotations/{id}/download_assets [get]
func (h *Handlers) DownloadAnnotationAssets(r *http.Request) (zipPath string, deleteAfterServe bool, err error) {
	datasetID := r.PathValue("dataSetId")
	if datasetID == "" {
		return "", false, fmt.Errorf("missing dataset ID")
	}
	annotationID := r.PathValue("id")
	if annotationID == "" {
		return "", false, fmt.Errorf("missing annotation ID")
	}

	zipPath, err = h.deps.AssetGen.GenerateAssets(datasetID, annotationID, nil)
	if err != nil {
		return "", false, err
	}

	return zipPath, true, nil
}
