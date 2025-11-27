package pagesparser

import (
	"fmt"
	"strconv"
	"strings"
)

func ToString(pages []int) string {
	var parts []string
	if len(pages) == 0 {
		return ""
	}
	start := pages[0]
	prev := pages[0]
	for i := 1; i < len(pages); i++ {
		if pages[i] == prev+1 {
			prev = pages[i]
			continue
		}
		if start == prev {
			parts = append(parts, fmt.Sprintf("%d", start))
		} else {
			parts = append(parts, fmt.Sprintf("%d-%d", start, prev))
		}
		start = pages[i]
		prev = pages[i]
	}
	if start == prev {
		parts = append(parts, fmt.Sprintf("%d", start))
	} else {
		parts = append(parts, fmt.Sprintf("%d-%d", start, prev))
	}
	return strings.Join(parts, ",")
}

func Parse(pageStr string) ([]int, error) {
	ranges := strings.Split(pageStr, ",")
	var pages []int
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if !strings.Contains(r, "-") {
			page, err := parsePageNumber(r)
			if err != nil {
				return nil, err
			}
			pages = append(pages, page)
			continue
		}
		bounds := strings.Split(r, "-")
		if len(bounds) != 2 {
			continue
		}
		start, err := parsePageNumber(bounds[0])
		if err != nil {
			return nil, err
		}
		end, err := parsePageNumber(bounds[1])
		if err != nil {
			return nil, err
		}
		if end < start {
			return nil, fmt.Errorf("invalid page number range: %s", r)
		}
		for i := start; i <= end; i++ {
			pages = append(pages, i)
		}
	}
	return pages, nil
}

func parsePageNumber(pageStr string) (int, error) {
	p, err := strconv.Atoi(pageStr)
	if err != nil {
		return -1, fmt.Errorf("invalid page number: %s", pageStr)
	}
	if p < 1 {
		return -1, fmt.Errorf("page number must be positive: %d", p)
	}
	return p, nil
}
