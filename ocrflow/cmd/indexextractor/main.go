package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/joho/godotenv"
)

const (
	rawIndexDir        = "cmd/indexextractor/data/raw/index"
	rawLettersTableDir = "cmd/indexextractor/data/raw/letters_table"
	defaultIndexCSV    = "cmd/indexextractor/data/index.csv"
	defaultLettersCSV  = "cmd/indexextractor/data/letters_table.csv"
	defaultAIProvider  = llm.ProviderOllama
	defaultAIModel     = "qwen3.6:35b"
)

const indexPrompt = `Extract every entry from this index page.
Return JSON only, in this exact shape:
{"entries":[{"name":"entry exactly as printed","page_references":[{"page_number":"one page reference exactly as printed","is_bold":false}],"reference":"cross-reference target without the v. marker, or empty for page-reference entries"}]}

Rules:
- Preserve the spelling, punctuation, capitalization, page references, and cross-reference targets from the image.
- Produce one entry object per printed index entry. Put all of its page references in page_references; do not repeat the entry name.
- Each page_references object must contain exactly one page reference, never a comma-separated list of references.
- For a cross-reference such as "La Boderie (Sr de), v. Lefèvre (Guy).", set name to "La Boderie (Sr de)", page_references to [], and reference to "Lefèvre (Guy)". Do not include "v." or the sentence-ending period in reference.
- Every entry must have either a non-empty page_references array or a populated reference, never both. Use an empty page_references array for a cross-reference.
- Set each page reference's is_bold to true only when that page reference is printed in bold type. Ignore whether the entry name is bold.
- Do not infer missing text and do not include headings, running headers, or page numbers of the index itself.
- Use an empty entries array when there are no index entries.`

const lettersTablePrompt = `Extract every row from this letters table page.
Return JSON only, in this exact shape:
{"entries":[{"letter_number":"number exactly as printed","letter_name":"name exactly as printed","page_number":"page reference exactly as printed"}]}

Rules:
- Preserve spelling, punctuation, capitalization, numbering, and page references from the image.
- Produce one object per table row.
- Do not infer missing text and do not include column headings, running headers, or the page number of the table itself.
- Use an empty entries array when there are no table rows.`

const transcriptionPrompt = `Transcribe all text in this page image exactly as printed.
Preserve spelling, punctuation, capitalization, line breaks, page references, and table or index layout as faithfully as possible.
Represent bold text with Markdown **bold** markers so that a later text-only extraction can identify bold page references.
Do not interpret, summarize, normalize, or omit text. Return only the transcription.`

type config struct {
	command        string
	kind           string
	indexDir       string
	lettersDir     string
	indexCSV       string
	lettersCSV     string
	provider       string
	model          string
	firstProvider  string
	firstModel     string
	secondProvider string
	secondModel    string
	resume         bool
	rerun          bool
	rerunImages    string
	extractionMode string
	skipFailures   bool
}

const (
	commandExtract  = "extract"
	commandValidate = "validate"
	commandStatus   = "status"
	kindAll         = "all"
	kindIndex       = "index"
	kindLetters     = "letters"
	modeOnePass     = "one-pass"
	modeTwoPass     = "two-pass"
)

type imageInput struct {
	Path   string
	Volume string
}

type indexEntry struct {
	Name       string `json:"name"`
	PageNumber string `json:"page_number"`
	Reference  string `json:"reference,omitempty"`
	IsBold     bool   `json:"is_bold"`
	Volume     string `json:"-"`
}

type letterEntry struct {
	LetterNumber string `json:"letter_number"`
	LetterName   string `json:"letter_name"`
	PageNumber   string `json:"page_number"`
	Volume       string `json:"-"`
}

type compactPageReference struct {
	PageNumber string `json:"page_number"`
	IsBold     bool   `json:"is_bold"`
}

type compactIndexEntry struct {
	Name           string                 `json:"name"`
	PageReferences []compactPageReference `json:"page_references"`
	Reference      string                 `json:"reference"`
}

type indexResponse struct {
	Entries []compactIndexEntry `json:"entries"`
}

type lettersResponse struct {
	Entries []letterEntry `json:"entries"`
}

type llmExecutor interface {
	Exec(provider, model, prompt, attachmentPath string) (string, error)
}

