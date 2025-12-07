package alto

import (
	"encoding/xml"
	"fmt"
	"os"
)

func LoadFromFile(path string) (*Alto, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ALTO: %w", err)
	}
	var a Alto
	if err := xml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("unmarshal ALTO: %w", err)
	}
	if a.Xmlns == "" {
		a.Xmlns = "http://www.loc.gov/standards/alto/ns-v4#"
	}
	if a.XmlnsXsi == "" {
		a.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	}
	if a.SchemaLocation == "" {
		a.SchemaLocation = "http://www.loc.gov/standards/alto/ns-v4# http://www.loc.gov/standards/alto/v4/alto-4-3.xsd"
	}
	return &a, nil
}

func SaveToFile(af *Alto, path string) error {
	data, err := xml.MarshalIndent(af, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ALTO: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write ALTO: %w", err)
	}
	return nil
}
