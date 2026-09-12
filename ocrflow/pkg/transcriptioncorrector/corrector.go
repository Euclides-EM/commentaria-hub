// Package transcriptioncorrector corrects page-oriented transcriptions with a
// multimodal LLM. It accepts Markdown and ALTO sources and preserves each
// correction round in the repository's page-NNNN Markdown layout.
package transcriptioncorrector

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
)

const DefaultRounds = 1

type ExecutionMode string

const (
	ExecutionModePageByPage ExecutionMode = "page_by_page"
	ExecutionModeDirectory  ExecutionMode = "directory"
	DefaultExecutionMode                  = ExecutionModePageByPage
)

// Config describes one transcription correction run.
type Config struct {
	MarkdownDirs      []string
	ALTODirs          []string
	TranscriptionDirs []string
	PageKeys          []string
	ImagesDir         string
	OutputDir         string
	Rounds            int
	// ExecutionMode defaults to page_by_page. Directory mode invokes one local
	// CLI agent with absolute paths and ignores Rounds.
	ExecutionMode ExecutionMode
	SkipExisting  bool
	Provider      string
	Model         string
	Logger        *log.Logger
}

// Executor is the subset of the shared LLM client used by the corrector.
type Executor interface {
	ExecPromptResultWithLogLabel(provider, model string, prompt llm.Prompt, attachmentPath, logLabel string) (llm.Result, error)
}

type WorkspaceExecutor interface {
	ExecWorkspaceResultWithLogLabel(provider, model string, prompt llm.Prompt, workingDir string, readPaths []string, logLabel string) (llm.Result, error)
}

type page struct {
	key       string
	imagePath string
}

type candidate struct {
	label string
	text  string
}

