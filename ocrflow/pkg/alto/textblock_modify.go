package alto

import (
	"errors"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/samber/lo"
)

func ApplyTextBlockCorrectionALTO(a *Alto, textBlockID string, oldContent []string, newContent []string) error {
	if len(newContent) == 0 {
		return errors.New("newContent is empty")
	}

	for pi := range a.Layout.Page {
		page := &a.Layout.Page[pi]
		for tbi := range page.PrintSpace.TextBlocks {
			textBlock := &page.PrintSpace.TextBlocks[tbi]
			if textBlock.ID != textBlockID {
				continue
			}

			// 1) locate the span matching oldContent inside this text block
			// 2) replace that span so it ends up with exactly len(newContent) lines
			if err := replaceLinesInTextBlock(&textBlock.Lines, oldContent, newContent); err != nil {
				return fmt.Errorf("textBlock %s: %w", textBlockID, err)
			}

			return nil
		}
	}

	return fmt.Errorf("text block %s not found", textBlockID)
}

/*
Integration entrypoint for the algorithm
*/

func replaceLinesInTextBlock(actualLines *[]TextLine, oldContent, newContent []string) error {
	lines := *actualLines
	if len(lines) == 0 {
		return errors.New("text block has no lines")
	}
	if len(newContent) == 0 {
		return errors.New("newContent is empty")
	}

	start, spanLen, ok := findBestSpan(lines, oldContent)
	if !ok {
		// Fallback to your previous heuristic: pick "best" lines by width
		// and overwrite the first min(len(lines), len(newContent)).
		log.Printf("Warning: oldContent span not found, falling back to width-based replacement for %d lines", len(newContent))
		selected := linesToReplace(len(newContent), lines) // returns copies
		// We need indices to mutate the actual lines, so compute them again.
		idx := linesToReplaceIdx(len(newContent), lines)
		for i := 0; i < len(idx) && i < len(newContent); i++ {
			setLineContent(&lines[idx[i]], newContent[i])
		}
		*actualLines = lines
		_ = selected // keep old behavior conceptually, not used directly
		return nil
	}

	end := start + spanLen
	oldSpan := lines[start:end]
	plan := alignSpan(oldSpan, newContent)
	newSpan := applyPlan(oldSpan, newContent, plan)

	out := make([]TextLine, 0, len(lines)-len(oldSpan)+len(newSpan))
	out = append(out, lines[:start]...)
	out = append(out, newSpan...)
	out = append(out, lines[end:]...)
	*actualLines = out
	return nil
}

/*
Span matching
*/

func findBestSpan(lines []TextLine, oldContent []string) (start int, spanLen int, ok bool) {
	// If oldContent is empty, match a single "best" line.
	if len(oldContent) == 0 {
		bestI := -1
		bestScore := math.Inf(1)
		for i := range lines {
			s := normalizeText(ExtractTextFromLine(lines[i]))
			score := normalizedEditDistance(s, "")
			if score < bestScore {
				bestScore = score
				bestI = i
			}
		}
		if bestI >= 0 {
			return bestI, 1, true
		}
		return 0, 0, false
	}

	target := normalizeText(strings.Join(oldContent, "\n"))
	if target == "" {
		return 0, 0, false
	}

	want := len(oldContent)
	minL := lo.Max([]int{1, want - 2})
	maxL := lo.Min([]int{len(lines), want + 2})

	bestStart := -1
	bestLen := 0
	bestScore := math.Inf(1)

	for L := minL; L <= maxL; L++ {
		for i := 0; i+L <= len(lines); i++ {
			s := spanText(lines[i : i+L])
			score := normalizedEditDistance(s, target)
			score += 0.02 * math.Abs(float64(L-want))
			if score < bestScore {
				bestScore = score
				bestStart = i
				bestLen = L
			}
		}
	}

	// Tune threshold if needed.
	if bestStart >= 0 && bestScore <= 0.35 {
		return bestStart, bestLen, true
	}
	return 0, 0, false
}

func spanText(lines []TextLine) string {
	parts := make([]string, 0, len(lines))
	for i := range lines {
		parts = append(parts, ExtractTextFromLine(lines[i]))
	}
	return normalizeText(strings.Join(parts, "\n"))
}