func main() {
	log.SetFlags(0)
	cfg, err := parseCLI(os.Args[1:], os.Stderr)
	if err != nil {
		log.Fatal(err)
	}
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env_private")
	if cfg.command == commandExtract {
		if err := validateEnvironment(cfg); err != nil {
			log.Fatal(err)
		}
	}

	var client llmExecutor
	if cfg.command == commandExtract {
		client = llm.NewClient(
			os.Getenv("OPENAI_API_KEY"),
			os.Getenv("OLLAMA_BASE_URL"),
			os.Getenv("OLLAMA_AUTH_TOKEN"),
		)
	}
	if err := executeCommand(cfg, client, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func parseCLI(args []string, errOut io.Writer) (config, error) {
	cfg := config{command: commandExtract, kind: kindAll, resume: true, extractionMode: modeOnePass}
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		cfg.command = args[0]
		args = args[1:]
	}
	if cfg.command != commandExtract && cfg.command != commandValidate && cfg.command != commandStatus {
		return cfg, fmt.Errorf("unknown command %q (want extract, validate, or status)", cfg.command)
	}
	fs := flag.NewFlagSet("indexextractor "+cfg.command, flag.ContinueOnError)
	fs.SetOutput(errOut)
	fs.StringVar(&cfg.kind, "kind", kindAll, "data kind: index, letters, or all")
	fs.StringVar(&cfg.indexDir, "index-dir", rawIndexDir, "directory containing index images grouped by volume")
	fs.StringVar(&cfg.lettersDir, "letters-dir", rawLettersTableDir, "directory containing letters-table images grouped by volume")
	fs.StringVar(&cfg.indexCSV, "index-output", defaultIndexCSV, "output CSV for index entries")
	fs.StringVar(&cfg.lettersCSV, "letters-output", defaultLettersCSV, "output CSV for letters-table entries")
	fs.StringVar(&cfg.provider, "ai-provider", defaultAIProvider, "AI provider: openai or ollama")
	fs.StringVar(&cfg.model, "ai-model", defaultAIModel, "vision-capable AI model")
	fs.StringVar(&cfg.firstProvider, "first-pass-ai-provider", "", "two-pass transcription provider (defaults to --ai-provider)")
	fs.StringVar(&cfg.firstModel, "first-pass-ai-model", "", "two-pass transcription model (defaults to --ai-model)")
	fs.StringVar(&cfg.secondProvider, "second-pass-ai-provider", "", "two-pass structured extraction provider (defaults to --ai-provider)")
	fs.StringVar(&cfg.secondModel, "second-pass-ai-model", "", "two-pass structured extraction model (defaults to --ai-model)")
	fs.BoolVar(&cfg.resume, "resume", true, "skip images already recorded in the output manifest")
	fs.BoolVar(&cfg.rerun, "rerun", false, "discard the selected manifest and output, then extract every image")
	fs.StringVar(&cfg.rerunImages, "rerun-images", "", "comma-separated image paths to extract again and replace in the manifest")
	fs.StringVar(&cfg.extractionMode, "extraction-mode", modeOnePass, "LLM workflow: one-pass or two-pass")
	fs.BoolVar(&cfg.skipFailures, "skip-failures", false, "do not retry pages whose last extraction attempt failed")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if fs.NArg() != 0 {
		return cfg, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}
	if cfg.rerun {
		cfg.resume = false
	}
	if cfg.rerun && strings.TrimSpace(cfg.rerunImages) != "" {
		return cfg, errors.New("--rerun and --rerun-images cannot be used together")
	}
	if cfg.kind == kindAll && strings.TrimSpace(cfg.rerunImages) != "" {
		return cfg, errors.New("--rerun-images requires --kind index or --kind letters")
	}
	return cfg, nil
}

func executeCommand(cfg config, client llmExecutor, out io.Writer) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}
	switch cfg.command {
	case commandExtract:
		return run(cfg, client, out)
	case commandValidate:
		return validateOutputs(cfg, out)
	case commandStatus:
		return reportStatus(cfg, out)
	default:
		return fmt.Errorf("unknown command %q", cfg.command)
	}
}

func run(cfg config, client llmExecutor, out io.Writer) error {
	if client == nil {
		return errors.New("extract requires an LLM client")
	}
	if includesKind(cfg.kind, kindIndex) {
		if err := runIndexExtraction(cfg, client, out); err != nil {
			return err
		}
	}
	if includesKind(cfg.kind, kindLetters) {
		if err := runLettersExtraction(cfg, client, out); err != nil {
			return err
		}
	}
	return nil
}

