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
	Status        string  `json:"status,omitempty"`         // "creating" | "ready" | "failed"
	CreationError string  `json:"creation_error,omitempty"` // set when status is "failed"
}
