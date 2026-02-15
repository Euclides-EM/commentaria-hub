package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// TrainModel godoc
// @Summary      Train Model
// @Description  Train a new model based on the provided configuration.
// @Tags         Train
// @Accept       json
// @Produce      json
// @Param        model  body      model.Training  true  "Training Configuration"
// @Security 	 BearerAuth
// @Success      200    {object}  model.Training
// @Router       /train [post]
func (h *Handlers) TrainModel(r *http.Request) (any, error) {
	var m model.Training
	if err := DecodeBody(r, &m); err != nil {
		return nil, err
	}
	return h.deps.TrainSvc.TrainYolo(&m)
}
