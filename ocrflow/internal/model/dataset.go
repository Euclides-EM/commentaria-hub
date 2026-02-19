package model

import "github.com/MiaMish/elements-dh/ocrflow/internal/model/common"

// Dataset creation status (async flow).
const (
	DatasetStatusCreating = "creating"
	DatasetStatusReady    = "ready"
	DatasetStatusFailed   = "failed"
)

type Dataset struct {
	common.Meta
	FacsimileID   string  `json:"facsimile_id"`
	EditionID     string  `json:"edition_id"`
	DPI           float64 `json:"dpi" default:"300"`
	Deskewed      bool    `json:"deskewed" default:"false"`
	Pages         string  `json:"pages,omitempty"`          // optional page spec e.g. "1,3-5,10"; parsed with pagesparser.Parse
	Status        string  `json:"status,omitempty"`         // "creating" | "ready" | "failed"
	CreationError string  `json:"creation_error,omitempty"` // set when status is "failed"
}
