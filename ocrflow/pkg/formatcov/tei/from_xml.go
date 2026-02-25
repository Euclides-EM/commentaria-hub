package tei

import (
	"encoding/xml"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

func ParseTEIFromXML(xmlData []byte) (*model.TEI, error) {
	var tei model.TEI
	err := xml.Unmarshal(xmlData, &tei)
	if err != nil {
		return nil, err
	}

	return &tei, nil
}
