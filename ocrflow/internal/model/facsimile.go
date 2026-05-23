package model

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

type Facsimile struct {
	common.Meta
	EditionID             string `json:"edition_id"`
	ScanURL               string `json:"scan_url"`
	MainTextPages         string `json:"main_text_pages"`
	DiagramCropsAvailable bool   `json:"diagram_crops_available"`
	DownloadAvailable     bool   `json:"download_available" readonly:"true"`
}

type FacsimileDriveImportResult struct {
	ImportedPDFs            []string `json:"importedPdfs"`
	ImportedDiagramArchives []string `json:"importedDiagramArchives"`
	ImportedDiagramCrops    []string `json:"importedDiagramCrops"`
	Skipped                 []string `json:"skipped"`
	Deleted                 []string `json:"deleted"`
}
