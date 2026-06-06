package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/joho/godotenv"
)

const defaultMetadataDir = "store/items_metadata"

type cliConfig struct {
	keys           string
	keysFile       string
	metadataDir    string
	aiProvider     string
	aiModel        string
	outputCSV      string
	checkpointFile string
	resume         bool
}

type resultRow struct {
	EditionID      string
	FeatureID      string
	SourceID       string
	SourceRevision string
	SourceName     string
	Value          string
}

func main() {
	log.SetFlags(0)

	cfg := parseFlags()
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env_private")

	if cfg.outputCSV == "" {
		log.Fatal("-output-csv is required")
	}

	editionStore := store.NewEditionCSV(cfg.metadataDir, nil)
	editions, err := editionStore.LoadAllEditions()
	if err != nil {
		log.Fatalf("load editions from metadata CSVs: %v", err)
	}
	keys, err := loadKeys(cfg, editions)
	if err != nil {
		log.Fatal(err)
	}
	completed, err := loadCompletedKeys(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if !cfg.resume {
		if err := resetCSV(cfg.outputCSV); err != nil {
			log.Fatalf("initialize output CSV: %v", err)
		}
	}

	llmClient := llm.NewClient(os.Getenv("OPENAI_API_KEY"), os.Getenv("OLLAMA_BASE_URL"), os.Getenv("OLLAMA_AUTH_TOKEN"))
	execID := "edition_offline_v9"
	totalRows := 0
	completedThisRun := 0
	fmt.Printf("Running offline edition classification revision %s for %d selected keys\n", defaultRevisionID, len(keys))
	for i, key := range keys {
		if _, ok := completed[key]; ok {
			fmt.Printf("[%d/%d] skipping completed key %s\n", i+1, len(keys), key)
			continue
		}
		edition := editions[key]
		if edition == nil {
			log.Fatalf("edition %s not found in %s", key, cfg.metadataDir)
		}

		fmt.Printf("[%d/%d] running key %s\n", i+1, len(keys), key)
		rows, err := executeKey(cfg, llmClient, execID, edition)
		if err != nil {
			log.Fatalf("execution failed for key %s: %v\nResume with the same command; completed keys are recorded in %s", key, err, checkpointPath(cfg))
		}
		if err := appendRowsCSV(cfg.outputCSV, rows); err != nil {
			log.Fatalf("append output CSV for key %s: %v", key, err)
		}
		if err := markCompletedKey(cfg, key); err != nil {
			log.Fatalf("write checkpoint for key %s: %v", key, err)
		}
		totalRows += len(rows)
		completedThisRun++
		fmt.Printf("[%d/%d] completed key %s with %d result rows\n", i+1, len(keys), key, len(rows))
	}

	fmt.Printf("Wrote result preview CSV: %s\n", cfg.outputCSV)
	fmt.Printf("Offline dry run complete: %d keys completed in this run, %d result rows produced, none written to DB\n", completedThisRun, totalRows)
}

func parseFlags() cliConfig {
	var cfg cliConfig
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfg.keys, "keys", "", "selected edition keys, separated by commas or whitespace; defaults to all metadata rows")
	fs.StringVar(&cfg.keysFile, "keys-file", "", "file containing selected edition keys, separated by commas or whitespace; CSV files may have an edition_id or key column")
	fs.StringVar(&cfg.metadataDir, "metadata-dir", defaultMetadataDir, "directory containing edition metadata CSV files")
	fs.StringVar(&cfg.aiProvider, "ai-provider", defaultAIProvider, "AI provider")
	fs.StringVar(&cfg.aiModel, "ai-model", defaultAIModel, "AI model")
	fs.StringVar(&cfg.outputCSV, "output-csv", "", "CSV preview path for produced results")
	fs.StringVar(&cfg.checkpointFile, "checkpoint-file", "", "path to resume checkpoint file; defaults to output CSV path plus .done")
	fs.BoolVar(&cfg.resume, "resume", true, "skip keys already recorded in the checkpoint file")
	fs.Parse(os.Args[1:])
	return cfg
}

