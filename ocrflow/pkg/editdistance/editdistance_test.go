package editdistance

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunes(t *testing.T) {
	require.Equal(t, 3, Runes([]rune("kitten"), []rune("sitting")))
	require.Equal(t, 1, Runes([]rune("Εὐκλείδης"), []rune("Εὐκλείδη")))
}

func TestBoundedRunes(t *testing.T) {
	distance := BoundedRunes([]rune("kitten"), []rune("sitting"), 3)
	require.Equal(t, 3, distance)

	distance = BoundedRunes([]rune("kitten"), []rune("sitting"), 2)
	require.Equal(t, 3, distance)
}
