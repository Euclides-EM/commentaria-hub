package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/joho/godotenv"
)

const (
	rawIndexDir        = "cmd/indexextractor/data/raw_test/index"
	rawLettersTableDir = "cmd/indexextractor/data/raw_test/letters_table"
	defaultIndexCSV    = "cmd/indexextractor/data/index.csv"
	defaultLettersCSV  = "cmd/indexextractor/data/letters_table.csv"
	defaultAIProvider  = llm.ProviderOpenAI
	defaultAIModel     = "gpt-5-mini"
)

const indexPrompt = `Extract every entry from this index page.
Return JSON only, in this exact shape:
{"entries":[{"name":"entry exactly as printed","page_number":"page reference exactly as printed","is_bold":false}]}

Rules:
- Preserve the spelling, punctuation, capitalization, and page reference from the image.
- Produce one object per entry. If an entry has multiple page references, keep them together exactly as printed.
- Set is_bold to true only when the entry name is printed in bold type.
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
	indexDir   string
	lettersDir string
	indexCSV   string
	lettersCSV string
	provider   string
	model      string
}

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
	cfg := parseFlags(os.Args[1:])
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env_private")
	if err := validateEnvironment(cfg); err != nil {
		log.Fatal(err)
	}

	if err := run(cfg, llm.NewClient(
		os.Getenv("OPENAI_API_KEY"),
		os.Getenv("OLLAMA_BASE_URL"),
		os.Getenv("OLLAMA_AUTH_TOKEN"),
	), os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(args []string) config {
	cfg := config{}
	fs := flag.NewFlagSet("indexextractor", flag.ExitOnError)
	fs.StringVar(&cfg.indexDir, "index-dir", rawIndexDir, "directory containing index images grouped by volume")
	fs.StringVar(&cfg.lettersDir, "letters-dir", rawLettersTableDir, "directory containing letters-table images grouped by volume")
	fs.StringVar(&cfg.indexCSV, "index-output", defaultIndexCSV, "output CSV for index entries")
	fs.StringVar(&cfg.lettersCSV, "letters-output", defaultLettersCSV, "output CSV for letters-table entries")
	fs.StringVar(&cfg.provider, "ai-provider", defaultAIProvider, "AI provider: openai or ollama")
	fs.StringVar(&cfg.model, "ai-model", defaultAIModel, "vision-capable AI model")
	_ = fs.Parse(args)
	return cfg
}

func run(cfg config, client llmExecutor, out io.Writer) error {
	if err := validateConfig(cfg); err != nil {
		return err
	}
	indexImages, err := discoverImages(cfg.indexDir)
	if err != nil {
		return fmt.Errorf("discover index images: %w", err)
	}
	letterImages, err := discoverImages(cfg.lettersDir)
	if err != nil {
		return fmt.Errorf("discover letters-table images: %w", err)
	}

	fmt.Fprintf(out, "Found %d index images and %d letters-table images\n", len(indexImages), len(letterImages))
	indexEntries, err := extractIndexEntries(cfg, client, indexImages, out)
	if err != nil {
		return err
	}
	letterEntries, err := extractLetterEntries(cfg, client, letterImages, out)
	if err != nil {
		return err
	}

	if err := writeIndexCSV(cfg.indexCSV, indexEntries); err != nil {
		return fmt.Errorf("write index CSV: %w", err)
	}
	if err := writeLettersCSV(cfg.lettersCSV, letterEntries); err != nil {
		return fmt.Errorf("write letters-table CSV: %w", err)
	}
	fmt.Fprintf(out, "Wrote %d index entries to %s\n", len(indexEntries), cfg.indexCSV)
	fmt.Fprintf(out, "Wrote %d letters-table entries to %s\n", len(letterEntries), cfg.lettersCSV)
	return nil
}

func validateConfig(cfg config) error {
	for name, value := range map[string]string{
		"index directory": cfg.indexDir, "letters-table directory": cfg.lettersDir,
		"index output": cfg.indexCSV, "letters-table output": cfg.lettersCSV,
		"AI provider": cfg.provider, "AI model": cfg.model,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s must not be empty", name)
		}
	}
	provider := strings.ToLower(strings.TrimSpace(cfg.provider))
	if provider != llm.ProviderOpenAI && provider != llm.ProviderOllama {
		return fmt.Errorf("unsupported AI provider %q", cfg.provider)
	}
	if filepath.Clean(cfg.indexCSV) == filepath.Clean(cfg.lettersCSV) {
		return errors.New("index output and letters-table output must be different files")
	}
	return nil
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

func extractIndexEntries(cfg config, client llmExecutor, images []imageInput, out io.Writer) ([]indexEntry, error) {
	var all []indexEntry
	for i, image := range images {
		fmt.Fprintf(out, "[%d/%d] Extracting index page %s\n", i+1, len(images), image.Path)
		raw, err := client.Exec(cfg.provider, cfg.model, indexPrompt, image.Path)
		if err != nil {
			return nil, fmt.Errorf("extract index image %s: %w", image.Path, err)
		}
		entries, err := parseIndexResponse(raw, image.Volume)
		if err != nil {
			return nil, fmt.Errorf("parse index response for %s: %w", image.Path, err)
		}
		all = append(all, entries...)
	}
	return all, nil
}

func extractLetterEntries(cfg config, client llmExecutor, images []imageInput, out io.Writer) ([]letterEntry, error) {
	var all []letterEntry
	for i, image := range images {
		fmt.Fprintf(out, "[%d/%d] Extracting letters-table page %s\n", i+1, len(images), image.Path)
		raw, err := client.Exec(cfg.provider, cfg.model, lettersTablePrompt, image.Path)
		if err != nil {
			return nil, fmt.Errorf("extract letters-table image %s: %w", image.Path, err)
		}
		entries, err := parseLettersResponse(raw, image.Volume)
		if err != nil {
			return nil, fmt.Errorf("parse letters-table response for %s: %w", image.Path, err)
		}
		all = append(all, entries...)
	}
	return all, nil
}

func parseIndexResponse(raw, volume string) ([]indexEntry, error) {
	response, err := parseStrictJSON[indexResponse](raw)
	if err != nil {
		return nil, err
	}
	if response.Entries == nil {
		return nil, errors.New("response requires an entries array")
	}
	for i := range response.Entries {
		entry := &response.Entries[i]
		entry.Name = strings.TrimSpace(entry.Name)
		entry.PageNumber = strings.TrimSpace(entry.PageNumber)
		entry.Volume = volume
		if entry.Name == "" || entry.PageNumber == "" {
			return nil, fmt.Errorf("entry %d requires name and page_number", i+1)
		}
	}
	return response.Entries, nil
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

func writeIndexCSV(path string, entries []indexEntry) error {
	return writeCSVAtomically(path, []string{"name", "page_number", "is_bold", "volume"}, func(writer *csv.Writer) error {
		for _, entry := range entries {
			if err := writer.Write([]string{entry.Name, entry.PageNumber, strconv.FormatBool(entry.IsBold), entry.Volume}); err != nil {
				return err
			}
		}
		return nil
	})
}

func writeLettersCSV(path string, entries []letterEntry) error {
	return writeCSVAtomically(path, []string{"letter_number", "letter_name", "page_number", "volume"}, func(writer *csv.Writer) error {
		for _, entry := range entries {
			if err := writer.Write([]string{entry.LetterNumber, entry.LetterName, entry.PageNumber, entry.Volume}); err != nil {
				return err
			}
		}
		return nil
	})
}

func writeCSVAtomically(path string, header []string, writeRows func(*csv.Writer) error) (returnErr error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(dir, ".indexextractor-*.csv")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() {
		if err := os.Remove(temporaryPath); err != nil && !errors.Is(err, os.ErrNotExist) && returnErr == nil {
			returnErr = err
		}
	}()

	writer := csv.NewWriter(temporary)
	if err := writer.Write(header); err == nil {
		err = writeRows(writer)
	}
	writer.Flush()
	if err == nil {
		err = writer.Error()
	}
	if closeErr := temporary.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}
