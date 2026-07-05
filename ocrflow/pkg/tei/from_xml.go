package tei

import (
	"encoding/xml"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func ParseTEIFromXML(xmlData []byte) (*model.TEI, error) {
	var tei model.TEI
	err := xml.Unmarshal(xmlData, &tei)
	if err != nil {
		return nil, err
	}

	return &tei, nil
}
