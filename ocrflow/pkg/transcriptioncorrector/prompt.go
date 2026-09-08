package transcriptioncorrector

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/pmezard/go-difflib/difflib"
)

func buildDirectoryPrompt(cfg Config, pages []page) (llm.Prompt, []string, error) {
	abs := func(path string) (string, error) {
		value, err := filepath.Abs(path)
		if err != nil {
			return "", fmt.Errorf("resolve path %s: %w", path, err)
		}
		return value, nil
	}
	imagesDir, err := abs(cfg.ImagesDir)
	if err != nil {
		return llm.Prompt{}, nil, err
	}
	outputDir, err := abs(cfg.OutputDir)
	if err != nil {
		return llm.Prompt{}, nil, err
	}
	type sourceGroup struct {
		label string
		paths []string
	}
	groups := []sourceGroup{{"Markdown", cfg.MarkdownDirs}, {"ALTO XML", cfg.ALTODirs}, {"mixed ALTO/Markdown", cfg.TranscriptionDirs}}
	readPaths := []string{imagesDir}
	var sources strings.Builder
	for _, group := range groups {
		for _, sourcePath := range group.paths {
			absolutePath, err := abs(sourcePath)
			if err != nil {
				return llm.Prompt{}, nil, err
			}
			readPaths = append(readPaths, absolutePath)
			fmt.Fprintf(&sources, "- %s: %s\n", group.label, absolutePath)
		}
	}
	pageKeys := make([]string, len(pages))
	for i, p := range pages {
		pageKeys[i] = p.key
	}
	static := fmt.Sprintf(`You are correcting scholarly Markdown transcriptions of early printed pages. Work directly with the local files whose absolute paths are supplied. For each requested page, inspect its image and every available candidate transcription, reconcile disagreements using the image as authority, and write the complete corrected transcription to the requested output path.

Do not modify input files or any file outside the output directory. Do not create correction-round files. Preserve historical text and follow this normative transcription dialect:

--- BEGIN TRANSCRIPTION MARKDOWN DIALECT ---
%s
--- END TRANSCRIPTION MARKDOWN DIALECT ---`, strings.TrimSpace(markdownDialect))
	dynamic := fmt.Sprintf(`Images directory: %s
Candidate transcription directories:
%sOutput directory: %s
Pages to correct: %s

For each page key, find the matching page-NNNN image and candidate files. ALTO XML candidates must be interpreted as OCR text. Write only the corrected transcription to OUTPUT_DIRECTORY/PAGE_KEY/original.md. Complete every requested page.`, imagesDir, sources.String(), outputDir, strings.Join(pageKeys, ", "))
	cacheHash := sha256.Sum256([]byte(static))
	return llm.Prompt{Static: static, Dynamic: dynamic, CacheKey: fmt.Sprintf("transcription-corrector-directory-%x", cacheHash[:12])}, readPaths, nil
}

type diffStats struct {
	added   int
	deleted int
}

//go:generate go run ./internal/gendialect -input ../../../docs/MARKDOWN_DIALECT.md -output markdown_dialect_generated.go

func buildPrompt(pageKey string, round, rounds int, candidates []candidate, previousPage string) llm.Prompt {
	static := fmt.Sprintf(`You are correcting a scholarly markdown transcription from an image of an early printed page.

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
4. page and folio numbers use <!-- Page number: ... -->;
5. no supplied boundary markers or instructions occur in the output.`, strings.TrimSpace(markdownDialect))

	var dynamic strings.Builder
	fmt.Fprintf(&dynamic, `Current page: %s
Correction round: %d of %d

CURRENT-PAGE TRANSCRIPTIONS:
`, pageKey, round, rounds)
	for i, c := range candidates {
		fmt.Fprintf(&dynamic, "\n--- BEGIN TRANSCRIPTION %d: %s ---\n%s\n--- END TRANSCRIPTION %d ---\n", i+1, c.label, strings.TrimSpace(c.text), i+1)
	}
	if strings.TrimSpace(previousPage) != "" {
		fmt.Fprintf(&dynamic, `
PREVIOUS-PAGE CONTEXT (context only; do not copy text from it into the current page):
--- BEGIN PREVIOUS PAGE ---
%s
--- END PREVIOUS PAGE ---
`, strings.TrimSpace(previousPage))
	} else {
		dynamic.WriteString("\nThere is no previous-page transcription for this page.\n")
	}
	cacheHash := sha256.Sum256([]byte(static))
	return llm.Prompt{
		Static:   static,
		Dynamic:  dynamic.String(),
		CacheKey: fmt.Sprintf("transcription-corrector-%x", cacheHash[:12]),
	}
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
