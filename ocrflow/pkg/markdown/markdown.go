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
		level, _ := ParseHeader(line)
		if level == 0 {
			continue
		}
		category := fmt.Sprintf("%s%d", HeaderPrefix, level)
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
