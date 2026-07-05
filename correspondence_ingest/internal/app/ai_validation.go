package app

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const validationManifestVersion = 1

const (
	validationTranscription = "transcriptions"
	validationParsing       = "parsing"
)

type validationDifference struct {
	Location string `json:"location"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Reason   string `json:"reason"`
}

type validationResponse struct {
	Accurate    bool                   `json:"accurate"`
	Summary     string                 `json:"summary"`
	Differences []validationDifference `json:"differences"`
}

type validationResult struct {
	ImagePath    string                 `json:"image_path"`
	Volume       string                 `json:"volume"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	ValidatedAt  time.Time              `json:"validated_at"`
	SourceDigest string                 `json:"source_digest"`
	Accurate     bool                   `json:"accurate"`
	Summary      string                 `json:"summary"`
	Differences  []validationDifference `json:"differences"`
	Failure      string                 `json:"failure,omitempty"`
	RawResponse  string                 `json:"raw_response,omitempty"`
}

type validationManifest struct {
	Version int                `json:"version"`
	Kind    string             `json:"kind"`
	Check   string             `json:"check"`
	Results []validationResult `json:"results"`
}

func validationStatePath(output, check string) string {
	return validationArtifactPath(output, ".validate-"+check+".json")
}

func validationReviewPath(output, check string) string {
	return validationArtifactPath(output, ".validate-"+check+".md")
}

func validationArtifactPath(output, suffix string) string {
	outputDir := filepath.Dir(output)
	if filepath.Base(outputDir) == "outputs" {
		return filepath.Join(filepath.Dir(outputDir), "reviews", filepath.Base(output)+suffix)
	}
	return output + suffix
}

func runAIValidation(cfg config, client llmExecutor, check string, out io.Writer) error {
	if client == nil {
		return errors.New("AI validation requires an LLM client")
	}
	if includesKind(cfg.kind, kindIndex) {
		if err := validateIndexWithAI(cfg, client, check, out); err != nil {
			return err
		}
	}
	if includesKind(cfg.kind, kindLetters) {
		if err := validateLettersWithAI(cfg, client, check, out); err != nil {
			return err
		}
	}
	return nil
}

func validateIndexWithAI(cfg config, client llmExecutor, check string, out io.Writer) error {
	manifest, err := loadIndexManifest(cfg.indexCSV, true)
	if err != nil {
		return err
	}
	images, err := discoverImages(cfg.indexDir)
	if err != nil {
		return err
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return err
	}
	state, err := loadValidationManifest(cfg.indexCSV, kindIndex, check, cfg.resume)
	if err != nil {
		return err
	}
	for i, image := range selected {
		page := findIndexPage(manifest.Pages, image.Path)
		if page == nil || page.Entries == nil || strings.TrimSpace(page.Transcription) == "" {
			fmt.Fprintf(out, "[%d/%d] Skipping index page without completed two-pass data %s\n", i+1, len(selected), image.Path)
			continue
		}
		structured, _ := json.MarshalIndent(page.Entries, "", "  ")
		digest := validationSourceDigest(check, page.Transcription, string(structured))
		if shouldSkipValidation(state, image.Path, digest, cfg, targeted) {
			fmt.Fprintf(out, "[%d/%d] Skipping validated index page %s\n", i+1, len(selected), image.Path)
			continue
		}
		result := executeAIValidation(cfg, client, check, image, page.Transcription, string(structured), digest)
		state.Results = upsertValidationResult(state.Results, result)
		if err := saveValidationCheckpoint(cfg.indexCSV, state); err != nil {
			return err
		}
		printValidationResult(out, i+1, len(selected), "index", result)
	}
	return finishValidation(cfg.indexCSV, state, out)
}

