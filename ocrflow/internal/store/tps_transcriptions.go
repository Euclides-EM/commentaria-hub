package store

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"strings"
)

type TPSTranscriptions struct {
}

func NewTPSTranscriptions() *TPSTranscriptions {
	return &TPSTranscriptions{}
}

func (t *TPSTranscriptions) Keys() ([]string, error) {
	reader := csv.NewReader(strings.NewReader(paratextTranscriptionsCSV))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	keyIdx := -1

	for i, h := range headers {
		if strings.ToLower(strings.TrimSpace(h)) == "key" {
			keyIdx = i
			break
		}
	}

	if keyIdx == -1 {
		return nil, fmt.Errorf("CSV missing required column: key")
	}

	var keys []string
	for _, record := range records[1:] {
		if len(record) > keyIdx {
			keys = append(keys, strings.TrimSpace(record[keyIdx]))
		}
	}

	return keys, nil
}
