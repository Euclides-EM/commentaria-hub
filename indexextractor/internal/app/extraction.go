package app

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

func runIndexExtraction(cfg config, client llmExecutor, out io.Writer) error {
	images, err := discoverImages(cfg.indexDir)
	if err != nil {
		return fmt.Errorf("discover index images: %w", err)
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return fmt.Errorf("select index images: %w", err)
	}
	manifest, err := loadIndexManifest(cfg.indexCSV, cfg.resume)
	if err != nil {
		return err
	}
	if !cfg.resume {
		if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
			return fmt.Errorf("initialize index manifest: %w", err)
		}
		if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
			return fmt.Errorf("initialize index CSV: %w", err)
		}
	}
	completed := indexCompleted(manifest)
	written, skipped := 0, 0
	for i, image := range selected {
		existing := findIndexPage(manifest.Pages, image.Path)
		if cfg.resume && !targeted && (completed[cleanPath(image.Path)] || (cfg.skipFailures && existing != nil && existing.Failure != nil)) {
			skipped++
			reason := "completed"
			if !completed[cleanPath(image.Path)] {
				reason = "failed"
			}
			fmt.Fprintf(out, "[%d/%d] Skipping %s index page %s\n", i+1, len(selected), reason, image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting index page %s\n", i+1, len(selected), image.Path)
		cached := ""
		if existing != nil && effectiveExtractionMode(cfg) == modeTwoPass {
			cached = existing.Transcription
		}
		entries, issues, transcription, err := extractIndexEntriesWithCheckpoint(cfg, client, image, cached, out, func(value string) error {
			manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{
				ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: modeTwoPass,
				Transcription: value, TranscriptionProvider: firstPassProvider(cfg), TranscriptionModel: firstPassModel(cfg), TranscribedAt: time.Now().UTC(),
			})
			if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
				return err
			}
			return renderIndexCSV(cfg.indexCSV, manifest)
		})
		if err != nil {
			var checkpointErr checkpointError
			if errors.As(err, &checkpointErr) {
				return err
			}
			failure := newPageFailure(cfg, transcription, err)
			page := findIndexPage(manifest.Pages, image.Path)
			if page == nil {
				manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: effectiveExtractionMode(cfg), Failure: failure})
			} else {
				page.Failure = failure
			}
			if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
				return fmt.Errorf("save index failure after %s: %w", image.Path, err)
			}
			if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
				return fmt.Errorf("render index CSV after failure %s: %w", image.Path, err)
			}
			fmt.Fprintf(out, "Warning: index extraction failed for %s during %s: %v; continuing\n", image.Path, failure.Phase, err)
			continue
		}
		for _, issue := range issues {
			fmt.Fprintf(out, "Warning: index response for %s: %s; skipping entry\n", image.Path, issue)
		}
		manifest.Pages = upsertIndexPage(manifest.Pages, indexPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: extractionProvider(cfg),
			Model: extractionModel(cfg), ExtractedAt: time.Now().UTC(), ExtractionMode: effectiveExtractionMode(cfg),
			Transcription: transcription, Entries: entries, Issues: issues,
		})
		if effectiveExtractionMode(cfg) == modeTwoPass {
			page := findIndexPage(manifest.Pages, image.Path)
			page.TranscriptionProvider, page.TranscriptionModel = firstPassProvider(cfg), firstPassModel(cfg)
			if page.TranscribedAt.IsZero() {
				page.TranscribedAt = time.Now().UTC()
			}
		}
		if err := saveJSONAtomically(manifestPath(cfg.indexCSV), manifest); err != nil {
			return fmt.Errorf("save index manifest after %s: %w", image.Path, err)
		}
		if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
			return fmt.Errorf("render index CSV after %s: %w", image.Path, err)
		}
		written += len(entries)
	}
	if err := renderIndexCSV(cfg.indexCSV, manifest); err != nil {
		return fmt.Errorf("render index CSV: %w", err)
	}
	fmt.Fprintf(out, "Index extraction complete: %d rows extracted, %d images skipped; manifest %s; output %s\n", written, skipped, manifestPath(cfg.indexCSV), cfg.indexCSV)
	return nil
}

func extractIndexEntries(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, error) {
	entries, _, err := extractIndexEntriesWithIssues(cfg, client, image, out)
	return entries, err
}

func extractIndexEntriesWithIssues(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, []string, error) {
	entries, issues, _, err := extractIndexEntriesWithAudit(cfg, client, image, out)
	return entries, issues, err
}

