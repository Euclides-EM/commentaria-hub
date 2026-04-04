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

// CompressedStrToInts parses a compressed string like "1, 2, 5-7" into []int{1, 2, 5, 6, 7}.
func CompressedStrToInts(s string) []int {
	if s == "" {
		return nil
	}
	var out []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		part = strings.ReplaceAll(part, "–", "-")
		if part == "" {
			continue
		}
		if strings.Contains(part, "-") {
			idx := strings.Index(part, "-")
			lo, hi := strings.TrimSpace(part[:idx]), strings.TrimSpace(part[idx+1:])
			loN, errLo := strconv.Atoi(lo)
			hiN, errHi := strconv.Atoi(hi)
			if errLo != nil || errHi != nil || loN > hiN {
				continue
			}
			for i := loN; i <= hiN; i++ {
				out = append(out, i)
			}
		} else {
			n, err := strconv.Atoi(part)
			if err != nil {
				continue
			}
			out = append(out, n)
		}
	}
	return out
}