func executeKey(cfg cliConfig, llmClient *llm.Client, execID string, ed *model.Edition) ([]resultRow, error) {
	prompt := buildPrompt(formatEditionInfo(ed))
	rawResponse, err := llmClient.Exec(cfg.aiProvider, cfg.aiModel, prompt, "")
	if err != nil {
		return nil, fmt.Errorf("failed to execute LLM prompt for edition %s using %s/%s: %w", ed.Key, cfg.aiProvider, cfg.aiModel, err)
	}
	rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response for edition %s: %w", ed.Key, err)
	}
	values, err := parseClassifierResponse(rawFields, ed.Key)
	if err != nil {
		return nil, err
	}

	rows := make([]resultRow, 0, len(values))
	for _, value := range values {
		rows = append(rows, resultRow{
			EditionID:      ed.Key,
			FeatureID:      defaultFeatureID,
			SourceID:       execID,
			SourceRevision: defaultRevisionID,
			SourceName:     "llm",
			Value:          value,
		})
	}
	if len(rows) == 0 {
		rows = append(rows, resultRow{
			EditionID:      ed.Key,
			FeatureID:      defaultFeatureID,
			SourceID:       execID,
			SourceRevision: defaultRevisionID,
			SourceName:     "llm",
		})
	}
	return rows, nil
}

func buildPrompt(editionInfo string) string {
	return fmt.Sprintf(`You are an AI agent designed to classify historical textbook editions into subject categories.

You will be given structured metadata about a specific edition.

Your task is to answer the classification question based only on the provided metadata and return it as a JSON object.
Each output value must use only the exact category/classification strings requested in the definition.
Do not quote, translate, paraphrase, or add metadata text unless the definition explicitly asks for that format.
Do not infer subject relevance from editor, publisher, city, language, date, or general reputation alone.
Use "unknown" when the metadata is insufficient or ambiguous; use "unrelated" when the metadata provides no meaningful evidence for a category.

Return only a valid JSON. Do not include any other output.

Output format:
{
  "m_classifier-rev-6a0": [...] // zero or more values
}

Definitions:
- m_classifier-rev-6a0: %s

Edition metadata:
%s
`, subjectClassifierPrompt, editionInfo)
}

func parseClassifierResponse(rawFields map[string]json.RawMessage, key string) ([]string, error) {
	const fieldName = "m_classifier-rev-6a0"
	for fn := range rawFields {
		if fn != fieldName {
			return nil, fmt.Errorf("llm response contained unknown field %q for edition %s", fn, key)
		}
	}
	rawValue, ok := rawFields[fieldName]
	if !ok {
		return nil, nil
	}
	var values []string
	if err := json.Unmarshal(rawValue, &values); err != nil {
		var val string
		if retryErr := json.Unmarshal(rawValue, &val); retryErr != nil {
			return nil, fmt.Errorf("failed to parse classifier response for edition %s: %w:\n%s", key, err, rawValue)
		}
		if val != "" {
			values = []string{val}
		}
	}
	return trimValues(values), nil
}