func validateConfig(cfg config) error {
	if cfg.kind != kindAll && cfg.kind != kindIndex && cfg.kind != kindLetters {
		return fmt.Errorf("unsupported kind %q (want index, letters, or all)", cfg.kind)
	}
	required := map[string]string{}
	if includesKind(cfg.kind, kindIndex) {
		required["index directory"] = cfg.indexDir
		required["index output"] = cfg.indexCSV
	}
	if includesKind(cfg.kind, kindLetters) {
		required["letters-table directory"] = cfg.lettersDir
		required["letters-table output"] = cfg.lettersCSV
	}
	if cfg.command == commandExtract {
		if effectiveExtractionMode(cfg) == modeTwoPass {
			required["first-pass AI provider"] = firstPassProvider(cfg)
			required["first-pass AI model"] = firstPassModel(cfg)
			required["second-pass AI provider"] = secondPassProvider(cfg)
			required["second-pass AI model"] = secondPassModel(cfg)
		} else {
			required["AI provider"] = cfg.provider
			required["AI model"] = cfg.model
		}
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not be empty", name)
		}
	}
	if cfg.command == commandExtract {
		providers := map[string]string{"AI": cfg.provider}
		if effectiveExtractionMode(cfg) == modeTwoPass {
			providers = map[string]string{"first-pass AI": firstPassProvider(cfg), "second-pass AI": secondPassProvider(cfg)}
		}
		for label, provider := range providers {
			provider = strings.ToLower(strings.TrimSpace(provider))
			if provider != llm.ProviderOpenAI && provider != llm.ProviderOllama {
				return fmt.Errorf("unsupported %s provider %q", label, provider)
			}
		}
		if effectiveExtractionMode(cfg) == modeTwoPass && (strings.TrimSpace(firstPassModel(cfg)) == "" || strings.TrimSpace(secondPassModel(cfg)) == "") {
			return errors.New("first-pass and second-pass AI models must not be empty")
		}
		if cfg.extractionMode != modeOnePass && cfg.extractionMode != modeTwoPass {
			return fmt.Errorf("unsupported extraction mode %q (want one-pass or two-pass)", cfg.extractionMode)
		}
	}
	if cfg.kind == kindAll && filepath.Clean(cfg.indexCSV) == filepath.Clean(cfg.lettersCSV) {
		return errors.New("index output and letters-table output must be different files")
	}
	return nil
}

func includesKind(selected, candidate string) bool {
	return selected == kindAll || selected == candidate
}

func validateEnvironment(cfg config) error {
	providers := []string{cfg.provider}
	if effectiveExtractionMode(cfg) == modeTwoPass {
		providers = []string{firstPassProvider(cfg), secondPassProvider(cfg)}
	}
	for _, provider := range providers {
		switch strings.ToLower(strings.TrimSpace(provider)) {
		case llm.ProviderOpenAI:
			if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
				return errors.New("OPENAI_API_KEY is required for the openai provider")
			}
		case llm.ProviderOllama:
			if strings.TrimSpace(os.Getenv("OLLAMA_BASE_URL")) == "" {
				return errors.New("OLLAMA_BASE_URL is required for the ollama provider")
			}
		}
	}
	return nil
}

func firstPassProvider(cfg config) string {
	if strings.TrimSpace(cfg.firstProvider) != "" {
		return cfg.firstProvider
	}
	return cfg.provider
}
func firstPassModel(cfg config) string {
	if strings.TrimSpace(cfg.firstModel) != "" {
		return cfg.firstModel
	}
	return cfg.model
}
func secondPassProvider(cfg config) string {
	if strings.TrimSpace(cfg.secondProvider) != "" {
		return cfg.secondProvider
	}
	return cfg.provider
}
func secondPassModel(cfg config) string {
	if strings.TrimSpace(cfg.secondModel) != "" {
		return cfg.secondModel
	}
	return cfg.model
}

func extractionProvider(cfg config) string {
	if effectiveExtractionMode(cfg) == modeTwoPass {
		return secondPassProvider(cfg)
	}
	return cfg.provider
}

func extractionModel(cfg config) string {
	if effectiveExtractionMode(cfg) == modeTwoPass {
		return secondPassModel(cfg)
	}
	return cfg.model
}

