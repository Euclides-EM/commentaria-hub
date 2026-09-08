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
	for _, line := range strings.Split(md.Content, "\n") {
		category := ""
		content := ""
		if level, headingContent := ParseHeader(line); level > 0 {
			category = fmt.Sprintf("%s%d", HeaderPrefix, level)
			content = headingContent
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
