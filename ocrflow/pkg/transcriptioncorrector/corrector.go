// Package transcriptioncorrector corrects page-oriented transcriptions with a
// multimodal LLM. It accepts Markdown and ALTO sources and preserves each
// correction round in the repository's page-NNNN Markdown layout.
package transcriptioncorrector

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"
)

const DefaultRounds = 2

// Config describes one transcription correction run.
type Config struct {
	MarkdownDirs      []string
	ALTODirs          []string
	TranscriptionDirs []string
	PageKeys          []string
	ImagesDir         string
	OutputDir         string
	Rounds            int
	Provider          string
	Model             string
	Logger            *log.Logger
}

// Executor is the subset of the shared LLM client used by the corrector.
type Executor interface {
	ExecWithLogLabel(provider, model, prompt, attachmentPath, logLabel string) (string, error)
}

type page struct {
	key       string
	imagePath string
}

type candidate struct {
	label string
	text  string
}

// Run executes all configured rounds sequentially. Pages are sequential
// because each request receives the corrected output of the preceding page.
func Run(cfg Config, client Executor) error {
	if client == nil {
		return errors.New("LLM executor is required")
	}
	if cfg.Rounds == 0 {
		cfg.Rounds = DefaultRounds
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}
	pages, err := discoverPages(cfg.ImagesDir)
	if err != nil {
		return err
	}
	pages, err = selectPages(pages, cfg.PageKeys)
	if err != nil {
		return err
	}
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	logger.Printf("start pages=%d rounds=%d markdown_sources=%d alto_sources=%d transcription_sources=%d provider=%s model=%s images=%s output=%s",
		len(pages), cfg.Rounds, len(cfg.MarkdownDirs), len(cfg.ALTODirs), len(cfg.TranscriptionDirs), cfg.Provider, cfg.Model, cfg.ImagesDir, cfg.OutputDir)
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
	for round := 1; round <= cfg.Rounds; round++ {
		roundStarted := time.Now()
		var previousPage string
		var roundAdded, roundDeleted int
		logger.Printf("round start round=%d/%d pages=%d", round, cfg.Rounds, len(pages))

		for pageIndex, p := range pages {
			pageStarted := time.Now()
			base, err := loadCandidates(cfg.MarkdownDirs, cfg.ALTODirs, cfg.TranscriptionDirs, p.key)
			if err != nil {
				return fmt.Errorf("round %d page %s: %w", round, p.key, err)
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
			response, err := client.ExecWithLogLabel(cfg.Provider, cfg.Model, prompt, p.imagePath, fmt.Sprintf("round=%d page=%s", round, p.key))
			if err != nil {
				return fmt.Errorf("round %d page %s LLM correction failed: %w", round, p.key, err)
			}
			corrected, err := normalizeResponse(response)
			if err != nil {
				return fmt.Errorf("round %d page %s invalid LLM response: %w", round, p.key, err)
			}

			for _, c := range promptCandidates {
				stats := lineDiff(c.text, corrected)
				roundAdded += stats.added
				roundDeleted += stats.deleted
				logger.Printf("page diff round=%d key=%s against=%q added_lines=%d deleted_lines=%d total_changes=%d",
					round, p.key, c.label, stats.added, stats.deleted, stats.added+stats.deleted)
			}

			roundPath := filepath.Join(cfg.OutputDir, p.key, fmt.Sprintf("round-%02d.md", round))
			if err := writeFileAtomic(roundPath, []byte(corrected)); err != nil {
				return fmt.Errorf("write round output for %s: %w", p.key, err)
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
			return fmt.Errorf("write final output for %s: %w", p.key, err)
		}
	}
	logger.Printf("complete pages=%d rounds=%d final_outputs=%s/page-NNNN/original.md", len(pages), cfg.Rounds, cfg.OutputDir)
	return nil
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
	if cfg.Rounds < 1 {
		return errors.New("rounds must be at least 1")
	}
	if strings.TrimSpace(cfg.Provider) == "" {
		return errors.New("LLM provider is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return errors.New("LLM model is required")
	}
	return validateDirectories(cfg)
}
