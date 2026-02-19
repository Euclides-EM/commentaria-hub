package api

import "net/http"

// ListCities godoc
// @Summary      List cities
// @Description  Returns cities metadata with city name, longitude, and latitude.
// @Tags         Geo Data
// @Produce      json
// @Success      200  {array}  model.City
// @Router       /cities [get]
func (h *Handlers) ListCities(_ *http.Request) (any, error) {
	return h.deps.GeoSvc.ListCities()
}
