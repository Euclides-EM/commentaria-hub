package testutils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

// EqualXML compares two XML documents semantically.
// It ignores formatting differences like indentation and extra whitespace.
func EqualXML(a, b []byte) error {
	na, err := normalizeXML(a)
	if err != nil {
		return fmt.Errorf("first XML invalid: %w", err)
	}

	nb, err := normalizeXML(b)
	if err != nil {
		return fmt.Errorf("second XML invalid: %w", err)
	}

	if !bytes.Equal(na, nb) {
		return fmt.Errorf("XML not equal\n--- A ---\n%s\n--- B ---\n%s", na, nb)
	}

	return nil
}

func normalizeXML(input []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(input))
	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch t := tok.(type) {
		case xml.StartElement:
			// Sort attributes for stable output
			// (optional but helps with equality)
			sortAttrs(&t)
			if err := encoder.EncodeToken(t); err != nil {
				return nil, err
			}
		default:
			if err := encoder.EncodeToken(tok); err != nil {
				return nil, err
			}
		}
	}

	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func sortAttrs(se *xml.StartElement) {
	if len(se.Attr) <= 1 {
		return
	}
	sort.Slice(se.Attr, func(i, j int) bool {
		if se.Attr[i].Name.Space == se.Attr[j].Name.Space {
			return se.Attr[i].Name.Local < se.Attr[j].Name.Local
		}
		return se.Attr[i].Name.Space < se.Attr[j].Name.Space
	})
}
