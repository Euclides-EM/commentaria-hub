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

func IntRange(pageStr string) ([]int, error) {
	ranges := strings.Split(pageStr, ",")
	var pages []int
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if !strings.Contains(r, "-") {
			page, err := PageNumber(r)
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
		start, err := PageNumber(bounds[0])
		if err != nil {
			return nil, err
		}
		end, err := PageNumber(bounds[1])
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

func Range(pageStr string) ([]string, error) {
	ranges := strings.Split(pageStr, ",")
	var pages []string
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if !strings.Contains(r, "-") {
			page, err := PageNumber(r)
			if err == nil {
				r = fmt.Sprintf("%d", page)
			}
			pages = append(pages, r)
			continue
		}
		bounds := strings.Split(r, "-")
		if len(bounds) != 2 {
			continue
		}
		start, err := PageNumber(bounds[0])
		if err != nil {
			return nil, err
		}
		end, err := PageNumber(bounds[1])
		if err != nil {
			return nil, err
		}
		if end < start {
			return nil, fmt.Errorf("invalid page number range: %s", r)
		}
		for i := start; i <= end; i++ {
			pages = append(pages, fmt.Sprintf("%d", i))
		}
	}
	return pages, nil
}

func PageNumber(pageStr string) (int, error) {
	p, err := strconv.Atoi(pageStr)
	if err != nil {
		return -1, fmt.Errorf("invalid page number: %s", pageStr)
	}
	if p < 1 {
		return -1, fmt.Errorf("page number must be positive: %d", p)
	}
	return p, nil
}

func PageToPNGFilename(p int) string {
	return PageToFilename(p, "png")
}

func PageOrKeyToPNGFilename(pageOrKey string) string {
	asInt, err := strconv.Atoi(pageOrKey)
	if err == nil {
		return PageToPNGFilename(asInt)
	}
	return fmt.Sprintf("%s.png", pageOrKey)
}

func PageToXMLFilename(p int) string {
	return PageToFilename(p, "xml")
}

func PageToFilename(p int, ext string) string {
	if ext == "" {
		return fmt.Sprintf("page-%04d", p)
	}
	return fmt.Sprintf("page-%04d.%s", p, ext)
}

func FileNameToPage(filename string) (int, error) {
	var pageNum int
	_, err := fmt.Sscanf(filename, "page-%04d.", &pageNum)
	if err != nil {
		return -1, fmt.Errorf("invalid page filename: %s", filename)
	}
	return pageNum, nil
}
