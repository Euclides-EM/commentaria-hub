package featureplat

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"strings"
)

//go:embed paratext_transcriptions.csv
var paratextTranscriptionsCSV string

type TPSTranscriptions struct {
}

func NewTPSTranscriptions() *TPSTranscriptions {
	return &TPSTranscriptions{}
}

func (t *TPSTranscriptions) Get(key string) (string, string, error) {
	// todo: it's a hack... but I'll leave with it for now.
	reader := csv.NewReader(strings.NewReader(paratextTranscriptionsCSV))
	records, err := reader.ReadAll()
	if err != nil {
		return "", "", err
	}

	// Find header row
	if len(records) == 0 {
		return "", "", fmt.Errorf("CSV file is empty")
	}

	headers := records[0]
	keyIdx := -1
	titleIdx := -1
	imprintIdx := -1

	for i, h := range headers {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "key":
			keyIdx = i
		case "title":
			titleIdx = i
		case "imprint":
			imprintIdx = i
		}
	}

	if keyIdx == -1 || titleIdx == -1 || imprintIdx == -1 {
		return "", "", fmt.Errorf("CSV missing required columns: key, title, imprint")
	}

	// Find the row with matching key
	for _, record := range records[1:] {
		if len(record) > keyIdx && strings.TrimSpace(record[keyIdx]) == key {
			title := ""
			imprint := ""
			if len(record) > titleIdx {
				title = record[titleIdx]
			}
			if len(record) > imprintIdx {
				imprint = record[imprintIdx]
			}
			return title, imprint, nil
		}
	}

	return "", "", fmt.Errorf("key '%s' not found in CSV", key)
}
