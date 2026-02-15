package formatcov

import (
	"fmt"
	"strconv"
	"strings"
)

func IntsToCompressedStr(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	seen := make(map[int]struct{})
	for _, n := range nums {
		seen[n] = struct{}{}
	}
	var sorted []int
	for n := range seen {
		sorted = append(sorted, n)
	}
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	var ranges []string
	start, prev := sorted[0], sorted[0]
	for i := 1; i <= len(sorted); i++ {
		curr := 0
		if i < len(sorted) {
			curr = sorted[i]
		} else {
			curr = prev + 2
		}
		if curr != prev+1 {
			if start == prev {
				ranges = append(ranges, strconv.Itoa(start))
			} else {
				ranges = append(ranges, fmt.Sprintf("%d-%d", start, prev))
			}
			if i < len(sorted) {
				start = sorted[i]
				prev = sorted[i]
			}
		} else {
			prev = curr
		}
	}
	return strings.Join(ranges, ", ")
}
