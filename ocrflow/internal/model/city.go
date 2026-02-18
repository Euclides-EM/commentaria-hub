package model

type City struct {
	Name      string  `json:"name" validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
	Latitude  float64 `json:"latitude" validate:"required"`
}
