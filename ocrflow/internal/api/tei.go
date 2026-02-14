package api

import (
	"net/http"
)

// GetTEI godoc
// @Summary      Get TEI data for a dataset
// @Description  Retrieve TEI data for a specific dataset, optionally filtered by key and features
// @Tags         TEI
// @Param        dataSetId   path      string  true  "Dataset ID"
// @Param        key         query     string  false "Key to filter TEI data"
// @Param        features    query     string  false "Comma-separated list of features to filter TEI data"
// @Produce      xml
// @Success      200  {string}  string "TEI XML data"
// @Router        /datasets/{dataSetId}/tei [get]
func (h *Handlers) GetTEI(r *http.Request) ([]byte, error) {
	dataSetId, err := extractDatasetID(r)
	if err != nil {
		return nil, err
	}
	key := r.URL.Query().Get("key")
	featureParams := r.URL.Query()["features"]
	return h.deps.TEISvc.GetTEI(dataSetId, key, featureParams)
}