func normalizeText(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.TrimSpace(s)
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

/*
Alignment planning
*/

type step struct {
	oldN int
	newN int
}

func alignSpan(oldSpan []TextLine, newContent []string) []step {
	m := len(oldSpan)
	n := len(newContent)

	oldTxt := make([]string, m)
	for i := 0; i < m; i++ {
		oldTxt[i] = normalizeText(ExtractTextFromLine(oldSpan[i]))
	}
	newTxt := make([]string, n)
	for j := 0; j < n; j++ {
		newTxt[j] = normalizeText(newContent[j])
	}

	const maxGroup = 3

	dp := make([][]float64, m+1)
	prev := make([][]step, m+1)
	for i := range dp {
		dp[i] = make([]float64, n+1)
		prev[i] = make([]step, n+1)
		for j := range dp[i] {
			dp[i][j] = math.Inf(1)
		}
	}
	dp[0][0] = 0

	concatOld := func(i, k int) string {
		var b strings.Builder
		for t := 0; t < k; t++ {
			if t > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(oldTxt[i+t])
		}
		return strings.TrimSpace(b.String())
	}
	concatNew := func(j, k int) string {
		var b strings.Builder
		for t := 0; t < k; t++ {
			if t > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(newTxt[j+t])
		}
		return strings.TrimSpace(b.String())
	}

	for i := 0; i <= m; i++ {
		for j := 0; j <= n; j++ {
			if !isFinite(dp[i][j]) {
				continue
			}

			// Merge: k old -> 1 new
			if j < n {
				for k := 1; k <= maxGroup && i+k <= m; k++ {
					cost := normalizedEditDistance(concatOld(i, k), newTxt[j])
					cost += 0.01 * float64(k-1)
					ni, nj := i+k, j+1
					if dp[i][j]+cost < dp[ni][nj] {
						dp[ni][nj] = dp[i][j] + cost
						prev[ni][nj] = step{oldN: k, newN: 1}
					}
				}
			}

			// Split: 1 old -> k new
			if i < m {
				for k := 1; k <= maxGroup && j+k <= n; k++ {
					cost := normalizedEditDistance(oldTxt[i], concatNew(j, k))
					cost += 0.01 * float64(k-1)
					ni, nj := i+1, j+k
					if dp[i][j]+cost < dp[ni][nj] {
						dp[ni][nj] = dp[i][j] + cost
						prev[ni][nj] = step{oldN: 1, newN: k}
					}
				}
			}
		}
	}

	if !isFinite(dp[m][n]) {
		// No path found, force 1:1 as much as possible
		// This should be rare with maxGroup>=3.
	}

	i, j := m, n
	var rev []step
	for i > 0 || j > 0 {
		st := prev[i][j]
		if st.oldN == 0 && st.newN == 0 {
			// Degenerate fallback to avoid infinite loops.
			if i > 0 && j > 0 {
				st = step{oldN: 1, newN: 1}
			} else if i > 0 {
				st = step{oldN: 1, newN: 0}
			} else {
				st = step{oldN: 0, newN: 1}
			}
		}
		rev = append(rev, st)
		i -= st.oldN
		j -= st.newN
	}

	plan := make([]step, len(rev))
	for a := 0; a < len(rev); a++ {
		plan[a] = rev[len(rev)-1-a]
	}
	return plan
}

func applyPlan(oldSpan []TextLine, newContent []string, plan []step) []TextLine {
	out := make([]TextLine, 0, len(newContent))
	oldIdx := 0
	newIdx := 0

	for _, st := range plan {
		if st.oldN <= 0 && st.newN <= 0 {
			continue
		}

		oldGroup := oldSpan[oldIdx : oldIdx+st.oldN]
		newGroup := newContent[newIdx : newIdx+st.newN]

		switch {
		case st.oldN == st.newN:
			for k := 0; k < st.newN; k++ {
				l := oldGroup[k]
				setLineContent(&l, newGroup[k])
				out = append(out, l)
			}

		case st.oldN > st.newN:
			for k := 0; k < st.newN; k++ {
				l := oldGroup[k]
				setLineContent(&l, newGroup[k])
				out = append(out, l)
			}

		case st.oldN < st.newN:
			for k := 0; k < st.oldN; k++ {
				l := oldGroup[k]
				setLineContent(&l, newGroup[k])
				out = append(out, l)
			}

			template := oldGroup[st.oldN-1]
			extra := st.newN - st.oldN
			inserted := makeSplitLines(template, extra)
			for e := 0; e < extra; e++ {
				l := inserted[e]
				setLineContent(&l, newGroup[st.oldN+e])
				out = append(out, l)
			}
		}

		oldIdx += st.oldN
		newIdx += st.newN
	}

	// Safety: enforce exact length.
	if len(out) != len(newContent) {
		if len(out) > len(newContent) {
			out = out[:len(newContent)]
		} else if len(out) < len(newContent) && len(out) > 0 {
			tpl := out[len(out)-1]
			missing := len(newContent) - len(out)
			pads := makeSplitLines(tpl, missing)
			for i := 0; i < missing; i++ {
				l := pads[i]
				setLineContent(&l, newContent[len(out)])
				out = append(out, l)
			}
		}
	}

	return out
}

/*
Line mutation helpers
*/

func setLineContent(line *TextLine, content string) {
	content = strings.TrimSpace(content)

	// Ensure exactly one AltoString.
	if len(line.Strings) == 0 {
		line.Strings = []AltoString{{
			HPOS:    line.HPOS,
			VPOS:    line.VPOS,
			Width:   line.Width,
			Height:  line.Height,
			WC:      1.0,
			Content: content,
		}}
		return
	}

	// Keep first string, drop rest.
	line.Strings = line.Strings[:1]
	line.Strings[0].Content = content
	line.Strings[0].WC = 1.0

	// Keep geometry consistent with line.
	line.Strings[0].HPOS = line.HPOS
	line.Strings[0].VPOS = line.VPOS
	line.Strings[0].Width = line.Width
	line.Strings[0].Height = line.Height
}

func makeSplitLines(tpl TextLine, k int) []TextLine {
	if k <= 0 {
		return nil
	}
	out := make([]TextLine, 0, k)

	if tpl.Height <= 0 {
		for i := 0; i < k; i++ {
			l := tpl
			l.ID = fmt.Sprintf("%s_split_%d", safeID(tpl.ID), i+1)
			l.Strings = nil
			out = append(out, l)
		}
		return out
	}

	chunk := tpl.Height / float64(k+1)
	if chunk <= 0 {
		chunk = tpl.Height
	}

	for i := 0; i < k; i++ {
		l := tpl
		l.ID = fmt.Sprintf("%s_split_%d", safeID(tpl.ID), i+1)
		l.VPOS = tpl.VPOS + chunk*float64(i+1)
		l.Height = chunk
		l.Strings = nil
		out = append(out, l)
	}
	return out
}

func safeID(id string) string {
	if id == "" {
		return "line"
	}
	return id
}

/*
Edit distance
*/

func normalizedEditDistance(a, b string) float64 {
	a = normalizeText(a)
	b = normalizeText(b)
	if a == b {
		return 0
	}
	if a == "" && b == "" {
		return 0
	}
	la := utf8.RuneCountInString(a)
	lb := utf8.RuneCountInString(b)
	den := float64(lo.Max([]int{la, lb}))
	if den == 0 {
		return 0
	}
	d := float64(levenshteinRunes([]rune(a), []rune(b)))
	return d / den
}

func levenshteinRunes(a, b []rune) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			cur[j] = lo.Min([]int{
				prev[j] + 1,
				cur[j-1] + 1,
				prev[j-1] + cost,
			})
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func isFinite(x float64) bool { return !math.IsInf(x, 0) && !math.IsNaN(x) }

/*
Your original selection heuristic, plus an idx helper
*/

func linesToReplace(n int, lines []TextLine) []TextLine {
	if n <= 0 || len(lines) == 0 {
		return nil
	}
	if n > len(lines) {
		n = len(lines)
	}

	idx := linesToReplaceIdx(n, lines)
	out := make([]TextLine, 0, n)
	for k := 0; k < n; k++ {
		out = append(out, lines[idx[k]])
	}
	return out
}

func linesToReplaceIdx(n int, lines []TextLine) []int {
	if n <= 0 || len(lines) == 0 {
		return nil
	}
	if n > len(lines) {
		n = len(lines)
	}

	idx := make([]int, len(lines))
	for i := range lines {
		idx[i] = i
	}

	sort.SliceStable(idx, func(i, j int) bool {
		if IsEmptyLine(lines[idx[i]]) {
			return false
		}
		if IsEmptyLine(lines[idx[j]]) {
			return true
		}
		return lines[idx[i]].Width > lines[idx[j]].Width
	})

	return idx[:n]
}
