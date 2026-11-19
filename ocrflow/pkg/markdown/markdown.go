package markdown

import (
	"fmt"
	"strings"
)

const HeaderPrefix = "header"

type Markdown struct {
	Content string
}

func (m *Markdown) GetCategories() []string {
	if m == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var categories []string
	for _, line := range strings.Split(m.Content, "\n") {
		category := ""
		if level, _ := ParseHeader(line); level > 0 {
			category = fmt.Sprintf("%s%d", HeaderPrefix, level)
		} else if level, _ := ParseCuratedHeading(line); level > 0 {
			category = fmt.Sprintf("%s%d", CuratedHeadingPrefix, level)
		}
		if category == "" {
			continue
		}
		if _, ok := seen[category]; ok {
			continue
		}
		seen[category] = struct{}{}
		categories = append(categories, category)
	}
	return categories
}

func ParseHeader(line string) (int, string) {
	line = strings.TrimSpace(line)
	if line == "" || !strings.HasPrefix(line, "#") {
		return 0, ""
	}

	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 {
		return 0, ""
	}
	if len(line) > level && line[level] != ' ' && line[level] != '\t' {
		return 0, ""
	}
	return level, strings.TrimSpace(line[level:])
}

// ParseHeaderBlock parses a Markdown heading and joins immediately following
// headings at the same level. Blank lines between the heading lines are ignored.
// The returned next index points to the first line after the last joined heading.
func ParseHeaderBlock(lines []string, start int) (level int, content string, next int) {
	if start < 0 || start >= len(lines) {
		return 0, "", start
	}

	level, content = ParseHeader(lines[start])
	if level == 0 {
		return 0, "", start
	}

	parts := []string{content}
	next = start + 1
	for {
		candidate := next
		for candidate < len(lines) && strings.TrimSpace(lines[candidate]) == "" {
			candidate++
		}
		candidateLevel, candidateContent := 0, ""
		if candidate < len(lines) {
			candidateLevel, candidateContent = ParseHeader(lines[candidate])
		}
		if candidateLevel != level {
			break
		}
		parts = append(parts, candidateContent)
		next = candidate + 1
	}

	return level, strings.Join(parts, " "), next
}
