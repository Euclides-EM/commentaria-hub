package markdown

import (
	"fmt"
	"strings"
)

func ExtractCategoryContentsFromMarkdown(md *Markdown, categories []string, lineBreakSeperator string, includeCuratedHeadings bool) ([]Category, error) {
	if md == nil {
		return nil, nil
	}

	allowed := make(map[string]struct{}, len(categories))
	for _, c := range categories {
		allowed[c] = struct{}{}
	}

	var results []Category
	lines := strings.Split(md.Content, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		category := ""
		content := ""
		if level, headingContent, next := ParseHeaderBlock(lines, i); level > 0 {
			category = fmt.Sprintf("%s%d", HeaderPrefix, level)
			content = headingContent
			i = next - 1
		} else if level, headingContent := ParseCuratedHeading(line); includeCuratedHeadings && level > 0 {
			category = fmt.Sprintf("%s%d", CuratedHeadingPrefix, level)
			content = headingContent
		}
		if category == "" {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[category]; !ok {
				continue
			}
		}
		results = append(results, Category{
			Category: category,
			Content:  ExpandDropcaps(content),
		})
	}
	return results, nil
}

type Category struct {
	Category string
	Content  string
}
