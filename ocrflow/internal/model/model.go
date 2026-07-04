package model

import (
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

type OCRModelLocation string

const (
	OCRModelLocationLocal    OCRModelLocation = "local"
	OCRModelLocationRoboflow OCRModelLocation = "roboflow"
)

type OCRModelAlgorithmFamily string

const (
	OCRModelAlgorithmFamilyYOLO OCRModelAlgorithmFamily = "yolo"
)

type Model struct {
	common.Meta     `json:",inline"`
	Type            common.OCRModelType     `json:"type"`
	Location        OCRModelLocation        `json:"location"`
	AlgorithmFamily OCRModelAlgorithmFamily `json:"algorithm_family,omitempty"`
	// LocalPath is the path to the model file on the local filesystem. It is relevant only for local models.
	LocalPath         string                  `json:"local_path" readonly:"true"`
	Categories        []string                `json:"categories,omitempty"`
	BaseModelID       string                  `json:"base_model_id,omitempty"`
	BaseAnnotations   []*annotation.Reference `json:"base_annotations,omitempty"`
	UsedInAnnotations []*annotation.Reference `json:"used_in_annotations,omitempty"`
}

type ModelTrainingStatus string

const ModelTrainingStatusSubmitted = "submitted"

type ModelTraining struct {
	common.Meta   `json:",inline"`
	Model         *Model              `json:"model"`
	Epochs        int                 `json:"epochs,omitempty"`
	Status        ModelTrainingStatus `json:"status"`
	StatusDetails map[string]string   `json:"status_details"`
	Backend       string              `json:"backend"`
	GPUFarmHost   string              `json:"gpu_farm_host"`
	RemoteRunDir  string              `json:"remote_run_dir"`
}
