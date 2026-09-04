package transcriptioncorrector

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/stretchr/testify/require"
)

type execCall struct {
	provider       string
	model          string
	prompt         llm.Prompt
	attachmentPath string
	logLabel       string
}

type fakeExecutor struct {
	responses []llm.Result
	calls     []execCall
}

func (f *fakeExecutor) ExecPromptResultWithLogLabel(provider, model string, prompt llm.Prompt, attachmentPath, logLabel string) (llm.Result, error) {
	f.calls = append(f.calls, execCall{
		provider: provider, model: model, prompt: prompt, attachmentPath: attachmentPath, logLabel: logLabel,
	})
	if len(f.responses) == 0 {
		return llm.Result{}, fmt.Errorf("unexpected call")
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func TestDiscoverPagesIgnoresVariantsAndSorts(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "page-0010.jpg"), []byte("image"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "page-0002.png"), []byte("image"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.png"), []byte("image"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "_variants"), 0o755))

	pages, err := discoverPages(dir)
	require.NoError(t, err)
	require.Equal(t, []page{
		{key: "page-0002", imagePath: filepath.Join(dir, "page-0002.png")},
		{key: "page-0010", imagePath: filepath.Join(dir, "page-0010.jpg")},
	}, pages)
}

func TestMarkdownPathSupportsCanonicalAndFlatLayouts(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "page-0001"), 0o755))
	canonical := filepath.Join(dir, "page-0001", "original.md")
	require.NoError(t, os.WriteFile(canonical, []byte("one"), 0o644))
	flat := filepath.Join(dir, "page-0002.md")
	require.NoError(t, os.WriteFile(flat, []byte("two"), 0o644))

	got, err := markdownPath(dir, "page-0001")
	require.NoError(t, err)
	require.Equal(t, canonical, got)
	got, err = markdownPath(dir, "page-0002")
	require.NoError(t, err)
	require.Equal(t, flat, got)
}

func TestLoadCandidatesConvertsALTOToMarkdown(t *testing.T) {
	dir := t.TempDir()
	doc := testALTO("Euclides", "Elementa")
	require.NoError(t, alto.SaveToFile(doc, filepath.Join(dir, "page-0001.xml")))

	candidates, err := loadCandidates(nil, []string{dir}, nil, "page-0001")
	require.NoError(t, err)
	require.Len(t, candidates, 1)
	require.Contains(t, candidates[0].label, "ALTOToMarkdown")
	require.Equal(t, "Euclides Elementa\n", candidates[0].text)
}

func TestLoadCandidatesReadsMixedTranscriptionDirectory(t *testing.T) {
	dir := t.TempDir()
	pageDir := filepath.Join(dir, "page-0001")
	require.NoError(t, os.MkdirAll(pageDir, 0o755))
	require.NoError(t, alto.SaveToFile(testALTO("ALTO", "text"), filepath.Join(pageDir, "original.xml")))
	require.NoError(t, os.WriteFile(filepath.Join(pageDir, "original.md"), []byte("Markdown text\n"), 0o644))

	candidates, err := loadCandidates(nil, nil, []string{dir}, "page-0001")
	require.NoError(t, err)
	require.Len(t, candidates, 2)
	require.Equal(t, "ALTO text\n", candidates[0].text)
	require.Equal(t, "Markdown text\n", candidates[1].text)
}

func TestNormalizeResponseStripsOnlySurroundingFence(t *testing.T) {
	got, err := normalizeResponse("```markdown\n# Heading\n\nText\n```\n")
	require.NoError(t, err)
	require.Equal(t, "# Heading\n\nText\n", got)

	got, err = normalizeResponse("Text\n\n```latin\nverbum\n```\n")
	require.NoError(t, err)
	require.Equal(t, "Text\n\n```latin\nverbum\n```\n", got)
}

func TestBuildPromptDefinesCanonicalMarkdownOutputContract(t *testing.T) {
	prompt := buildPrompt("page-0086", 2, 3, []candidate{{label: "source", text: "text"}}, "previous")

	require.Contains(t, prompt.Static, "Your entire response is written directly to the transcription file")
	require.Contains(t, prompt.Static, "--- BEGIN TRANSCRIPTION MARKDOWN DIALECT ---")
	require.Contains(t, prompt.Static, "## LLM operational contract")
	require.Contains(t, prompt.Static, "<!-- Running title: TEXT -->")
	require.Contains(t, prompt.Static, "Never use headings for page furniture.")
	require.Contains(t, prompt.Static, "<!-- Page number: TEXT -->")
	require.Contains(t, prompt.Static, "<!-- Catchword: TEXT -->")
	require.Contains(t, prompt.Static, `statements such as "the image confirms..."`)
	require.Contains(t, prompt.Static, "text copied from the previous-page context")
	require.Contains(t, prompt.Dynamic, "Current page: page-0086")
	require.Contains(t, prompt.Dynamic, "--- BEGIN PREVIOUS PAGE ---")
	require.NotEmpty(t, prompt.CacheKey)
}

