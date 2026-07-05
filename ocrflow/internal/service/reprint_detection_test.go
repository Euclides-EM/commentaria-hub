package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/stretchr/testify/require"
)

func TestNormalizeReprintDetectionValues(t *testing.T) {
	require.Equal(t, "éléments d euclide", normalizeText(" Éléments, d'Euclide! "))
	require.Equal(t, "commandino;euclid", normalizeList([]string{"Euclid", " Commandino "}))
}

func TestTitlesAlmostEqualAllowsMinorTypographicalVariation(t *testing.T) {
	original := "DE GLI ELEMENTI DI EVCLIDE Li Primi sei Libri Tradotti in lingua Italiana. ALL’ILLVSTRISS. SENATO DI BOLOGNA."
	suspected := "DE GLI ELEM-ENTI DI EUCLIDE Li Primi sei Libri Tradotti in lingua Italiana. ALL’ILLVSTRISS. SENATO DI BOLOGNA."
	require.True(t, titlesAlmostEqual(original, suspected))
	require.False(t, titlesAlmostEqual(original, "A substantially different mathematical work with another title"))
}

func TestNormalizedLanguagesMustMatch(t *testing.T) {
	require.NotEqual(t, normalizeList([]string{"English"}), normalizeList([]string{"Italian"}))
}

func TestReprintMetadataMatches(t *testing.T) {
	originalTitle := "DE GLI ELEMENTI DI EVCLIDE Li Primi sei Libri Tradotti in lingua Italiana. ALL’ILLVSTRISS. SENATO DI BOLOGNA."
	suspectedTitle := "DE GLI ELEM-ENTI DI EUCLIDE Li Primi sei Libri Tradotti in lingua Italiana. ALL’ILLVSTRISS. SENATO DI BOLOGNA."
	original := &model.Edition{Title: &originalTitle, Editor: []string{"Euclid", "Commandino"}, Languages: []string{"Italian"}}
	suspected := &model.Edition{Title: &suspectedTitle, Editor: []string{"Commandino", "Other"}, Languages: []string{"Italian"}}
	require.True(t, reprintMetadataMatches(original, suspected))

	suspected.Editor = []string{"Other"}
	require.False(t, reprintMetadataMatches(original, suspected))
	suspected.Editor = []string{"Commandino"}
	suspected.Languages = []string{"Latin"}
	require.False(t, reprintMetadataMatches(original, suspected))
}
