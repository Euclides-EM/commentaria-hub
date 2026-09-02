package main

import (
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/stretchr/testify/require"
)

func TestParseFlagsAcceptsRepeatedSourcesAndCredentials(t *testing.T) {
	cfg, err := parseFlags([]string{
		"-markdown-dir", "first",
		"-alto-dir", "alto-one",
		"-alto-dir", "alto-two",
		"-images-dir", "images",
		"-output-dir", "out",
		"-rounds", "3",
		"-ai-provider", "OpenAI",
		"-ai-model", "gpt-test",
		"-auth-token", "secret",
		"second",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"first", "second"}, cfg.corrector.MarkdownDirs)
	require.Equal(t, []string{"alto-one", "alto-two"}, cfg.corrector.ALTODirs)
	require.Equal(t, 3, cfg.corrector.Rounds)
	require.Equal(t, llm.ProviderOpenAI, cfg.corrector.Provider)
	require.Equal(t, "secret", cfg.openAIAPIKey)
}

func TestParseFlagsAcceptsALTOWithoutMarkdown(t *testing.T) {
	cfg, err := parseFlags([]string{
		"-alto-dir", "alto",
		"-images-dir", "images",
		"-output-dir", "out",
		"-ai-provider", "claude-code",
		"-ai-model", "sonnet",
	})
	require.NoError(t, err)
	require.Empty(t, cfg.corrector.MarkdownDirs)
	require.Equal(t, []string{"alto"}, cfg.corrector.ALTODirs)
}

func TestParseFlagsValidatesRequiredValues(t *testing.T) {
	_, err := parseFlags([]string{"-images-dir", "images", "-output-dir", "out", "-rounds", "0", "-ai-provider", "ollama", "-ai-model", "vision"})
	require.ErrorContains(t, err, "markdown-dir")

	_, err = parseFlags([]string{"-images-dir", "images", "-output-dir", "out", "-rounds", "0", "-ai-provider", "ollama", "-ai-model", "vision", "source"})
	require.ErrorContains(t, err, "rounds")
}
