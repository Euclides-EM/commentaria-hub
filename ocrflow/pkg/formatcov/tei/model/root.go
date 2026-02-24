package model

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type TEI struct {
	XMLName   xml.Name  `xml:"TEI"`
	Xmlns     string    `xml:"xmlns,attr,omitempty"`
	Header    Header    `xml:"teiHeader"`
	Facsimile Facsimile `xml:"facsimile,omitempty"`
	Text      Text      `xml:"text"`
}

func (t *TEI) ToXML() ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("TEI is nil")
	}

	// Ensure namespace is set
	if t.Xmlns == "" {
		t.Xmlns = "http://www.tei-c.org/ns/1.0"
	}

	var buf bytes.Buffer

	// XML declaration
	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(&buf)
	defer enc.Close()

	enc.Indent("", "  ")

	if err := enc.Encode(t); err != nil {
		return nil, err
	}

	if err := enc.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
