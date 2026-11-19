package markdown

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
)

// ExtractIndexContentsFromMarkdown returns headings in document order, marking
// those outside the requested index layers as hidden, and, in the same pass,
// all layer names present in the document.
// Printed headings and untyped curated headings belong to the default layer.
func ExtractIndexContentsFromMarkdown(md *Markdown, categories, indexTypes []string, normalize bool) ([]Category, []string, error) {
	if md == nil {
		return nil, []string{DefaultIndexType}, nil
	}

	allowedCategories := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		allowedCategories[category] = struct{}{}
	}
	selectedTypes := make(map[string]struct{}, len(indexTypes))
	for _, indexType := range indexTypes {
		selectedTypes[indexType] = struct{}{}
	}
	if len(selectedTypes) == 0 {
		selectedTypes[DefaultIndexType] = struct{}{}
	}

	available := map[string]struct{}{DefaultIndexType: {}}
	var results []Category
	lines := strings.Split(md.Content, "\n")
	for i := 0; i < len(lines); i++ {
		category, content, indexType := "", "", DefaultIndexType
		if level, headingContent, next := ParseHeaderBlock(lines, i); level > 0 {
			category = fmt.Sprintf("%s%d", HeaderPrefix, level)
			content = headingContent
			i = next - 1
		} else if heading, ok := ParseTypedCuratedHeading(lines[i]); ok {
			category = fmt.Sprintf("%s%d", CuratedHeadingPrefix, heading.Level)
			content = heading.Text
			indexType = heading.Type
			available[indexType] = struct{}{}
		}
		if category == "" {
			continue
		}
		if len(allowedCategories) > 0 {
			if _, ok := allowedCategories[category]; !ok {
				continue
			}
		}
		if _, ok := selectedTypes[indexType]; !ok {
			results = append(results, Category{
				Category:  category,
				IndexType: indexType,
				Hidden:    true,
			})
			continue
		}
		content = StripEmphasis(ResolvePrinterErrorCorrections(ExpandDropcaps(content), normalize))
		if normalize {
			content = normalizeIndexHeadingContent(content)
		}
		results = append(results, Category{
			Category:  category,
			Content:   content,
			IndexType: indexType,
		})
	}

	availableTypes := make([]string, 0, len(available))
	for indexType := range available {
		availableTypes = append(availableTypes, indexType)
	}
	slices.Sort(availableTypes)
	if i := slices.Index(availableTypes, DefaultIndexType); i > 0 {
		availableTypes[0], availableTypes[i] = availableTypes[i], availableTypes[0]
	}
	return results, availableTypes, nil
}

func normalizeIndexHeadingContent(content string) string {
	return strings.TrimRightFunc(strings.TrimLeftFunc(content, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}), unicode.IsSpace)
}

type Category struct {
	Category  string
	Content   string
	IndexType string
	Hidden    bool
}
