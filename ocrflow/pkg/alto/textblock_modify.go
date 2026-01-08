package alto

import (
	"log"
	"sort"
)

func ApplyTextBlockCorrectionALTO(a *Alto, textBlockID string, newContent []string) error {
	for _, page := range a.Layout.Page {
		for _, textBlock := range page.PrintSpace.TextBlocks {
			if textBlock.ID != textBlockID {
				continue
			}
			lines := topNLongestLines(len(newContent), textBlock.Lines)
			for i, line := range lines {
				if len(line.Strings) > 0 {
					line.Strings[0].Content = newContent[i]
					line.Strings[0].WC = 1.0
					continue
				}
				if len(line.Strings) > 1 {
					line.Strings = line.Strings[0:1]
					log.Printf("Warning: Truncated extra strings in line %s of text block %s", line.ID, textBlockID)
					continue
				}
				line.Strings = []AltoString{
					{
						HPOS:    line.HPOS,
						VPOS:    line.VPOS,
						Width:   line.Width,
						Height:  line.Height,
						WC:      1.0,
						Content: newContent[i],
					},
				}
			}
		}
	}
	return nil
}

func topNLongestLines(n int, lines []TextLine) []TextLine {
	if n <= 0 || len(lines) == 0 {
		return nil
	}
	if n > len(lines) {
		n = len(lines)
	}

	// Work on indices so the original slice order/content is untouched.
	idx := make([]int, len(lines))
	for i := range lines {
		idx[i] = i
	}

	// Stable sort: empty lines always go to the end, otherwise by width descending.
	sort.SliceStable(idx, func(i, j int) bool {
		if IsEmptyLine(lines[idx[i]]) {
			return false
		}
		if IsEmptyLine(lines[idx[j]]) {
			return true
		}
		return lines[idx[i]].Width > lines[idx[j]].Width
	})

	out := make([]TextLine, 0, n)
	for k := 0; k < n; k++ {
		out = append(out, lines[idx[k]])
	}
	return out
}
