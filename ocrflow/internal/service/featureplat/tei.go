package featureplat

type TEI struct {
	s map[string]string
}

func NewTEI() *TEI {
	return &TEI{
		s: make(map[string]string),
	}
}

func (t *TEI) GetTEI(collectionId, key, features string) (string, error) {
	if tei, ok := t.s[key]; ok {
		return tei, nil
	}
	return "", nil
}
