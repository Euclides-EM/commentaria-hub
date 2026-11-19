package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitCategories(t *testing.T) {
	require.Equal(t, []string{"MainZone", "MarginTextZone"}, splitCategories(" MainZone, MarginTextZone, "))
}

func TestParseReassignment(t *testing.T) {
	rule, err := parseReassignment("MainZone:MainZone-Head--Section:5:0.85")
	require.NoError(t, err)
	require.Equal(t, "MainZone", rule.FromCategory)
	require.Equal(t, "MainZone-Head--Section", rule.ToCategory)
	require.Equal(t, 5.0, rule.PrecisionPx)
	require.Equal(t, 0.85, rule.MinOverlap)
}