func validateLettersWithAI(cfg config, client llmExecutor, check string, out io.Writer) error {
	manifest, err := loadLettersManifest(cfg.lettersCSV, true)
	if err != nil {
		return err
	}
	images, err := discoverImages(cfg.lettersDir)
	if err != nil {
		return err
	}
	selected, targeted, err := selectImages(images, cfg.rerunImages)
	if err != nil {
		return err
	}
	state, err := loadValidationManifest(cfg.lettersCSV, kindLetters, check, cfg.resume)
	if err != nil {
		return err
	}
	for i, image := range selected {
		page := findLettersPage(manifest.Pages, image.Path)
		if page == nil || page.Entries == nil || strings.TrimSpace(page.Transcription) == "" {
			fmt.Fprintf(out, "[%d/%d] Skipping letters page without completed two-pass data %s\n", i+1, len(selected), image.Path)
			continue
		}
		structured, _ := json.MarshalIndent(page.Entries, "", "  ")
		digest := validationSourceDigest(check, page.Transcription, string(structured))
		if shouldSkipValidation(state, image.Path, digest, cfg, targeted) {
			fmt.Fprintf(out, "[%d/%d] Skipping validated letters page %s\n", i+1, len(selected), image.Path)
			continue
		}
		result := executeAIValidation(cfg, client, check, image, page.Transcription, string(structured), digest)
		state.Results = upsertValidationResult(state.Results, result)
		if err := saveValidationCheckpoint(cfg.lettersCSV, state); err != nil {
			return err
		}
		printValidationResult(out, i+1, len(selected), "letters", result)
	}
	return finishValidation(cfg.lettersCSV, state, out)
}

func executeAIValidation(cfg config, client llmExecutor, check string, image imageInput, transcription, structured, digest string) validationResult {
	prompt, attachment := transcriptionValidationPrompt(transcription), image.Path
	if check == validationParsing {
		prompt, attachment = parsingValidationPrompt(transcription, structured), ""
	}
	result := validationResult{ImagePath: cleanPath(image.Path), Volume: image.Volume, Provider: cfg.provider, Model: cfg.model, ValidatedAt: time.Now().UTC(), SourceDigest: digest}
	raw, err := client.Exec(cfg.provider, cfg.model, prompt, attachment)
	if err != nil {
		result.Failure = err.Error()
		return result
	}
	response, err := parseStrictJSON[validationResponse](raw)
	if err != nil {
		result.Failure = fmt.Sprintf("parse validation response: %v", err)
		result.RawResponse = raw
		return result
	}
	if response.Differences == nil {
		result.Failure = "validation response requires a differences array"
		result.RawResponse = raw
		return result
	}
	if !response.Accurate && len(response.Differences) == 0 {
		result.Failure = "inaccurate validation response requires at least one difference"
		result.RawResponse = raw
		return result
	}
	result.Accurate, result.Summary, result.Differences = response.Accurate, strings.TrimSpace(response.Summary), response.Differences
	return result
}

func transcriptionValidationPrompt(transcription string) string {
	return `Compare the supplied page image against the transcription below. Decide whether the transcription is accurate and complete, including spelling, punctuation, line order, numbers, and Markdown bold markers. Ignore harmless Markdown layout differences only when they do not alter content or bold meaning.
Return JSON only in this exact shape:
{"accurate":true,"summary":"short verdict","differences":[{"location":"where on page","expected":"text visible in image","actual":"text in transcription","reason":"short explanation"}]}
Use an empty differences array only when accurate is true. When inaccurate, list every concrete difference you can identify. Do not rewrite the whole page.

<transcription>
` + transcription + "\n</transcription>"
}

func parsingValidationPrompt(transcription, structured string) string {
	return `Act as a meticulous machine reviewer. Compare the transcription (the source of truth) with the parsed JSON. Decide whether every source entry is represented exactly once and every field is faithful. Check omissions, duplicates, grouping, cross-references, page numbers, names, and bold flags. Do not use outside knowledge.
Return JSON only in this exact shape:
{"accurate":true,"summary":"short verdict","differences":[{"location":"source entry or JSON item","expected":"correct JSON value or entry derived from transcription","actual":"current JSON value, entry, or <missing>","reason":"short explanation"}]}
Use an empty differences array only when accurate is true. When inaccurate, provide complete, actionable differences that another agent can apply without seeing an image. Do not rewrite unchanged data.

<transcription>
` + transcription + "\n</transcription>\n\n<parsed_json>\n" + structured + "\n</parsed_json>"
}

