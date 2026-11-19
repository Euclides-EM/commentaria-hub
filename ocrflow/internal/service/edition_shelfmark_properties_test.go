package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
)

func TestEditionHasAnyShelfmarkProperty(t *testing.T) {
	tests := []struct {
		name       string
		shelfmarks []model.EditionShelfmark
		allowed    []string
		want       bool
	}{
		{
			name:       "shelfmark available",
			shelfmarks: []model.EditionShelfmark{{Shelfmark: "  Cambridge, CUL Adv.b.1.2  "}},
			allowed:    []string{shelfmarkAvailable},
			want:       true,
		},
		{
			name:       "facsimile available",
			shelfmarks: []model.EditionShelfmark{{Scan: "https://example.org/scan"}},
			allowed:    []string{facsimileAvailable},
			want:       true,
		},
		{
			name:       "known copyright on one shelfmark means status is not unknown",
			shelfmarks: []model.EditionShelfmark{{Copyright: "Public domain"}, {Copyright: "  "}},
			allowed:    []string{copyrightStatusUnknown},
			want:       false,
		},
		{
			name:       "copyright unknown on all shelfmarks",
			shelfmarks: []model.EditionShelfmark{{Copyright: ""}, {Copyright: "  "}},
			allowed:    []string{copyrightStatusUnknown},
			want:       true,
		},
		{
			name: "external transcription available",
			shelfmarks: []model.EditionShelfmark{{
				TranscriptionAvailable: model.EditionShelfmarkTranscriptionExternal,
			}},
			allowed: []string{externalTranscriptionAvailable},
			want:    true,
		},
		{
			name: "internal transcription available",
			shelfmarks: []model.EditionShelfmark{{
				TranscriptionAvailable: model.EditionShelfmarkTranscriptionInternal,
			}},
			allowed: []string{internalTranscriptionAvailable},
			want:    true,
		},
		{
			name: "external structural metadata available",
			shelfmarks: []model.EditionShelfmark{{
				StructuralMetadataAvailable: model.EditionShelfmarkStructuralMetadataAvailabilityExternal,
			}},
			allowed: []string{externalStructuralMetadata},
			want:    true,
		},
		{
			name: "internal structural metadata available",
			shelfmarks: []model.EditionShelfmark{{
				StructuralMetadataAvailable: model.EditionShelfmarkStructuralMetadataAvailabilityInternal,
			}},
			allowed: []string{internalStructuralMetadata},
			want:    true,
		},
		{
			name:       "selected properties use OR matching",
			shelfmarks: []model.EditionShelfmark{{Scan: "https://example.org/scan"}},
			allowed:    []string{shelfmarkAvailable, facsimileAvailable},
			want:       true,
		},
		{
			name:       "no shelfmarks do not imply unknown copyright",
			shelfmarks: nil,
			allowed:    []string{copyrightStatusUnknown},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edition := &model.Edition{Shelfmarks: tt.shelfmarks}
			if got := editionHasAnyShelfmarkProperty(edition, tt.allowed); got != tt.want {
				t.Fatalf("editionHasAnyShelfmarkProperty() = %v, want %v", got, tt.want)
			}
		})
	}
}
