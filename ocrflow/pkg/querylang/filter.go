package querylang

type Filter struct {
}

func ParseFilter(filterStr string) (*Filter, error) {
	return &Filter{}, nil
}
