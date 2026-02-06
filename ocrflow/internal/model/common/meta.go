package common

import "time"

type Meta struct {
	ID          string    `json:"id" readonly:"true"`
	CreatedAt   time.Time `json:"created_at,omitempty" readonly:"true"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" readonly:"true"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
}

func NewMeta(id string) Meta {
	return Meta{
		ID: id,
	}
}

func (m Meta) WithName(name string) Meta {
	m.Name = name
	return m
}

func (m Meta) WithDescription(description string) Meta {
	m.Description = description
	return m
}