// Run executes the correction and returns aggregate provider usage
// across every successfully completed page and round.
func Run(cfg Config, client Executor) (totalUsage llm.Usage, err error) {
	if client == nil {
		return totalUsage, errors.New("LLM executor is required")
	}
	if cfg.ExecutionMode == "" {
		cfg.ExecutionMode = DefaultExecutionMode
	}
	if cfg.ExecutionMode == ExecutionModePageByPage && cfg.Rounds == 0 {
		cfg.Rounds = DefaultRounds
	}
	if err := validateConfig(cfg); err != nil {
		return totalUsage, err
	}
	pages, err := discoverPages(cfg.ImagesDir)
	if err != nil {
		return totalUsage, err
	}
	pages, err = selectPages(pages, cfg.PageKeys)
	if err != nil {
		return totalUsage, err
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}
	if cfg.ExecutionMode == ExecutionModeDirectory {
		workspaceClient, ok := client.(WorkspaceExecutor)
		if !ok {
			return totalUsage, errors.New("LLM executor does not support directory execution")
		}
		return runDirectory(cfg, pages, workspaceClient, logger)
	}

	logger.Printf("start mode=%s pages=%d rounds=%d markdown_sources=%d alto_sources=%d transcription_sources=%d provider=%s model=%s images=%s output=%s",
		cfg.ExecutionMode, len(pages), cfg.Rounds, len(cfg.MarkdownDirs), len(cfg.ALTODirs), len(cfg.TranscriptionDirs), cfg.Provider, cfg.Model, cfg.ImagesDir, cfg.OutputDir)
	for i, dir := range cfg.MarkdownDirs {
		logger.Printf("markdown source=%d path=%s", i+1, dir)
	}
	for i, dir := range cfg.ALTODirs {
		logger.Printf("ALTO source=%d path=%s conversion=ALTOToMarkdown", i+1, dir)
	}
	for i, dir := range cfg.TranscriptionDirs {
		logger.Printf("mixed transcription source=%d path=%s formats=ALTO/Markdown", i+1, dir)
	}

	previousRound := make(map[string]string, len(pages))
	requestCount := 0
	costReportCount := 0
	cacheHealth := newCacheHealthTracker(cfg.Provider, cfg.Model)
	defer func() {
		status := "complete"
		if err != nil {
			status = "failed"
		}
		cost := "unavailable"
		if requestCount > 0 && costReportCount == requestCount && totalUsage.CostUSD != nil {
			cost = fmt.Sprintf("%.6f", *totalUsage.CostUSD)
		} else {
			totalUsage.CostUSD = nil
		}
		logger.Printf(
			"%s pages=%d rounds=%d requests=%d tokens_input=%d tokens_cached=%d tokens_cache_creation=%d tokens_output=%d tokens_reasoning=%d tokens_total=%d cost_usd=%s cost_reports=%d/%d final_outputs=%s/page-NNNN/original.md",
			status, len(pages), cfg.Rounds, requestCount,
			totalUsage.InputTokens, totalUsage.CachedInputTokens, totalUsage.CacheCreationInputTokens,
			totalUsage.OutputTokens, totalUsage.ReasoningTokens, totalUsage.TotalTokens,
			cost, costReportCount, requestCount, cfg.OutputDir,
		)
		cacheHealth.LogSummary(logger)
	}()
	for round := 1; round <= cfg.Rounds; round++ {
		roundStarted := time.Now()
		var previousPage string
		var roundAdded, roundDeleted int
		logger.Printf("round start round=%d/%d pages=%d", round, cfg.Rounds, len(pages))

		for pageIndex, p := range pages {
			pageStarted := time.Now()
			roundPath := filepath.Join(cfg.OutputDir, p.key, fmt.Sprintf("round-%02d.md", round))
			if cfg.SkipExisting {
				existing, err := os.ReadFile(roundPath)
				if err == nil {
					previousRound[p.key] = string(existing)
					previousPage = string(existing)
					logger.Printf("page skipped round=%d/%d page=%d/%d key=%s output=%s", round, cfg.Rounds, pageIndex+1, len(pages), p.key, roundPath)
					continue
				}
				if !errors.Is(err, os.ErrNotExist) {
					return totalUsage, fmt.Errorf("read existing round output for %s: %w", p.key, err)
				}
			}
			base, err := loadCandidates(cfg.MarkdownDirs, cfg.ALTODirs, cfg.TranscriptionDirs, p.key)
			if err != nil {
				return totalUsage, fmt.Errorf("round %d page %s: %w", round, p.key, err)
			}
			promptCandidates := append([]candidate(nil), base...)
			if draft, ok := previousRound[p.key]; ok {
				promptCandidates = append(promptCandidates, candidate{label: fmt.Sprintf("correction from round %d", round-1), text: draft})
			}
			logger.Printf("page start round=%d/%d page=%d/%d key=%s image=%s candidates=%d previous_page=%t",
				round, cfg.Rounds, pageIndex+1, len(pages), p.key, p.imagePath, len(promptCandidates), previousPage != "")
			for _, c := range promptCandidates {
				logger.Printf("page input round=%d key=%s source=%q bytes=%d lines=%d", round, p.key, c.label, len(c.text), lineCount(c.text))
			}

			prompt := buildPrompt(p.key, round, cfg.Rounds, promptCandidates, previousPage)
			result, err := client.ExecPromptResultWithLogLabel(cfg.Provider, cfg.Model, prompt, p.imagePath, fmt.Sprintf("round=%d page=%s", round, p.key))
			if err != nil {
				return totalUsage, fmt.Errorf("round %d page %s LLM correction failed: %w", round, p.key, err)
			}
			requestCount++
			totalUsage.Add(result.Usage)
			cacheHealth.Observe(result.Usage, logger)
			if result.Usage.CostUSD != nil {
				costReportCount++
			}
			corrected, err := normalizeResponse(result.Text)
			if err != nil {
				return totalUsage, fmt.Errorf("round %d page %s invalid LLM response: %w", round, p.key, err)
			}

			for _, c := range promptCandidates {
				stats := lineDiff(c.text, corrected)
				roundAdded += stats.added
				roundDeleted += stats.deleted
				logger.Printf("page diff round=%d key=%s against=%q added_lines=%d deleted_lines=%d total_changes=%d",
					round, p.key, c.label, stats.added, stats.deleted, stats.added+stats.deleted)
			}

			if err := writeFileAtomic(roundPath, []byte(corrected)); err != nil {
				return totalUsage, fmt.Errorf("write round output for %s: %w", p.key, err)
			}
			previousRound[p.key] = corrected
			previousPage = corrected
			logger.Printf("page complete round=%d/%d page=%d/%d key=%s bytes=%d lines=%d duration=%s output=%s",
				round, cfg.Rounds, pageIndex+1, len(pages), p.key, len(corrected), lineCount(corrected), time.Since(pageStarted).Round(time.Millisecond), roundPath)
		}
		logger.Printf("round complete round=%d/%d pages=%d added_lines=%d deleted_lines=%d duration=%s",
			round, cfg.Rounds, len(pages), roundAdded, roundDeleted, time.Since(roundStarted).Round(time.Millisecond))
	}

	for _, p := range pages {
		finalPath := filepath.Join(cfg.OutputDir, p.key, "original.md")
		if err := writeFileAtomic(finalPath, []byte(previousRound[p.key])); err != nil {
			return totalUsage, fmt.Errorf("write final output for %s: %w", p.key, err)
		}
	}
	return totalUsage, nil
}

