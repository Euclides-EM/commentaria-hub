package api

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/job"
)

// ListFacsimiles godoc
// @Summary      List Facsimiles (bulk get)
// @Description  Get facsimiles, optionally filtered by edition ID.
// @Tags         Facsimiles
// @Param        edition_id  query     []string  false  "Filter by edition ID"  collectionFormat(multi)
// @Produce      json
// @Success      200  {array}  model.Facsimile
// @Router       /facsimilies [get]
func (h *Handlers) ListFacsimiles(r *http.Request) (any, error) {
	editionIDs := r.URL.Query()["edition_id"]
	return h.deps.FacsimileSvc.ListFacsimiles(editionIDs)
}

// CreateFacsimile godoc
// @Summary      Create Facsimile
// @Description  Create a new facsimile
// @Tags         Facsimiles
// @Accept       json
// @Produce      json
// @Param        facsimile  body      model.Facsimile  true  "Facsimile to create"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Facsimile
// @Router       /facsimilies [post]
func (h *Handlers) CreateFacsimile(r *http.Request) (any, error) {
	var facsimile model.Facsimile
	if err := DecodeBody(r, &facsimile); err != nil {
		return nil, err
	}
	return h.deps.FacsimileSvc.CreateFacsimile(&facsimile)
}

// GetFacsimile godoc
// @Summary      Get Facsimile by ID
// @Description  Get a single facsimile by its ID.
// @Tags         Facsimiles
// @Param        id  path      string  true  "Facsimile ID"
// @Produce      json
// @Success      200  {object}  model.Facsimile
// @Failure      404  "Facsimile not found"
// @Router       /facsimilies/{id} [get]
func (h *Handlers) GetFacsimile(r *http.Request) (any, error) {
	id := r.PathValue("id")
	if id == "" {
		return nil, fmt.Errorf("missing facsimile ID")
	}
	fac, err := h.deps.FacsimileSvc.GetFacsimile(id)
	if err != nil {
		return nil, err
	}
	return fac, nil
}

// UpdateFacsimile godoc
// @Summary      Update Facsimile
// @Description  Update an existing facsimile identified by ID.
// @Tags         Facsimiles
// @Accept       json
// @Produce      json
// @Param        id          path      string  true  "Facsimile ID"
// @Param        facsimile   body      model.Facsimile  true  "Facsimile data to update"
// @Security 	 BearerAuth
// @Success      200  {object}  model.Facsimile
// @Router       /facsimilies/{id} [put]
func (h *Handlers) UpdateFacsimile(r *http.Request) (any, error) {
	id := r.PathValue("id")
	if id == "" {
		return nil, fmt.Errorf("missing facsimile ID")
	}
	var facsimile model.Facsimile
	if err := DecodeBody(r, &facsimile); err != nil {
		return nil, err
	}
	if facsimile.ID != "" && facsimile.ID != id {
		return nil, fmt.Errorf("facsimile id in body (%s) does not match id in path (%s)", facsimile.ID, id)
	}
	facsimile.ID = id
	return h.deps.FacsimileSvc.UpdateFacsimile(&facsimile)
}

// ImportFacsimilesFromDrive godoc
// @Summary      Import facsimiles and diagram crops from Google Drive inbox
// @Description  Copies PDFs from the configured Google Drive folder into FACSIMILES_PDF_DIR, installs diagram crop archives into FACSIMILES_DIAGRAMS_PATH, updates metadata, then deletes successfully imported files from Drive.
// @Tags         Facsimiles
// @Produce      json
// @Security 	 BearerAuth
// @Param        async  query     bool  false  "Create a background import job instead of waiting for completion"
// @Success      200  {object}  model.FacsimileDriveImportResult
// @Success      200  {object}  job.Job
// @Router       /facsimilies/import-from-drive [post]
func (h *Handlers) ImportFacsimilesFromDrive(r *http.Request) (any, error) {
	async, err := strconv.ParseBool(r.URL.Query().Get("async"))
	if err == nil && async {
		return h.deps.JobSvc.CreateJob(&job.Job{
			Task: job.FacsimileDriveImport,
			Meta: common.NewMeta("").WithName("Import facsimiles from Drive").WithDescription("Import PDFs and diagram crops from the configured Google Drive inbox"),
		})
	}
	return h.deps.FacsimileSvc.ImportFromDriveInbox()
}

// DownloadFacsimilePDF godoc
// @Summary      Download facsimile PDF
// @Description  Downloads a local facsimile PDF by facsimile ID.
// @Tags         Facsimiles
// @Produce      application/pdf
// @Param        id  path      string  true  "Facsimile ID"
// @Security 	 BearerAuth
// @Success      200  {file}  string  "Facsimile PDF"
// @Router       /facsimilies/{id}/pdf [get]
func (h *Handlers) DownloadFacsimilePDF(r *http.Request) (filePath string, downloadName string, err error) {
	id := r.PathValue("id")
	if id == "" {
		return "", "", fmt.Errorf("missing facsimile ID")
	}
	filePath, err = h.deps.FacsimileSvc.GetFacsimilePDFPath(id)
	if err != nil {
		return "", "", err
	}
	return filePath, filepath.Base(filePath), nil
}

// DownloadEditionFacsimilePDF godoc
// @Summary      Download edition facsimile PDF
// @Description  Downloads the first local facsimile PDF for an edition.
// @Tags         Facsimiles
// @Produce      application/pdf
// @Param        editionId  path      string  true  "Edition ID"
// @Security 	 BearerAuth
// @Success      200  {file}  string  "Facsimile PDF"
// @Router       /editions/{editionId}/facsimile.pdf [get]
func (h *Handlers) DownloadEditionFacsimilePDF(r *http.Request) (filePath string, downloadName string, err error) {
	editionID, err := extractEditionId(r)
	if err != nil {
		return "", "", err
	}
	filePath, err = h.deps.FacsimileSvc.GetEditionFacsimilePDFPath(editionID)
	if err != nil {
		return "", "", err
	}
	return filePath, filepath.Base(filePath), nil
}
