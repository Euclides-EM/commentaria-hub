package transcriptioncorrector

import (
	"errors"
	"fmt"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

type diffStats struct {
	added   int
	deleted int
}

//go:generate go run ./internal/gendialect -input ../../../docs/MARKDOWN_DIALECT.md -output markdown_dialect_generated.go

func buildPrompt(pageKey string, round, rounds int, candidates []candidate, previousPage string) string {
	var b strings.Builder
	fmt.Fprintf(&b, `You are correcting a scholarly markdown transcription from an image of an early printed page.

Use the attached image as the authority. Silently compare it with the supplied candidate transcriptions, reconcile disagreements, and make only evidence-based corrections. Preserve historical spelling, capitalization, punctuation, special characters, and textual order. Do not modernize, translate, summarize, invent illegible text, or describe your reasoning.

OUTPUT CONTRACT
Your entire response is written directly to the transcription file. Return only the complete corrected Markdown for the current page.

Never include:
- explanations, observations, confidence statements, or correction summaries;
- statements such as "the image confirms..." or "the only correction needed...";
- a surrounding Markdown code fence;
- text copied from the previous-page context;
- labels such as "Running title" or "Catchword" on their own line.

Follow the embedded Transcription Markdown dialect exactly. It is normative for output formatting:

--- BEGIN TRANSCRIPTION MARKDOWN DIALECT ---
%s
--- END TRANSCRIPTION MARKDOWN DIALECT ---

Before responding, silently verify that:
1. the response begins with transcription content, not commentary;
2. every catchword is contained in a single <!-- Catchword: ... --> comment;
3. running titles use <!-- Running title: ... -->;
4. page numbers use <!-- Page: ... -->;
5. no supplied boundary markers or instructions occur in the output.

Current page: %s
Correction round: %d of %d

CURRENT-PAGE TRANSCRIPTIONS:
`, strings.TrimSpace(markdownDialect), pageKey, round, rounds)
	for i, c := range candidates {
		fmt.Fprintf(&b, "\n--- BEGIN TRANSCRIPTION %d: %s ---\n%s\n--- END TRANSCRIPTION %d ---\n", i+1, c.label, strings.TrimSpace(c.text), i+1)
	}
	if strings.TrimSpace(previousPage) != "" {
		fmt.Fprintf(&b, `
PREVIOUS-PAGE CONTEXT (context only; do not copy text from it into the current page):
--- BEGIN PREVIOUS PAGE ---
%s
--- END PREVIOUS PAGE ---
`, strings.TrimSpace(previousPage))
	} else {
		b.WriteString("\nThere is no previous-page transcription for this page.\n")
	}
	return b.String()
}

func normalizeResponse(response string) (string, error) {
	text := strings.TrimSpace(strings.ReplaceAll(response, "\r\n", "\n"))
	if text == "" {
		return "", errors.New("response is empty")
	}
	lines := strings.Split(text, "\n")
	if len(lines) >= 2 && strings.HasPrefix(strings.TrimSpace(lines[0]), "```") && strings.TrimSpace(lines[len(lines)-1]) == "```" {
		text = strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	if text == "" {
		return "", errors.New("response contains no markdown")
	}
	return text + "\n", nil
}

func lineDiff(before, after string) diffStats {
	matcher := difflib.NewMatcher(difflib.SplitLines(before), difflib.SplitLines(after))
	var stats diffStats
	for _, opcode := range matcher.GetOpCodes() {
		switch opcode.Tag {
		case 'd':
			stats.deleted += opcode.I2 - opcode.I1
		case 'i':
			stats.added += opcode.J2 - opcode.J1
		case 'r':
			stats.deleted += opcode.I2 - opcode.I1
			stats.added += opcode.J2 - opcode.J1
		}
	}
	return stats
}

func lineCount(text string) int {
	text = strings.TrimSuffix(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if text == "" {
		return 0
	}
	return strings.Count(text, "\n") + 1
}