func extractIndexEntriesWithAudit(cfg config, client llmExecutor, image imageInput, out io.Writer) ([]indexEntry, []string, string, error) {
	return extractIndexEntriesWithCheckpoint(cfg, client, image, "", out, nil)
}

func extractIndexEntriesWithCheckpoint(cfg config, client llmExecutor, image imageInput, cached string, out io.Writer, checkpoint func(string) error) ([]indexEntry, []string, string, error) {
	transcription, prompt, attachment, err := extractionInput(cfg, client, image, indexPrompt, cached, checkpoint)
	if err != nil {
		return nil, nil, transcription, fmt.Errorf("transcribe index image %s: %w", image.Path, err)
	}
	var parseErr error
	for attempt := 1; attempt <= responseValidationAttempts; attempt++ {
		raw, err := client.Exec(extractionProvider(cfg), extractionModel(cfg), prompt, attachment)
		if err != nil {
			return nil, nil, transcription, fmt.Errorf("extract index image %s: %w", image.Path, err)
		}
		entries, issues, err := parseIndexResponseWithIssues(raw, image.Volume)
		if err == nil {
			return entries, issues, transcription, nil
		}
		parseErr = err
		if attempt < responseValidationAttempts {
			fmt.Fprintf(out, "Invalid index response for %s (%v); retrying extraction (%d/%d)\n", image.Path, err, attempt+1, responseValidationAttempts)
		}
	}
	return nil, nil, transcription, fmt.Errorf("parse index response for %s after %d attempts: %w", image.Path, responseValidationAttempts, parseErr)
}

func runLettersExtraction(cfg config, client llmExecutor, out io.Writer) error {
	images, err := discoverImages(cfg.lettersDir)
	if err != nil {
		return fmt.Errorf("discover letters-table images: %w", err)
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return fmt.Errorf("select letters-table images: %w", err)
	}
	manifest, err := loadLettersManifest(cfg.lettersCSV, cfg.resume)
	if err != nil {
		return err
	}
	if !cfg.resume {
		if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
			return fmt.Errorf("initialize letters-table manifest: %w", err)
		}
		if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
			return fmt.Errorf("initialize letters-table CSV: %w", err)
		}
	}
	completed := lettersCompleted(manifest)
	written, skipped := 0, 0
	for i, image := range selected {
		existing := findLettersPage(manifest.Pages, image.Path)
		if cfg.resume && !targeted && (completed[cleanPath(image.Path)] || (cfg.skipFailures && existing != nil && existing.Failure != nil)) {
			skipped++
			reason := "completed"
			if !completed[cleanPath(image.Path)] {
				reason = "failed"
			}
			fmt.Fprintf(out, "[%d/%d] Skipping %s letters-table page %s\n", i+1, len(selected), reason, image.Path)
			continue
		}
		fmt.Fprintf(out, "[%d/%d] Extracting letters-table page %s\n", i+1, len(selected), image.Path)
		cached := ""
		if existing != nil && effectiveExtractionMode(cfg) == modeTwoPass {
			cached = existing.Transcription
		}
		transcription, prompt, attachment, err := extractionInput(cfg, client, image, lettersTablePrompt, cached, func(value string) error {
			manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{
				ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: modeTwoPass,
				Transcription: value, TranscriptionProvider: firstPassProvider(cfg), TranscriptionModel: firstPassModel(cfg), TranscribedAt: time.Now().UTC(),
			})
			if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
				return err
			}
			return renderLettersCSV(cfg.lettersCSV, manifest)
		})
		if err != nil {
			var checkpointErr checkpointError
			if errors.As(err, &checkpointErr) {
				return err
			}
			err = fmt.Errorf("transcribe letters-table image %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		raw, err := client.Exec(extractionProvider(cfg), extractionModel(cfg), prompt, attachment)
		if err != nil {
			err = fmt.Errorf("extract letters-table image %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		entries, issues, err := parseLettersResponseWithIssues(raw, image.Volume)
		if err != nil {
			err = fmt.Errorf("parse letters-table response for %s: %w", image.Path, err)
			if persistErr := recordLettersFailure(cfg, &manifest, image, transcription, err); persistErr != nil {
				return persistErr
			}
			fmt.Fprintf(out, "Warning: letters-table extraction failed for %s during %s: %v; continuing\n", image.Path, newPageFailure(cfg, transcription, err).Phase, err)
			continue
		}
		for _, issue := range issues {
			fmt.Fprintf(out, "Warning: letters-table response for %s: %s; skipping entry\n", image.Path, issue)
		}
		manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{
			ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: extractionProvider(cfg),
			Model: extractionModel(cfg), ExtractedAt: time.Now().UTC(), ExtractionMode: effectiveExtractionMode(cfg),
			Transcription: transcription, Entries: entries, Issues: issues,
		})
		if effectiveExtractionMode(cfg) == modeTwoPass {
			page := findLettersPage(manifest.Pages, image.Path)
			page.TranscriptionProvider, page.TranscriptionModel = firstPassProvider(cfg), firstPassModel(cfg)
			if page.TranscribedAt.IsZero() {
				page.TranscribedAt = time.Now().UTC()
			}
		}
		if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), manifest); err != nil {
			return fmt.Errorf("save letters-table manifest after %s: %w", image.Path, err)
		}
		if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
			return fmt.Errorf("render letters-table CSV after %s: %w", image.Path, err)
		}
		written += len(entries)
	}
	if err := renderLettersCSV(cfg.lettersCSV, manifest); err != nil {
		return fmt.Errorf("render letters-table CSV: %w", err)
	}
	fmt.Fprintf(out, "Letters-table extraction complete: %d rows extracted, %d images skipped; manifest %s; output %s\n", written, skipped, manifestPath(cfg.lettersCSV), cfg.lettersCSV)
	return nil
}

