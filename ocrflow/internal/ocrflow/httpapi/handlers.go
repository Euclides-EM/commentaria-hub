package httpapi

import (
	_ "github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/docs" // swagger docs
)

type Handlers struct {
	deps *Dependencies
}

func NewHandlers(deps *Dependencies) *Handlers {
	return &Handlers{deps: deps}
}
