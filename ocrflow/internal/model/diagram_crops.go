package model

type DiagramCrops struct {
	Key             string               `json:"key,omitempty"`
	ImageURLsByName map[string]string    `json:"imageURLsByName"`
	HasDiagrams     bool                 `json:"hasDiagrams"`
	Volumes         []*DiagramCropVolume `json:"volumes,omitempty"`
}

type DiagramCropVolume struct {
	Volume          int               `json:"volume,omitempty"`
	Key             string            `json:"key,omitempty"`
	ImageURLsByName map[string]string `json:"imageUrlsByName"`
	HasDiagrams     bool              `json:"hasDiagrams"`
}
