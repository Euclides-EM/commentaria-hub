package model

type FacsimileDriveImportResult struct {
	Imported []string `json:"imported"`
	Skipped  []string `json:"skipped"`
	Deleted  []string `json:"deleted"`
}
