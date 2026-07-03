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
	rawIndexDir        = "cmd/indexextractor/data/raw_test/index"
	rawLettersTableDir = "cmd/indexextractor/data/raw_test/letters_table"
	defaultIndexCSV    = "cmd/indexextractor/data/index.csv"
	defaultLettersCSV  = "cmd/indexextractor/data/letters_table.csv"
	defaultAIProvider  = llm.ProviderOllama
	defaultAIModel     = "qwen3.6:35b"
)

const indexPrompt = `Extract every entry from this index page.
Return JSON only, in this exact shape:
{"entries":[{"name":"entry exactly as printed","page_number":"one page reference exactly as printed","is_bold":false}]}

Rules:
- Preserve the spelling, punctuation, capitalization, and page reference from the image.
- Produce one object per individual page reference. Repeat the same entry name when it has multiple page references.
- page_number must contain exactly one page reference, never a comma-separated list of references.
- Set is_bold to true only when that page reference is printed in bold type. Ignore whether the entry name is bold.
- A cross-reference such as "v. Debeaune" is one page reference; do not split it into words.
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

type config struct {
	command     string
	kind        string
	indexDir    string
	lettersDir  string
	indexCSV    string
	lettersCSV  string
	provider    string
	model       string
	resume      bool
	rerun       bool
	rerunImages string
}

const (
	commandExtract  = "extract"
	commandValidate = "validate"
	commandStatus   = "status"
	kindAll         = "all"
	kindIndex       = "index"
	kindLetters     = "letters"
)

type imageInput struct {
	Path   string
	Volume string
}

type indexEntry struct {
	Name       string `json:"name"`
	PageNumber string `json:"page_number"`
	IsBold     bool   `json:"is_bold"`
	Volume     string `json:"-"`
}

type letterEntry struct {
	LetterNumber string `json:"letter_number"`
	LetterName   string `json:"letter_name"`
	PageNumber   string `json:"page_number"`
	Volume       string `json:"-"`
}

type indexResponse struct {
	Entries []indexEntry `json:"entries"`
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
	cfg := config{command: commandExtract, kind: kindAll, resume: true}
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
	fs.BoolVar(&cfg.resume, "resume", true, "skip images already recorded in the output manifest")
	fs.BoolVar(&cfg.rerun, "rerun", false, "discard the selected manifest and output, then extract every image")
	fs.StringVar(&cfg.rerunImages, "rerun-images", "", "comma-separated image paths to extract again and replace in the manifest")
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
		required["AI provider"] = cfg.provider
		required["AI model"] = cfg.model
	}
	for name, value := range required {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not be empty", name)
		}
	}
	if cfg.command == commandExtract {
		provider := strings.ToLower(strings.TrimSpace(cfg.provider))
		if provider != llm.ProviderOpenAI && provider != llm.ProviderOllama {
			return fmt.Errorf("unsupported AI provider %q", cfg.provider)
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
	switch strings.ToLower(strings.TrimSpace(cfg.provider)) {
	case llm.ProviderOpenAI:
		if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) == "" {
			return errors.New("OPENAI_API_KEY is required for the openai provider")
		}
	case llm.ProviderOllama:
		if strings.TrimSpace(os.Getenv("OLLAMA_BASE_URL")) == "" {
			return errors.New("OLLAMA_BASE_URL is required for the ollama provider")
		}
	}
	return nil
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
	response, err := parseStrictJSON[indexResponse](raw)
	if err != nil {
		return nil, err
	}
	if response.Entries == nil {
		return nil, errors.New("response requires an entries array")
	}
	entries := make([]indexEntry, 0, len(response.Entries))
	for i := range response.Entries {
		entry := response.Entries[i]
		entry.Name = strings.TrimSpace(entry.Name)
		entry.PageNumber = strings.TrimSpace(entry.PageNumber)
		if isIndexSectionHeading(entry) {
			continue
		}
		entry.Volume = volume
		if entry.Name == "" || entry.PageNumber == "" {
			return nil, fmt.Errorf("entry %d requires name and page_number", i+1)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func isIndexSectionHeading(entry indexEntry) bool {
	if entry.PageNumber != "" || utf8.RuneCountInString(entry.Name) != 1 {
		return false
	}
	r, _ := utf8.DecodeRuneInString(entry.Name)
	return unicode.IsLetter(r)
}

func parseLettersResponse(raw, volume string) ([]letterEntry, error) {
	response, err := parseStrictJSON[lettersResponse](raw)
	if err != nil {
		return nil, err
	}
	if response.Entries == nil {
		return nil, errors.New("response requires an entries array")
	}
	for i := range response.Entries {
		entry := &response.Entries[i]
		entry.LetterNumber = strings.TrimSpace(entry.LetterNumber)
		entry.LetterName = strings.TrimSpace(entry.LetterName)
		entry.PageNumber = strings.TrimSpace(entry.PageNumber)
		entry.Volume = volume
		if entry.LetterNumber == "" || entry.LetterName == "" || entry.PageNumber == "" {
			return nil, fmt.Errorf("entry %d requires letter_number, letter_name, and page_number", i+1)
		}
	}
	return response.Entries, nil
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
