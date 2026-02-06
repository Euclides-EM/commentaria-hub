package featureplat

type Handlers struct {
	deps *Dependencies
}

func NewFeatureAppHandlers(deps *Dependencies) *Handlers {
	return &Handlers{deps: deps}
}