func TestGeneratedMarkdownDialectMatchesSharedDocumentation(t *testing.T) {
	document, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "MARKDOWN_DIALECT.md"))
	require.NoError(t, err)
	const beginMarker = "<!-- BEGIN LLM CONTRACT -->"
	const endMarker = "<!-- END LLM CONTRACT -->"
	documentText := string(document)
	start := strings.Index(documentText, beginMarker)
	finish := strings.Index(documentText, endMarker)
	require.Greater(t, start, -1)
	require.Greater(t, finish, start)
	expected := strings.TrimSpace(documentText[start+len(beginMarker):finish]) + "\n"
	require.Equal(t, expected, markdownDialect,
		"run `go generate ./pkg/transcriptioncorrector` after editing docs/MARKDOWN_DIALECT.md")
}

func TestLineDiffCountsAddedAndDeletedLines(t *testing.T) {
	stats := lineDiff("one\ntwo\nthree\n", "one\nchanged\nthree\nfour\n")
	require.Equal(t, diffStats{added: 2, deleted: 1}, stats)
}

func TestRunLogsAggregateUsageAndProviderReportedCost(t *testing.T) {
	root := t.TempDir()
	imagesDir := filepath.Join(root, "images")
	sourceDir := filepath.Join(root, "source")
	outputDir := filepath.Join(root, "output")
	require.NoError(t, os.MkdirAll(imagesDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "page-0001"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(imagesDir, "page-0001.png"), []byte("image"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "page-0001", "original.md"), []byte("source\n"), 0o644))

	costOne := 0.012345
	costTwo := 0.023456
	fake := &fakeExecutor{responses: []llm.Result{
		{Text: "round one", Usage: llm.Usage{InputTokens: 100, CachedInputTokens: 20, CacheCreationInputTokens: 5, CacheMetricsAvailable: true, OutputTokens: 30, TotalTokens: 155, CostUSD: &costOne}},
		{Text: "round two", Usage: llm.Usage{InputTokens: 110, CachedInputTokens: 25, CacheCreationInputTokens: 10, CacheMetricsAvailable: true, OutputTokens: 40, TotalTokens: 185, CostUSD: &costTwo}},
	}}
	var logs strings.Builder
	cfg := Config{
		MarkdownDirs: []string{sourceDir}, ImagesDir: imagesDir, OutputDir: outputDir,
		Rounds: 2, Provider: llm.ProviderClaudeCode, Model: "fable", Logger: log.New(&logs, "", 0),
	}

	usage, err := Run(cfg, fake)
	require.NoError(t, err)
	require.EqualValues(t, 210, usage.InputTokens)
	require.EqualValues(t, 45, usage.CachedInputTokens)
	require.EqualValues(t, 15, usage.CacheCreationInputTokens)
	require.EqualValues(t, 70, usage.OutputTokens)
	require.EqualValues(t, 340, usage.TotalTokens)
	require.NotNil(t, usage.CostUSD)
	require.InDelta(t, 0.035801, *usage.CostUSD, 0.0000001)
	require.Contains(t, logs.String(), "requests=2")
	require.Contains(t, logs.String(), "tokens_input=210")
	require.Contains(t, logs.String(), "tokens_cached=45")
	require.Contains(t, logs.String(), "tokens_cache_creation=15")
	require.Contains(t, logs.String(), "tokens_output=70")
	require.Contains(t, logs.String(), "tokens_total=340")
	require.Contains(t, logs.String(), "cost_usd=0.035801")
	require.Contains(t, logs.String(), "cost_reports=2/2")
	require.Contains(t, logs.String(), "cache_read_requests=1")
	require.Contains(t, logs.String(), "cached_tokens=45")
}

func TestRunSkipsExistingRoundOutput(t *testing.T) {
	root := t.TempDir()
	imagesDir := filepath.Join(root, "images")
	sourceDir := filepath.Join(root, "source")
	outputDir := filepath.Join(root, "output")
	reusedOutput := filepath.Join(outputDir, "page-0001", "round-01.md")
	require.NoError(t, os.MkdirAll(imagesDir, 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "page-0001"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(reusedOutput), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(imagesDir, "page-0001.png"), []byte("image"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "page-0001", "original.md"), []byte("source\n"), 0o644))
	require.NoError(t, os.WriteFile(reusedOutput, []byte("existing\n"), 0o644))

	fake := &fakeExecutor{}
	usage, err := Run(Config{
		MarkdownDirs: []string{sourceDir}, ImagesDir: imagesDir, OutputDir: outputDir,
		Rounds: 1, SkipExisting: true, Provider: llm.ProviderClaudeCode, Model: "fable",
	}, fake)
	require.NoError(t, err)
	require.Empty(t, fake.calls)
	require.Equal(t, llm.Usage{}, usage)
	final, err := os.ReadFile(filepath.Join(outputDir, "page-0001", "original.md"))
	require.NoError(t, err)
	require.Equal(t, "existing\n", string(final))
}

func testALTO(words ...string) *alto.Alto {
	tokens := make([]alto.String, 0, len(words))
	for _, word := range words {
		tokens = append(tokens, alto.String{Content: word})
	}
	return &alto.Alto{Layout: alto.Layout{Page: []alto.Page{{PrintSpace: alto.PrintSpace{TextBlocks: []alto.TextBlock{{
		Lines: []alto.TextLine{{Strings: tokens}},
	}}}}}}}
}