func discoverImages(root string) ([]imageInput, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	var images []imageInput
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !isImagePath(path) {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(relative), "/")
		if len(parts) < 2 || strings.TrimSpace(parts[0]) == "" {
			return fmt.Errorf("image %s must be inside a volume directory", path)
		}
		images = append(images, imageInput{Path: path, Volume: parts[0]})
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(images) == 0 {
		return nil, fmt.Errorf("no supported images found in %s", root)
	}
	sort.Slice(images, func(i, j int) bool { return images[i].Path < images[j].Path })
	return images, nil
}

func isImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}

func parseIndexResponse(raw, volume string) ([]indexEntry, error) {
	entries, _, err := parseIndexResponseWithIssues(raw, volume)
	return entries, err
}

func parseIndexResponseWithIssues(raw, volume string) ([]indexEntry, []string, error) {
	response, err := parseStrictJSON[indexResponse](raw)
	if err != nil {
		return nil, nil, err
	}
	if response.Entries == nil {
		return nil, nil, errors.New("response requires an entries array")
	}
	entries := make([]indexEntry, 0, len(response.Entries))
	issues := make([]string, 0)
	for i := range response.Entries {
		compact := response.Entries[i]
		compact.Name = strings.TrimSpace(compact.Name)
		compact.Reference = strings.TrimSpace(compact.Reference)
		if isIndexSectionHeading(indexEntry{Name: compact.Name}) && len(compact.PageReferences) == 0 && compact.Reference == "" {
			continue
		}
		if compact.Name == "" || (len(compact.PageReferences) == 0) == (compact.Reference == "") {
			issues = append(issues, fmt.Sprintf(
				"entry %d requires name and exactly one of page_references or reference (name=%q, page_references=%d, reference=%q)",
				i+1, compact.Name, len(compact.PageReferences), compact.Reference,
			))
			continue
		}
		if compact.Reference != "" {
			entries = append(entries, indexEntry{Name: compact.Name, Reference: compact.Reference, Volume: volume})
			continue
		}
		for j := range compact.PageReferences {
			pageRef := compact.PageReferences[j]
			pageRef.PageNumber = strings.TrimSpace(pageRef.PageNumber)
			if pageRef.PageNumber == "" {
				issues = append(issues, fmt.Sprintf("entry %d page reference %d requires page_number (name=%q)", i+1, j+1, compact.Name))
				continue
			}
			entries = append(entries, indexEntry{Name: compact.Name, PageNumber: pageRef.PageNumber, IsBold: pageRef.IsBold, Volume: volume})
		}
	}
	return entries, issues, nil
}

func isIndexSectionHeading(entry indexEntry) bool {
	if entry.PageNumber != "" || entry.Reference != "" || utf8.RuneCountInString(entry.Name) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(entry.Name)
	return unicode.IsLetter(r)
}

func parseLettersResponse(raw, volume string) ([]letterEntry, error) {
	entries, _, err := parseLettersResponseWithIssues(raw, volume)
	return entries, err
}

func parseLettersResponseWithIssues(raw, volume string) ([]letterEntry, []string, error) {
	response, err := parseStrictJSON[lettersResponse](raw)
	if err != nil {
		return nil, nil, err
	}
	if response.Entries == nil {
		return nil, nil, errors.New("response requires an entries array")
	}
	entries := make([]letterEntry, 0, len(response.Entries))
	issues := make([]string, 0)
	for i := range response.Entries {
		entry := response.Entries[i]
		entry.LetterNumber = strings.TrimSpace(entry.LetterNumber)
		entry.LetterName = strings.TrimSpace(entry.LetterName)
		entry.PageNumber = strings.TrimSpace(entry.PageNumber)
		entry.Volume = volume
		if entry.LetterNumber == "" || entry.LetterName == "" || entry.PageNumber == "" {
			issues = append(issues, fmt.Sprintf(
				"entry %d requires letter_number, letter_name, and page_number (letter_number=%q, letter_name=%q, page_number=%q)",
				i+1, entry.LetterNumber, entry.LetterName, entry.PageNumber,
			))
			continue
		}
		entries = append(entries, entry)
	}
	return entries, issues, nil
}

func parseStrictJSON[T any](raw string) (T, error) {
	var result T
	object, err := llm.ParseJSON[json.RawMessage](raw)
	if err != nil {
		return result, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(object)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}
