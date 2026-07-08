package markdown

import (
	"fmt"
	"strings"
)

func ExtractCategoryContentsFromMarkdown(md *Markdown, categories []string, lineBreakSeperator string) ([]Category, error) {
	if md == nil {
		return nil, nil
	}

	allowed := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		allowed[c] = struct{}{}
	}

	var results []Category
	for _, line := range strings.Split(md.Content, "\n") {
		level, content := ParseHeader(line)
		if level == 0 {
			continue
		}
		category := fmt.Sprintf("%s%d", HeaderPrefix, level)
		if len(allowed) > 0 {
			if _, ok := allowed[category]; !ok {
				continue
			}
		}
		results = append(results, Category{
			Category: category,
			Content:  content,
		})
	}
	return results, nil
}

type Category struct {
	Category string
	Content  string
}