func loadValidationManifest(output, kind, check string, resume bool) (validationManifest, error) {
	state := validationManifest{Version: validationManifestVersion, Kind: kind, Check: check, Results: []validationResult{}}
	if !resume {
		return state, saveValidationCheckpoint(output, state)
	}
	err := loadJSON(validationStatePath(output, check), &state)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if state.Version != validationManifestVersion || state.Kind != kind || state.Check != check {
		return state, fmt.Errorf("invalid validation metadata in %s", validationStatePath(output, check))
	}
	return state, nil
}

func shouldSkipValidation(state validationManifest, path, digest string, cfg config, targeted bool) bool {
	if !cfg.resume || targeted {
		return false
	}
	for _, result := range state.Results {
		if cleanPath(result.ImagePath) == cleanPath(path) && result.SourceDigest == digest && result.Provider == cfg.provider && result.Model == cfg.model {
			return result.Failure == "" || cfg.skipFailures
		}
	}
	return false
}

func validationSourceDigest(check, transcription, structured string) string {
	if check == validationTranscription {
		structured = ""
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(check+"\x00"+transcription+"\x00"+structured)))
}

func upsertValidationResult(results []validationResult, result validationResult) []validationResult {
	for i := range results {
		if cleanPath(results[i].ImagePath) == result.ImagePath {
			results[i] = result
			return sortValidationResults(results)
		}
	}
	return sortValidationResults(append(results, result))
}

func sortValidationResults(results []validationResult) []validationResult {
	sort.Slice(results, func(i, j int) bool { return results[i].ImagePath < results[j].ImagePath })
	return results
}

func saveValidationCheckpoint(output string, state validationManifest) error {
	if err := saveJSONAtomically(validationStatePath(output, state.Check), state); err != nil {
		return err
	}
	return renderValidationReview(validationReviewPath(output, state.Check), state)
}

func renderValidationReview(path string, state validationManifest) error {
	return writeFileAtomically(path, func(file *os.File) error {
		fmt.Fprintf(file, "# %s %s validation review\n\n", state.Kind, state.Check)
		for _, result := range sortValidationResults(state.Results) {
			status := "ACCURATE"
			if result.Failure != "" {
				status = "FAILED"
			} else if !result.Accurate {
				status = "REVIEW REQUIRED"
			}
			fmt.Fprintf(file, "## %s — %s\n\n- Volume: `%s`\n- Image: `%s`\n- Validator: `%s` / `%s`\n- Validated: `%s`\n", status, result.ImagePath, result.Volume, result.ImagePath, result.Provider, result.Model, result.ValidatedAt.Format(time.RFC3339))
			if result.Failure != "" {
				fmt.Fprintf(file, "- Failure: %s\n\n", result.Failure)
				if result.RawResponse != "" {
					fmt.Fprintln(file, "### Raw response")
					for _, line := range strings.Split(result.RawResponse, "\n") {
						fmt.Fprintf(file, "    %s\n", line)
					}
					fmt.Fprintln(file)
				}
				continue
			}
			fmt.Fprintf(file, "- Summary: %s\n\n", result.Summary)
			for i, difference := range result.Differences {
				fmt.Fprintf(file, "### Difference %d\n\n- Location: %s\n- Expected: `%s`\n- Actual: `%s`\n- Reason: %s\n\n", i+1, difference.Location, markdownInline(difference.Expected), markdownInline(difference.Actual), difference.Reason)
			}
		}
		return nil
	})
}

func markdownInline(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "`", "'"), "\n", " ")
}

func printValidationResult(out io.Writer, current, total int, kind string, result validationResult) {
	if result.Failure != "" {
		fmt.Fprintf(out, "[%d/%d] %s validation failed for %s: %s\n", current, total, kind, result.ImagePath, result.Failure)
		return
	}
	fmt.Fprintf(out, "[%d/%d] %s %s: accurate=%t differences=%d\n", current, total, kind, result.ImagePath, result.Accurate, len(result.Differences))
}

func finishValidation(output string, state validationManifest, out io.Writer) error {
	if err := saveValidationCheckpoint(output, state); err != nil {
		return err
	}
	fmt.Fprintf(out, "Validation checkpoint: %s; human review: %s\n", validationStatePath(output, state.Check), validationReviewPath(output, state.Check))
	return nil
}
