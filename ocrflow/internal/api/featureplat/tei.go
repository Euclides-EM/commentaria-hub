package featureplat

import "net/http"

// GetTEI godoc
// @Summary      Get TEI data for a collection
// @Description  Retrieve TEI data for a specific collection, optionally filtered by key and features
// @Tags         TEI
// @Param        id          path      string  true  "Collection ID"
// @Param        key         query     string  false "Key to filter TEI data"
// @Param        features    query     string  false "Comma-separated list of features to filter TEI data"
// @Produce      xml
// @Success      200  {string}  string "TEI XML data"
// @Router       /collections/{id}/tei [get]
func (h *Handlers) GetTEI(r *http.Request) (any, error) {
	collectionId, err := extractCollectionID(r)
	if err != nil {
		return nil, err
	}
	key := r.URL.Query().Get("key")
	features := r.URL.Query().Get("features")
	return h.deps.TEISvc.GetTEI(collectionId, key, features)
}
