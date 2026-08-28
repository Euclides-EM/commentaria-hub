package service

import (
	"errors"
	"slices"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

func TestMatchingCommentariaFacsimileReturnsAmbiguousNames(t *testing.T) {
	_, err := matchingCommentariaFacsimile("Paris_1615",
		&model.Facsimile{Meta: common.Meta{Name: "Paris_1615_archive"}, ScanURL: "file:///tmp/Paris_1615_archive.pdf"},
		[]*model.Facsimile{
			{Meta: common.Meta{ID: "fac1", Name: "Paris_1615_bnf"}, ScanURL: "file:///tmp/Paris_1615_bnf.pdf"},
			{Meta: common.Meta{ID: "fac2", Name: "Paris_1615_google"}, ScanURL: "file:///tmp/Paris_1615_google.pdf"},
		},
	)
	var ambiguous *AmbiguousCommentariaFacsimilesError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("error = %v, want AmbiguousCommentariaFacsimilesError", err)
	}
	if !slices.Equal(ambiguous.FacsimileNames, []string{"Paris_1615_bnf", "Paris_1615_google"}) {
		t.Fatalf("facsimile names = %v", ambiguous.FacsimileNames)
	}
}
