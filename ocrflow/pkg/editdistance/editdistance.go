// Package editdistance provides Unicode-aware Levenshtein distance helpers.
package editdistance

// Runes returns the Levenshtein edit distance between two rune slices.
func Runes(a, b []rune) int {
	if len(a) > len(b) {
		a, b = b, a
	}
	previous := make([]int, len(a)+1)
	for i := range previous {
		previous[i] = i
	}
	for row, right := range b {
		current := make([]int, len(a)+1)
		current[0] = row + 1
		for column, left := range a {
			cost := 0
			if left != right {
				cost = 1
			}
			current[column+1] = min(
				current[column]+1,
				previous[column+1]+1,
				previous[column]+cost,
			)
		}
		previous = current
	}
	return previous[len(a)]
}

// BoundedRunes returns maxDistance+1 as soon as the distance is known to
// exceed maxDistance.
func BoundedRunes(a, b []rune, maxDistance int) int {
	if maxDistance < 0 || abs(len(a)-len(b)) > maxDistance {
		return maxDistance + 1
	}
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		rowMin := current[0]
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			current[j] = min(
				previous[j]+1,
				current[j-1]+1,
				previous[j-1]+cost,
			)
			rowMin = min(rowMin, current[j])
		}
		if rowMin > maxDistance {
			return maxDistance + 1
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