func effectiveExtractionMode(cfg config) string {
	if cfg.extractionMode == modeTwoPass {
		return modeTwoPass
	}
	return modeOnePass
}

func extractionInput(cfg config, client llmExecutor, image imageInput, structuredPrompt, cached string, checkpoint func(string) error) (transcription, prompt, attachment string, err error) {
	if effectiveExtractionMode(cfg) == modeOnePass {
		return "", structuredPrompt, image.Path, nil
	}
	transcription = cached
	if strings.TrimSpace(transcription) == "" {
		transcription, err = client.Exec(firstPassProvider(cfg), firstPassModel(cfg), transcriptionPrompt, image.Path)
		if err != nil {
			return "", "", "", err
		}
		if strings.TrimSpace(transcription) == "" {
			return "", "", "", errors.New("LLM returned an empty transcription")
		}
		if checkpoint != nil {
			if err := checkpoint(transcription); err != nil {
				return transcription, "", "", checkpointError{fmt.Errorf("checkpoint transcription: %w", err)}
			}
		}
	}
	prompt = structuredPrompt + "\n\nParse the following transcription. It is the only source; do not infer text that is absent. Markdown **bold** markers indicate bold print.\n\n<transcription>\n" + transcription + "\n</transcription>"
	return transcription, prompt, "", nil
}

func newPageFailure(cfg config, transcription string, err error) *pageFailure {
	phase, provider, model := failurePhaseSingle, extractionProvider(cfg), extractionModel(cfg)
	if effectiveExtractionMode(cfg) == modeTwoPass {
		phase, provider, model = failurePhaseFirst, firstPassProvider(cfg), firstPassModel(cfg)
		if strings.TrimSpace(transcription) != "" {
			phase, provider, model = failurePhaseSecond, secondPassProvider(cfg), secondPassModel(cfg)
		}
	}
	return &pageFailure{Phase: phase, Provider: provider, Model: model, Error: err.Error(), FailedAt: time.Now().UTC()}
}

func recordLettersFailure(cfg config, manifest *lettersManifest, image imageInput, transcription string, failureErr error) error {
	failure := newPageFailure(cfg, transcription, failureErr)
	page := findLettersPage(manifest.Pages, image.Path)
	if page == nil {
		manifest.Pages = upsertLettersPage(manifest.Pages, lettersPageResult{ImagePath: cleanPath(image.Path), Volume: image.Volume, ExtractionMode: effectiveExtractionMode(cfg), Failure: failure})
	} else {
		page.Failure = failure
	}
	if err := saveJSONAtomically(manifestPath(cfg.lettersCSV), *manifest); err != nil {
		return fmt.Errorf("save letters-table failure after %s: %w", image.Path, err)
	}
	if err := renderLettersCSV(cfg.lettersCSV, *manifest); err != nil {
		return fmt.Errorf("render letters-table CSV after failure %s: %w", image.Path, err)
	}
	return nil
}
