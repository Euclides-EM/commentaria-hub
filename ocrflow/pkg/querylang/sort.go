package querylang

type Sort []string

func ParseSort(sortStr string) (Sort, error) {
	// Placeholder implementation
	if sortStr == "" {
		return []string{}, nil
	}
	return []string{sortStr}, nil
}
