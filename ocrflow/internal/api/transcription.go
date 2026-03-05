package api

import (
	"net/http"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
)

// ListEditionTranscriptionsDetails godoc
// @Summary      List Transcriptions
// @Description  Retrieve a list of transcriptions, optionally filtered by edition or facsimile.
// @Tags         Transcriptions
// @Accept       json
// @Produce      json
// @Param        edition_id  query     []string  false  "Filter by edition ID"  collectionFormat(multi)
// @Security 	 BearerAuth
// @Success      200    {array}   model.EditionTranscription  "List of transcriptions"
// @Router       /editions/transcriptions [get]
func (h *Handlers) ListEditionTranscriptionsDetails(r *http.Request) (any, error) {
	editionIDs := r.URL.Query()["edition_id"]
	return h.deps.EditionTranscriptionSvc.ListTranscriptionsByEditionIDs(editionIDs)
}

// UpdateEditionTranscriptionsDetails godoc
// @Summary	  Update Preferred Transcription
// @Description  Update the preferred transcription for a specific edition.
// @Tags         Transcriptions
// @Accept       json
// @Produce      json
// @Param        edition_id  path      string  true  "Edition ID"
// @Param        body		body      model.EditionTranscription  true  "Preferred transcription details"
// @Security 	 BearerAuth
// @Success      200    {object}  model.EditionTranscription  "Updated preferred transcription"
// @Router       /editions/{edition_id}/transcriptions [put]
func (h *Handlers) UpdateEditionTranscriptionsDetails(r *http.Request) (any, error) {
	editionID := r.Context().Value("edition_id").(string)
	var req model.EditionTranscription
	if err := DecodeBody(r, &req); err != nil {
		return nil, err
	}
	return h.deps.EditionTranscriptionSvc.Update(editionID, &req)
}
