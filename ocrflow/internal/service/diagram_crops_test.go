package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
)

func TestFacsimileEditionDiagramFallbackAllowed(t *testing.T) {
	fac := &model.Facsimile{Meta: common.NewMeta("fac_one"), EditionID: "Paris_1615"}
	if !facsimileEditionDiagramFallbackAllowed(fac, []*model.Facsimile{fac}) {
		t.Fatal("single facsimile edition should allow edition-key diagram fallback")
	}

	other := &model.Facsimile{Meta: common.NewMeta("fac_two"), EditionID: "Paris_1615"}
	if facsimileEditionDiagramFallbackAllowed(fac, []*model.Facsimile{fac, other}) {
		t.Fatal("multiple unmapped facsimiles should not allow edition-key diagram fallback")
	}

	fac.ShelfmarkID = "shm_one"
	if !facsimileEditionDiagramFallbackAllowed(fac, []*model.Facsimile{fac, other}) {
		t.Fatal("only mapped facsimile in an edition should allow edition-key diagram fallback")
	}
}
