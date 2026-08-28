package model

import (
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type FacsimileConnectionConfirmationStatus string

const (
	FacsimileConnectionStatusGuessedByMachine FacsimileConnectionConfirmationStatus = "guessed_by_machine"
	FacsimileConnectionStatusGuessedByHuman   FacsimileConnectionConfirmationStatus = "guessed_by_human"
	FacsimileConnectionStatusHumanConfirmed   FacsimileConnectionConfirmationStatus = "human_confirmed"
)

type Facsimile struct {
	common.Meta
	EditionID                             string                                `json:"edition_id"`
	ShelfmarkID                           string                                `json:"shelfmark_id,omitempty"`
	ScanURL                               string                                `json:"scan_url"`
	MainTextPages                         string                                `json:"main_text_pages"`
	DiagramCropsAvailable                 bool                                  `json:"diagram_crops_available"`
	DownloadAvailable                     bool                                  `json:"download_available" readonly:"true"`
	FileSizeBytes                         int64                                 `json:"file_size_bytes,omitempty"`
	ImportedAt                            *time.Time                            `json:"imported_at,omitempty"`
	FacsimileConnectionConfirmationStatus FacsimileConnectionConfirmationStatus `json:"facsimile_connection_confirmation_status,omitempty"`
}

type FacsimileDriveImportResult struct {
	ImportedPDFs            []string `json:"importedPdfs"`
	ImportedDiagramArchives []string `json:"importedDiagramArchives"`
	ImportedDiagramCrops    []string `json:"importedDiagramCrops"`
	Skipped                 []string `json:"skipped"`
	Deleted                 []string `json:"deleted"`
}