func formatEditionInfo(ed *model.Edition) string {
	var b strings.Builder

	if ed.IsManuscript {
		intro := "This is a manuscript"
		if ed.ManuscriptClass != "" {
			intro += " of the " + ed.ManuscriptClass + " class"
			if ed.ManuscriptSubclass != nil {
				intro += " (" + *ed.ManuscriptSubclass + ")"
			}
		}
		switch {
		case ed.ManuscriptYearFrom != nil && ed.ManuscriptYearTo != nil:
			intro += fmt.Sprintf(", dated approximately %d-%d", *ed.ManuscriptYearFrom, *ed.ManuscriptYearTo)
		case ed.ManuscriptYearFrom != nil:
			intro += fmt.Sprintf(", dated from approximately %d", *ed.ManuscriptYearFrom)
		case ed.ManuscriptYearTo != nil:
			intro += fmt.Sprintf(", dated up to approximately %d", *ed.ManuscriptYearTo)
		}
		if ed.ShortTitle != "" {
			intro += ", known as \"" + ed.ShortTitle + "\""
		}
		b.WriteString(intro + ".\n")
	} else {
		intro := "This is a printed edition"
		if ed.ShortTitle != "" {
			intro += " known as \"" + ed.ShortTitle + "\""
		}
		var where []string
		if ed.Year != nil {
			where = append(where, "published in "+*ed.Year)
		}
		if len(ed.Cities) > 0 {
			where = append(where, "in "+strings.Join(ed.Cities, " and "))
		}
		if len(ed.Languages) > 0 {
			langs := make([]string, len(ed.Languages))
			for i, l := range ed.Languages {
				if len(l) > 0 {
					langs[i] = strings.ToUpper(l[:1]) + l[1:]
				}
			}
			where = append(where, "originally in "+strings.Join(langs, " and "))
		}
		if len(where) > 0 {
			intro += ", " + strings.Join(where, ", ")
		}
		b.WriteString(intro + ".\n")

		var people []string
		if len(ed.Editor) > 0 {
			people = append(people, strings.Join(ed.Editor, " and ")+" edited it")
		}
		if len(ed.Publisher) > 0 {
			people = append(people, "published by "+strings.Join(ed.Publisher, " and "))
		}
		if len(people) > 0 {
			b.WriteString(strings.Join(people, "; ") + ".\n")
		}

		if ed.Format != nil || ed.Volumes != nil {
			var phys []string
			if ed.Format != nil {
				phys = append(phys, fmt.Sprintf("%d deg", *ed.Format))
			}
			if ed.Volumes != nil {
				phys = append(phys, fmt.Sprintf("%d volume(s)", *ed.Volumes))
			}
			b.WriteString("Physical format: " + strings.Join(phys, ", ") + ".\n")
		}

		if ed.ReprintOf != nil {
			b.WriteString("This edition is a reprint.\n")
		}

		if ed.TitleEN != nil {
			b.WriteString("\nTitle page reads: \n" + *ed.TitleEN + "\n")
		} else if ed.Title != nil {
			b.WriteString("\nTitle page reads: \n" + *ed.Title + "\n")
		}
	}

	if ed.IsElements {
		content := "\nThe edition covers Euclid's Elements"
		if len(ed.Books) > 0 {
			content += ", specifically books " + formatBookRanges(ed.Books)
		}
		content += "."
		b.WriteString(content + "\n")
	}
	if len(ed.AdditionalContent) > 0 {
		b.WriteString("Additional content included: " + strings.Join(ed.AdditionalContent, ", ") + ".\n")
	}
	if ed.HasDiagrams != nil {
		if *ed.HasDiagrams {
			b.WriteString("The edition contains diagrams.\n")
		} else {
			b.WriteString("The edition does not contain diagrams.\n")
		}
	}
	if ed.Notes != "" {
		b.WriteString("\nNotes: " + ed.Notes + "\n")
	}

	return strings.TrimSpace(b.String())
}

func formatBookRanges(books []int) string {
	if len(books) == 0 {
		return ""
	}
	var parts []string
	start, end := books[0], books[0]
	for _, n := range books[1:] {
		if n == end+1 {
			end = n
			continue
		}
		parts = append(parts, formatBookRange(start, end))
		start, end = n, n
	}
	parts = append(parts, formatBookRange(start, end))
	return strings.Join(parts, ", ")
}

