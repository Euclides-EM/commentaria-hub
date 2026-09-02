package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
)

type Config struct {
	From       string
	To         string
	InputPath  string
	OutputPath string
	PageRange  string
	DPI        float64
}

type conversion struct {
	from string
	to   string
}

func run(config Config) error {
	config.From = strings.ToLower(strings.TrimSpace(config.From))
	config.To = strings.ToLower(strings.TrimSpace(config.To))
	config.InputPath = strings.TrimSpace(config.InputPath)
	config.OutputPath = strings.TrimSpace(config.OutputPath)
	config.PageRange = strings.TrimSpace(config.PageRange)

	if config.From == "" {
		return errors.New("missing -from")
	}
	if config.To == "" {
		return errors.New("missing -to")
	}
	if config.InputPath == "" {
		return errors.New("missing -input")
	}
	if config.OutputPath == "" {
		return errors.New("missing -output")
	}
	if config.DPI <= 0 {
		return fmt.Errorf("dpi must be positive, got %v", config.DPI)
	}

	workflow := conversion{from: config.From, to: config.To}
	if config.PageRange != "" && workflow != (conversion{from: "pdf", to: "png"}) {
		return errors.New("-range is only supported for pdf -> png")
	}

	switch workflow {
	case conversion{from: "pdf", to: "png"}:
		if err := requireFileExtension(config.InputPath, ".pdf"); err != nil {
			return err
		}
		return convertToProcessedPNGs(config)
	case conversion{from: "image", to: "png"}:
		if err := requireImageInput(config.InputPath); err != nil {
			return err
		}
		return convertToProcessedPNGs(config)
	case conversion{from: "image", to: "pdf"}:
		return imageDirToPDF(config.InputPath, config.OutputPath)
	case conversion{from: "pagexml", to: "alto"}:
		return formatcov.PageXMLFilesToALTO(config.InputPath, config.OutputPath)
	case conversion{from: "alto", to: "markdown"}:
		return formatcov.ALTOFilesToMarkdown(config.InputPath, config.OutputPath)
	default:
		return fmt.Errorf("unsupported conversion %q -> %q\n%s", config.From, config.To, supportedWorkflows)
	}
}

func convertToProcessedPNGs(config Config) error {
	if err := os.MkdirAll(config.OutputPath, 0o755); err != nil {
		return fmt.Errorf("create output directory %q: %w", config.OutputPath, err)
	}

	var pages []int
	if config.PageRange != "" {
		var err error
		pages, err = pagesparser.IntRange(config.PageRange)
		if err != nil {
			return fmt.Errorf("invalid range %q: %w", config.PageRange, err)
		}
	}

	rawDir, err := futils.MkdirTemp("formatcov-raw")
	if err != nil {
		return fmt.Errorf("create raw temp directory: %w", err)
	}
	defer os.RemoveAll(rawDir)

	denoiseDir, err := futils.MkdirTemp("formatcov-denoise")
	if err != nil {
		return fmt.Errorf("create denoise temp directory: %w", err)
	}
	defer os.RemoveAll(denoiseDir)

	if err := prepareInputPNGs(config.InputPath, rawDir, config.DPI, pages); err != nil {
		return err
	}
	if err := formatcov.DenoisePNGs(rawDir, denoiseDir); err != nil {
		return err
	}
	return formatcov.DeskewPNGs(denoiseDir, config.OutputPath)
}

func requireFileExtension(path, extension string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat input %q: %w", path, err)
	}
	if info.IsDir() || !strings.EqualFold(filepath.Ext(path), extension) {
		return fmt.Errorf("input %q must be a %s file", path, extension)
	}
	return nil
}

func requireImageInput(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat input %q: %w", path, err)
	}
	if info.IsDir() {
		return nil
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif":
		return nil
	default:
		return fmt.Errorf("input %q must be a PNG, JPG, JPEG, or GIF image, or a flat image directory", path)
	}
}
