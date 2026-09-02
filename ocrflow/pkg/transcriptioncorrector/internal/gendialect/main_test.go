package main

import (
	"strings"
	"testing"
)

func TestExtractLLMContract(t *testing.T) {
	document := []byte("intro\n" + contractBeginMarker + "\ncontract\n" + contractEndMarker + "\ndetails\n")

	contract, err := extractLLMContract(document)
	if err != nil {
		t.Fatal(err)
	}
	if string(contract) != "contract\n" {
		t.Fatalf("contract = %q, want %q", contract, "contract\\n")
	}
}

func TestExtractLLMContractRejectsMissingOrDuplicateMarkers(t *testing.T) {
	for name, document := range map[string]string{
		"missing":   "no markers",
		"duplicate": contractBeginMarker + contractBeginMarker + contractEndMarker,
		"reversed":  contractEndMarker + contractBeginMarker,
		"empty":     contractBeginMarker + contractEndMarker,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := extractLLMContract([]byte(document))
			if err == nil || !strings.Contains(err.Error(), "marker") && name != "empty" {
				t.Fatalf("error = %v, want marker validation error", err)
			}
		})
	}
}
