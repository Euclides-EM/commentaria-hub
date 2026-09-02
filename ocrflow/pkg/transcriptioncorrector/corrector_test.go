package transcriptioncorrector

import (
	"fmt"
	"io"
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
	prompt         string
	attachmentPath string
	logLabel       string
}

type fakeExecutor struct {
	responses []string
	calls     []execCall
}

func (f *fakeExecutor) ExecWithLogLabel(provider, model, prompt, attachmentPath, logLabel string) (string, error) {
	f.calls = append(f.calls, execCall{
		provider: provider, model: model, prompt: prompt, attachmentPath: attachmentPath, logLabel: logLabel,
	})
	if len(f.responses) == 0 {
		return "", fmt.Errorf("unexpected call")
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

	require.Contains(t, prompt, "Your entire response is written directly to the transcription file")
	require.Contains(t, prompt, "--- BEGIN TRANSCRIPTION MARKDOWN DIALECT ---")
	require.Contains(t, prompt, "# Transcription Markdown dialect")
	require.Contains(t, prompt, "<!-- Running title: Des dritt buͤchs Euclidis -->")
	require.Contains(t, prompt, "Do not use `<!-- # Des dritt buͤchs Euclidis -->`")
	require.Contains(t, prompt, "<!-- Page: 86 -->")
	require.Contains(t, prompt, "<!-- Catchword: baid -->")
	require.Contains(t, prompt, `statements such as "the image confirms..."`)
	require.Contains(t, prompt, "text copied from the previous-page context")
}

func TestGeneratedMarkdownDialectMatchesSharedDocumentation(t *testing.T) {
	document, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "MARKDOWN_DIALECT.md"))
	require.NoError(t, err)
	require.Equal(t, string(document), markdownDialect,
		"run `go generate ./pkg/transcriptioncorrector` after editing docs/MARKDOWN_DIALECT.md")
}

func TestLineDiffCountsAddedAndDeletedLines(t *testing.T) {
	stats := lineDiff("one\ntwo\nthree\n", "one\nchanged\nthree\nfour\n")
	require.Equal(t, diffStats{added: 2, deleted: 1}, stats)
}

func TestRunCarriesPreviousPageAndPreviousRoundContext(t *testing.T) {
	root := t.TempDir()
	imagesDir := filepath.Join(root, "images")
	sourceOne := filepath.Join(root, "source-one")
	sourceTwo := filepath.Join(root, "source-two")
	altoSource := filepath.Join(root, "alto-source")
	outputDir := filepath.Join(root, "output")
	require.NoError(t, os.MkdirAll(imagesDir, 0o755))
	require.NoError(t, os.MkdirAll(altoSource, 0o755))
	for _, key := range []string{"page-0001", "page-0002"} {
		require.NoError(t, os.WriteFile(filepath.Join(imagesDir, key+".png"), []byte("image"), 0o644))
		for _, source := range []string{sourceOne, sourceTwo} {
			require.NoError(t, os.MkdirAll(filepath.Join(source, key), 0o755))
		}
		require.NoError(t, os.WriteFile(filepath.Join(sourceOne, key, "original.md"), []byte("first candidate "+key+"\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(sourceTwo, key, "original.md"), []byte("second candidate "+key+"\n"), 0o644))
		require.NoError(t, alto.SaveToFile(testALTO("ALTO", key), filepath.Join(altoSource, key+".xml")))
	}

	fake := &fakeExecutor{responses: []string{
		"round 1 page 1", "round 1 page 2", "round 2 page 1", "round 2 page 2",
	}}
	cfg := Config{
		MarkdownDirs: []string{sourceOne, sourceTwo}, ALTODirs: []string{altoSource}, ImagesDir: imagesDir, OutputDir: outputDir,
		Rounds: 2, Provider: llm.ProviderOllama, Model: "vision", Logger: log.New(io.Discard, "", 0),
	}
	require.NoError(t, Run(cfg, fake))
	require.Len(t, fake.calls, 4)
	require.NotContains(t, fake.calls[0].prompt, "BEGIN PREVIOUS PAGE")
	require.Contains(t, fake.calls[0].prompt, "ALTO page-0001")
	require.Contains(t, fake.calls[0].prompt, "converted with ALTOToMarkdown")
	require.Contains(t, fake.calls[1].prompt, "round 1 page 1")
	require.Contains(t, fake.calls[2].prompt, "correction from round 1")
	require.Contains(t, fake.calls[2].prompt, "round 1 page 1")
	require.NotContains(t, fake.calls[2].prompt, "round 1 page 2\n--- END PREVIOUS PAGE")
	require.Contains(t, fake.calls[3].prompt, "round 2 page 1")
	for _, call := range fake.calls {
		require.Equal(t, llm.ProviderOllama, call.provider)
		require.Equal(t, "vision", call.model)
		require.True(t, strings.HasSuffix(call.attachmentPath, ".png"))
	}

	finalOne, err := os.ReadFile(filepath.Join(outputDir, "page-0001", "original.md"))
	require.NoError(t, err)
	require.Equal(t, "round 2 page 1\n", string(finalOne))
	roundOne, err := os.ReadFile(filepath.Join(outputDir, "page-0002", "round-01.md"))
	require.NoError(t, err)
	require.Equal(t, "round 1 page 2\n", string(roundOne))
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
