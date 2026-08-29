package model

import (
	"encoding/xml"
	"strings"
	"testing"
)

func TestSurfaceMarshalPreservesZeroOrigin(t *testing.T) {
	data, err := xml.Marshal(Surface{LRX: 11075, LRY: 15975})
	if err != nil {
		t.Fatalf("marshal surface: %v", err)
	}

	got := string(data)
	for _, attr := range []string{`ulx="0"`, `uly="0"`, `lrx="11075"`, `lry="15975"`} {
		if !strings.Contains(got, attr) {
			t.Fatalf("marshaled surface %q does not contain %s", got, attr)
		}
	}
}
