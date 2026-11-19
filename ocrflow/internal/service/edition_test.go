package service

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/stretchr/testify/require"
)

func TestParseSubjectCategories(t *testing.T) {
	require.Equal(t, []model.EditionSubjectCategory{
		{Category: "Music Theory", Classification: "primary"},
		{Category: "Theoretical Mathematics", Classification: "secondary"},
	}, parseSubjectCategories([]string{
		"Music Theory::primary",
		"malformed value",
		"Theoretical Mathematics::secondary",
	}))
}
