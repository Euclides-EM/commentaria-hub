package model

import "time"

type Meta struct {
	ID        string    `json:"id" readonly:"true"`
	CreatedAt time.Time `json:"created_at,omitempty" readonly:"true"`
	UpdatedAt time.Time `json:"updated_at,omitempty" readonly:"true"`
}

func NewMeta(id string) Meta {
	return Meta{
		ID: id,
	}
}
