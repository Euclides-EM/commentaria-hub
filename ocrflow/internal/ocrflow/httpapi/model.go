package httpapi

import (
	"encoding/json"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"net/http"
)

// ListModels godoc
// @Summary      List Models
// @Description  Get a list of available models.
// @Tags         Models
// @Produce      json
// @Success      200  {array}   model.Model
// @Router       /models [get]
func (h *Handlers) ListModels(r *http.Request) (any, error) {
	return h.deps.ModelSvc.List()
}

// TrainModel godoc
// @Summary      Train Model
// @Description  Train a new model based on the provided configuration.
// @Tags         Train
// @Accept       json
// @Produce      json
// @Param        model  body      model.Training  true  "Training Configuration"
// @Success      200    {object}  model.Training
// @Router       /train [post]
func (h *Handlers) TrainModel(r *http.Request) (any, error) {
	decoder := json.NewDecoder(r.Body)
	var m model.Training
	if err := decoder.Decode(&m); err != nil {
		return nil, fmt.Errorf("failed to decode request body: %w", err)
	}
	return h.deps.TrainSvc.TrainYolo(&m)
}
