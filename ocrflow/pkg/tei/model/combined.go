package model

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type CombinedTEI struct {
	XMLName      xml.Name          `xml:"batchTeiResponse"`
	Xmlns        string            `xml:"xmlns,attr,omitempty"`
	DatasetID    string            `xml:"datasetId,attr,omitempty"`
	AnnotationID string            `xml:"annotationId,attr,omitempty"`
	Items        []CombinedTEIItem `xml:"item"`
}

type CombinedTEIItem struct {
	Key     string      `xml:"key,attr"`
	TEI     *TEI        `xml:"TEI,omitempty"`
	Missing *MissingTEI `xml:"missing,omitempty"`
	Error   *TEIError   `xml:"error,omitempty"`
}

type MissingTEI struct {
	Reason string `xml:"reason,attr,omitempty"`
}

type TEIError struct {
	Code    string `xml:"code,attr,omitempty"`
	Message string `xml:",chardata"`
}

func (c *CombinedTEI) ToXML() ([]byte, error) {
	if c == nil {
		return nil, fmt.Errorf("CombinedTEI is nil")
	}

	if c.Xmlns == "" {
		c.Xmlns = "http://example.com/api/tei-batch"
	}

	for i := range c.Items {
		if c.Items[i].TEI != nil && c.Items[i].TEI.Xmlns == "" {
			c.Items[i].TEI.Xmlns = "http://www.tei-c.org/ns/1.0"
		}
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(&buf)
	defer enc.Close()

	enc.Indent("", "  ")

	if err := enc.Encode(c); err != nil {
		return nil, err
	}

	if err := enc.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