func formatBookRange(start, end int) string {
	if start == end {
		return fmt.Sprintf("%d", start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}

func loadKeys(cfg cliConfig, editions map[string]*model.Edition) ([]string, error) {
	var all []string
	all = append(all, splitList(cfg.keys)...)
	if strings.TrimSpace(cfg.keysFile) != "" {
		keys, err := readKeysFile(cfg.keysFile)
		if err != nil {
			return nil, err
		}
		all = append(all, keys...)
	}
	keys := uniqNonEmpty(all)
	if len(keys) == 0 {
		for key := range editions {
			keys = append(keys, key)
		}
		slices.Sort(keys)
	}
	if len(keys) == 0 {
		return nil, errors.New("at least one edition key is required")
	}
	for _, key := range keys {
		if editions[key] == nil {
			return nil, fmt.Errorf("edition %s not found in metadata CSVs", key)
		}
	}
	return keys, nil
}

func readKeysFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read keys file: %w", err)
	}
	text := string(data)
	if strings.Contains(text, "\n") && strings.Contains(strings.SplitN(text, "\n", 2)[0], ",") {
		r := csv.NewReader(strings.NewReader(text))
		r.FieldsPerRecord = -1
		rows, err := r.ReadAll()
		if err == nil && len(rows) > 0 {
			header := indexHeader(rows[0])
			for _, name := range []string{"edition_id", "key"} {
				if _, ok := header[name]; ok {
					var keys []string
					for _, row := range rows[1:] {
						keys = append(keys, cell(row, header, name))
					}
					return keys, nil
				}
			}
		}
	}
	return filterKeyList(splitList(text)), nil
}

func splitList(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
}

func filterKeyList(keys []string) []string {
	var out []string
	for _, key := range keys {
		key = strings.TrimSpace(key)
		switch strings.ToLower(key) {
		case "", "edition_id", "key":
			continue
		default:
			out = append(out, key)
		}
	}
	return out
}

func uniqNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	var out []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func trimValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func checkpointPath(cfg cliConfig) string {
	if strings.TrimSpace(cfg.checkpointFile) != "" {
		return cfg.checkpointFile
	}
	if strings.TrimSpace(cfg.outputCSV) != "" {
		return cfg.outputCSV + ".done"
	}
	return ""
}

func loadCompletedKeys(cfg cliConfig) (map[string]struct{}, error) {
	completed := make(map[string]struct{})
	path := checkpointPath(cfg)
	if !cfg.resume {
		if path != "" {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
		return completed, nil
	}
	if path == "" {
		return completed, nil
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return completed, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if len(row) == 0 || row[0] == "" || row[0] == "edition_id" {
			continue
		}
		completed[row[0]] = struct{}{}
	}
	if len(completed) > 0 {
		fmt.Printf("Resuming from checkpoint %s with %d completed keys\n", path, len(completed))
	}
	return completed, nil
}

func markCompletedKey(cfg cliConfig, key string) error {
	path := checkpointPath(cfg)
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	needsHeader, err := fileNeedsHeader(file)
	if err != nil {
		return err
	}

	w := csv.NewWriter(file)
	if needsHeader {
		if err := w.Write([]string{"edition_id"}); err != nil {
			return err
		}
	}
	if err := w.Write([]string{key}); err != nil {
		return err
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return file.Sync()
}

func resetCSV(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	if err := w.Write([]string{"edition_id", "feature_id", "source_id", "source_revision", "source_name", "value"}); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func appendRowsCSV(path string, rows []resultRow) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	needsHeader, err := fileNeedsHeader(file)
	if err != nil {
		return err
	}

	w := csv.NewWriter(file)
	if needsHeader {
		if err := w.Write([]string{"edition_id", "feature_id", "source_id", "source_revision", "source_name", "value"}); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := w.Write([]string{row.EditionID, row.FeatureID, row.SourceID, row.SourceRevision, row.SourceName, row.Value}); err != nil {
			return err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return file.Sync()
}

func fileNeedsHeader(file *os.File) (bool, error) {
	info, err := file.Stat()
	if err != nil {
		return false, err
	}
	return info.Size() == 0, nil
}

func indexHeader(row []string) map[string]int {
	header := make(map[string]int, len(row))
	for i, name := range row {
		header[strings.TrimSpace(name)] = i
	}
	return header
}

func cell(row []string, header map[string]int, name string) string {
	i, ok := header[name]
	if !ok || i < 0 || i >= len(row) {
		return ""
	}
	return row[i]
}