func validateConfig(cfg Config) error {
	if len(cfg.MarkdownDirs)+len(cfg.ALTODirs)+len(cfg.TranscriptionDirs) == 0 {
		return errors.New("at least one Markdown or ALTO input directory is required")
	}
	if strings.TrimSpace(cfg.ImagesDir) == "" {
		return errors.New("images directory is required")
	}
	if strings.TrimSpace(cfg.OutputDir) == "" {
		return errors.New("output directory is required")
	}
	if cfg.ExecutionMode != ExecutionModePageByPage && cfg.ExecutionMode != ExecutionModeDirectory {
		return fmt.Errorf("unsupported execution mode %q (use %s or %s)", cfg.ExecutionMode, ExecutionModePageByPage, ExecutionModeDirectory)
	}
	if cfg.ExecutionMode == ExecutionModePageByPage && cfg.Rounds < 1 {
		return errors.New("rounds must be at least 1")
	}
	if cfg.ExecutionMode == ExecutionModeDirectory && cfg.Provider != llm.ProviderClaudeCode && cfg.Provider != llm.ProviderCodex {
		return fmt.Errorf("directory execution requires a local CLI provider (use %s or %s)", llm.ProviderClaudeCode, llm.ProviderCodex)
	}
	if strings.TrimSpace(cfg.Provider) == "" {
		return errors.New("LLM provider is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return errors.New("LLM model is required")
	}
	return validateDirectories(cfg)
}

func runDirectory(cfg Config, pages []page, client WorkspaceExecutor, logger *log.Logger) (llm.Usage, error) {
	selected := make([]page, 0, len(pages))
	for _, p := range pages {
		finalPath := filepath.Join(cfg.OutputDir, p.key, "original.md")
		if cfg.SkipExisting {
			if info, err := os.Stat(finalPath); err == nil && !info.IsDir() {
				logger.Printf("page skipped mode=%s key=%s output=%s", cfg.ExecutionMode, p.key, finalPath)
				continue
			} else if err != nil && !errors.Is(err, os.ErrNotExist) {
				return llm.Usage{}, fmt.Errorf("access existing output for %s: %w", p.key, err)
			}
		}
		selected = append(selected, p)
	}
	if len(selected) == 0 {
		return llm.Usage{}, nil
	}

	prompt, readPaths, err := buildDirectoryPrompt(cfg, selected)
	if err != nil {
		return llm.Usage{}, err
	}
	logger.Printf("start mode=%s pages=%d rounds=ignored markdown_sources=%d alto_sources=%d transcription_sources=%d provider=%s model=%s images=%s output=%s",
		cfg.ExecutionMode, len(selected), len(cfg.MarkdownDirs), len(cfg.ALTODirs), len(cfg.TranscriptionDirs), cfg.Provider, cfg.Model, cfg.ImagesDir, cfg.OutputDir)
	result, err := client.ExecWorkspaceResultWithLogLabel(cfg.Provider, cfg.Model, prompt, cfg.OutputDir, readPaths, "mode=directory")
	if err != nil {
		return llm.Usage{}, fmt.Errorf("directory LLM correction failed: %w", err)
	}
	logDirectoryUsage(logger, cfg, len(selected), result.Usage)
	for _, p := range selected {
		finalPath := filepath.Join(cfg.OutputDir, p.key, "original.md")
		contents, err := os.ReadFile(finalPath)
		if err != nil {
			return result.Usage, fmt.Errorf("directory LLM correction did not produce %s: %w", finalPath, err)
		}
		normalized, err := normalizeResponse(string(contents))
		if err != nil {
			return result.Usage, fmt.Errorf("directory LLM correction produced invalid output for %s: %w", p.key, err)
		}
		if err := writeFileAtomic(finalPath, []byte(normalized)); err != nil {
			return result.Usage, fmt.Errorf("normalize directory output for %s: %w", p.key, err)
		}
	}
	return result.Usage, nil
}

func logDirectoryUsage(logger *log.Logger, cfg Config, pageCount int, usage llm.Usage) {
	cost := "unavailable"
	if usage.CostUSD != nil {
		cost = fmt.Sprintf("%.6f", *usage.CostUSD)
	}
	logger.Printf("provider complete mode=%s pages=%d requests=1 tokens_input=%d tokens_cached=%d tokens_cache_creation=%d tokens_output=%d tokens_reasoning=%d tokens_total=%d cost_usd=%s output=%s/page-NNNN/original.md",
		cfg.ExecutionMode, pageCount, usage.InputTokens, usage.CachedInputTokens, usage.CacheCreationInputTokens, usage.OutputTokens, usage.ReasoningTokens, usage.TotalTokens, cost, cfg.OutputDir)
}
