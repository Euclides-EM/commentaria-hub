package rbmodel

type DetectionFile struct {
	InferenceID string  `json:"inference_id"`
	Time        float64 `json:"time"`
	Image       struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"image"`
	Predictions []Prediction `json:"predictions"`
}

type Prediction struct {
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
	Width       float64 `json:"width"`
	Height      float64 `json:"height"`
	Confidence  float64 `json:"confidence"`
	Class       string  `json:"class"`
	ClassID     int     `json:"class_id"`
	DetectionID string  `json:"detection_id"`
}
