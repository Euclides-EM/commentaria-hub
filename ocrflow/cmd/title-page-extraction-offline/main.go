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
	"unicode"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/textmatch"
	"github.com/joho/godotenv"
)

const (
	defaultDatasetID     = "tps"
	defaultAIProvider    = "ollama"
	defaultAIModel       = "gpt-oss:120b"
	defaultMetadataDir   = "store/items_metadata"
	defaultTranscription = "paratext_transcriptions.csv"
	defaultItemsPrint    = "items_print.csv"
)

type cliConfig struct {
	datasetID       string
	keys            string
	keysFile        string
	features        string
	aiProvider      string
	aiModel         string
	metadataDir     string
	transcriptsCSV  string
	itemsPrintCSV   string
	outputCSV       string
	checkpointFile  string
	resume          bool
	includeColophon bool
}

type offlineFeature struct {
	ID          string
	Name        string
	RevisionID  string
	Prompt      string
	IsList      bool
	IsDefault   bool
	ImprintOnly bool
}

type textRecord struct {
	Key      string
	Title    string
	Imprint  string
	Colophon string
	Language string
}

type resultRow struct {
	EditionID      string
	FeatureID      string
	FeatureName    string
	SourceID       string
	SourceRevision string
	SourceName     string
	Value          string
	Properties     string
}

type parseResult struct {
	rows                   []resultRow
	hallucinatedFeatureIDs []string
}

func main() {
	log.SetFlags(0)

	cfg := parseFlags()
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env_private")

	if cfg.outputCSV == "" {
		log.Fatal("-output-csv is required")
	}

	records, err := loadTextRecords(cfg)
	if err != nil {
		log.Fatal(err)
	}
	keys, err := loadKeys(cfg, records)
	if err != nil {
		log.Fatal(err)
	}
	features, err := selectFeatures(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if !cfg.resume {
		if err := resetCSV(cfg.outputCSV); err != nil {
			log.Fatalf("initialize output CSV: %v", err)
		}
	}
	completed, err := loadCompletedKeys(cfg)
	if err != nil {
		log.Fatal(err)
	}

	llmClient := llm.NewClient(os.Getenv("OPENAI_API_KEY"), os.Getenv("OLLAMA_BASE_URL"), os.Getenv("OLLAMA_AUTH_TOKEN"))
	execID := "offline_v8"
	totalRows := 0
	completedThisRun := 0
	fmt.Printf("Running offline title-page extraction on dataset %s for %d keys and %d features\n", cfg.datasetID, len(keys), len(features))

	for i, key := range keys {
		if _, ok := completed[key]; ok {
			fmt.Printf("[%d/%d] skipping completed key %s\n", i+1, len(keys), key)
			continue
		}
		record, ok := records[key]
		if !ok {
			log.Fatalf("key %s not found in %s", key, cfg.transcriptsCSV)
		}
		fmt.Printf("[%d/%d] running key %s\n", i+1, len(keys), key)
		rows, err := executeKey(cfg, llmClient, execID, record, features)
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
	fs.StringVar(&cfg.datasetID, "dataset", defaultDatasetID, "dataset id for output scope")
	fs.StringVar(&cfg.keys, "keys", "", "selected edition keys, separated by commas or whitespace; defaults to all transcription rows")
	fs.StringVar(&cfg.keysFile, "keys-file", "", "file containing selected edition keys, separated by commas or whitespace; CSV files may have an edition_id or key column")
	fs.StringVar(&cfg.features, "features", "", "feature ids to run, separated by commas or whitespace; defaults to current default TPS feature set")
	fs.StringVar(&cfg.aiProvider, "ai-provider", defaultAIProvider, "AI provider")
	fs.StringVar(&cfg.aiModel, "ai-model", defaultAIModel, "AI model")
	fs.StringVar(&cfg.metadataDir, "metadata-dir", defaultMetadataDir, "directory containing paratext_transcriptions.csv and items_print.csv")
	fs.StringVar(&cfg.transcriptsCSV, "transcripts-csv", "", "optional explicit paratext transcriptions CSV path")
	fs.StringVar(&cfg.itemsPrintCSV, "items-print-csv", "", "optional explicit items_print metadata CSV path")
	fs.StringVar(&cfg.outputCSV, "output-csv", "", "CSV preview path for produced results")
	fs.StringVar(&cfg.checkpointFile, "checkpoint-file", "", "path to resume checkpoint file; defaults to output CSV path plus .done")
	fs.BoolVar(&cfg.resume, "resume", true, "skip keys already recorded in the checkpoint file")
	fs.BoolVar(&cfg.includeColophon, "include-colophon", false, "append colophon text to the imprint section")
	fs.Parse(os.Args[1:])

	if cfg.transcriptsCSV == "" {
		cfg.transcriptsCSV = filepath.Join(cfg.metadataDir, defaultTranscription)
	}
	if cfg.itemsPrintCSV == "" {
		cfg.itemsPrintCSV = filepath.Join(cfg.metadataDir, defaultItemsPrint)
	}
	return cfg
}

func executeKey(cfg cliConfig, llmClient *llm.Client, execID string, record textRecord, features []offlineFeature) ([]resultRow, error) {
	imprintFeatures, titleFeatures := partitionFeatures(features, func(f offlineFeature) bool {
		return f.ImprintOnly
	})

	var rows []resultRow
	if len(titleFeatures) > 0 && strings.TrimSpace(record.Title) != "" {
		r, err := executeFeatureGroup(cfg, llmClient, execID, record, titleFeatures, "a title page excluding the imprint section", record.Title)
		if err != nil {
			return nil, err
		}
		rows = append(rows, r...)
	}
	if len(imprintFeatures) > 0 {
		text := record.Imprint
		if cfg.includeColophon && strings.TrimSpace(record.Colophon) != "" {
			text = strings.TrimSpace(text + "\n" + record.Colophon)
		}
		if strings.TrimSpace(text) != "" {
			r, err := executeFeatureGroup(cfg, llmClient, execID, record, imprintFeatures, "the imprint section of a title page", text)
			if err != nil {
				return nil, err
			}
			rows = append(rows, r...)
		}
	}
	return rows, nil
}

func executeFeatureGroup(cfg cliConfig, llmClient *llm.Client, execID string, record textRecord, features []offlineFeature, textDescription string, sourceText string) ([]resultRow, error) {
	active := features
	rowsByFeatureID := make(map[string][]resultRow, len(features))
	for attempt := 0; attempt < 3 && len(active) > 0; attempt++ {
		fieldToIndex, definitions, outputFormat := buildPromptComponents(active)
		prompt := buildPrompt(textDescription, record.Language, outputFormat, strings.Join(definitions, "\n"), sourceText)
		contextDesc := fmt.Sprintf("dataset %s and key %s", cfg.datasetID, record.Key)
		rawResponse, err := llmClient.Exec(cfg.aiProvider, cfg.aiModel, prompt, "")
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for %s using %s/%s: %w", contextDesc, cfg.aiProvider, cfg.aiModel, err)
		}
		rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM response for %s: %w", contextDesc, err)
		}
		parsed, err := parseLLMResults(rawFields, active, fieldToIndex, execID, record.Key, contextDesc, sourceText)
		if err != nil {
			return nil, err
		}
		for _, row := range parsed.rows {
			rowsByFeatureID[row.FeatureID] = append(rowsByFeatureID[row.FeatureID], row)
		}
		if len(parsed.hallucinatedFeatureIDs) == 0 {
			break
		}
		active = filterFeatures(active, parsed.hallucinatedFeatureIDs)
	}

	var rows []resultRow
	for _, f := range features {
		rows = append(rows, rowsByFeatureID[f.ID]...)
	}
	return rows, nil
}

func buildPrompt(textDescription, textLanguage, outputFormat, definitions, fullText string) string {
	if strings.TrimSpace(textLanguage) == "" {
		textLanguage = "an unknown language"
	}
	return fmt.Sprintf(`You are an AI agent designed to extract structured metadata from title pages of early modern European textbooks.

You will be given:
- The transcribed text of %s in %s.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.

Extraction rules:
- Extract the exact source span that best satisfies each feature definition. Some features require a minimal unit; others require the fuller title, name, or descriptive phrase. Follow the span guidance in each feature definition.
- Do not over-trim titles, names, book counts, language references, or descriptors that are part of the requested feature.
- Omit only surrounding text, outer punctuation, or layout noise that is not part of the selected span.
- Preserve the original spelling, capitalization, whitespace, line breaks, and punctuation within the extracted span exactly as they appear in the transcription.
- Early modern orthography may differ from modern spelling and letter usage. For example, "v" and "u" or "i" and "j" may be interchangeable ("vpon," "Iesus"), and other historical spellings may vary. Treat these as normal forms and reproduce the text exactly as written, without modernization or normalization.
- Words or phrases may be split across lines or interrupted by characters such as "-", "=" or similar separators. Interpret these as part of the transcription layout and extract the relevant text accurately.
- Some text may apply to more than one field, so the same text may appear in multiple fields if applicable.
- If a list-valued feature contains multiple distinct values in a coordinated phrase, return each distinct value separately unless the feature definition asks for a combined phrase.
- Do not normalize, modernize, interpret, or correct the text.

Return only a valid JSON object. Do not include explanations or any other output.

Output format:
{
  %s
}

Definitions:
%s

Transcribed text:
%s
`, textDescription, textLanguage, outputFormat, definitions, fullText)
}

func buildPromptComponents(features []offlineFeature) (map[string]int, []string, string) {
	fieldToIndex := make(map[string]int, len(features))
	var definitions []string
	var outputFormat strings.Builder
	for i, f := range features {
		fieldName := promptFieldName(f)
		fieldToIndex[fieldName] = i
		definitions = append(definitions, fmt.Sprintf("- %s: %s", fieldName, f.Prompt))
		if f.IsList {
			fmt.Fprintf(&outputFormat, `  "%s": [...], // zero or more values`+"\n", fieldName)
		} else {
			fmt.Fprintf(&outputFormat, `  "%s": "...", // a single value or empty if not applicable`+"\n", fieldName)
		}
	}
	return fieldToIndex, definitions, strings.TrimSpace(outputFormat.String())
}

func parseLLMResults(rawFields map[string]json.RawMessage, features []offlineFeature, fieldToIndex map[string]int, execID string, key string, contextDesc string, sourceText string) (parseResult, error) {
	for fn, rawValue := range rawFields {
		if _, ok := fieldToIndex[fn]; !ok {
			return parseResult{}, fmt.Errorf("llm response contained unknown field %q for %s\n%s", fn, contextDesc, rawValue)
		}
	}

	var rows []resultRow
	var hallucinatedFeatureIDs []string
	for _, f := range features {
		fn := promptFieldName(f)
		rawValue, ok := rawFields[fn]
		var values []string
		if ok {
			if f.IsList {
				if err := json.Unmarshal(rawValue, &values); err != nil {
					var val string
					if retryErr := json.Unmarshal(rawValue, &val); retryErr != nil {
						return parseResult{}, fmt.Errorf("failed to parse list response for field %q in %s: %w:\n%s", fn, contextDesc, err, rawValue)
					}
					if val != "" {
						values = []string{val}
					}
				}
			} else {
				var val string
				if err := json.Unmarshal(rawValue, &val); err != nil {
					return parseResult{}, fmt.Errorf("failed to parse scalar response for field %q in %s: %w\n%s", fn, contextDesc, err, rawValue)
				}
				if val != "" {
					values = []string{val}
				}
			}
		}

		for _, v := range values {
			v = trimFeatureValue(v)
			if v == "" {
				continue
			}
			if len(textmatch.FindLoosePhraseMatches(sourceText, v)) == 0 {
				if span, ok := textmatch.FindFuzzyPhraseMatch(sourceText, v, 2); ok {
					sourceValue := strings.TrimSpace(sourceText[span[0]:span[1]])
					log.Printf("warning: llm fuzzy-grounded near hallucination: feature=%s revision=%s key=%s context=%s value=%q source_value=%q", f.ID, f.RevisionID, key, contextDesc, v, sourceValue)
					v = sourceValue
				} else {
					log.Printf("!!! llm hallucination omitted: feature=%s revision=%s key=%s context=%s value=%q", f.ID, f.RevisionID, key, contextDesc, v)
					if !slices.Contains(hallucinatedFeatureIDs, f.ID) {
						hallucinatedFeatureIDs = append(hallucinatedFeatureIDs, f.ID)
					}
					continue
				}
			}
			rows = append(rows, resultRow{
				EditionID:      key,
				FeatureID:      f.ID,
				FeatureName:    f.Name,
				SourceID:       execID,
				SourceRevision: f.RevisionID,
				SourceName:     "llm",
				Value:          v,
			})
		}
	}
	return parseResult{rows: rows, hallucinatedFeatureIDs: hallucinatedFeatureIDs}, nil
}

func promptFieldName(f offlineFeature) string {
	prefix := f.RevisionID
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	return fmt.Sprintf("%s-rev-%s", f.ID, prefix)
}

func loadTextRecords(cfg cliConfig) (map[string]textRecord, error) {
	langs, err := loadLanguages(cfg.itemsPrintCSV)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(cfg.transcriptsCSV)
	if err != nil {
		return nil, fmt.Errorf("open transcriptions CSV: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read transcriptions CSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, errors.New("transcriptions CSV is empty")
	}
	header := indexHeader(rows[0])
	required := []string{"key", "title", "imprint"}
	for _, name := range required {
		if _, ok := header[name]; !ok {
			return nil, fmt.Errorf("transcriptions CSV missing %q column", name)
		}
	}

	records := make(map[string]textRecord, len(rows)-1)
	for _, row := range rows[1:] {
		key := cell(row, header, "key")
		if strings.TrimSpace(key) == "" {
			continue
		}
		records[key] = textRecord{
			Key:      key,
			Title:    cell(row, header, "title"),
			Imprint:  cell(row, header, "imprint"),
			Colophon: cell(row, header, "colophon"),
			Language: langs[key],
		}
	}
	return records, nil
}

func loadLanguages(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open items_print CSV: %w", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read items_print CSV: %w", err)
	}
	if len(rows) == 0 {
		return nil, errors.New("items_print CSV is empty")
	}
	header := indexHeader(rows[0])
	out := make(map[string]string, len(rows)-1)
	for _, row := range rows[1:] {
		key := cell(row, header, "key")
		if key != "" {
			out[key] = cell(row, header, "language")
		}
	}
	return out, nil
}

func loadKeys(cfg cliConfig, records map[string]textRecord) ([]string, error) {
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
		for key := range records {
			keys = append(keys, key)
		}
		slices.Sort(keys)
	}
	if len(keys) == 0 {
		return nil, errors.New("at least one key is required")
	}
	for _, key := range keys {
		if _, ok := records[key]; !ok {
			return nil, fmt.Errorf("key %s not found in transcriptions CSV", key)
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
			for _, name := range []string{"edition_id", "key", "page_key"} {
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

func filterKeyList(keys []string) []string {
	var out []string
	for _, key := range keys {
		key = strings.TrimSpace(key)
		switch strings.ToLower(key) {
		case "", "edition_id", "key", "page_key":
			continue
		default:
			out = append(out, key)
		}
	}
	return out
}

func selectFeatures(cfg cliConfig) ([]offlineFeature, error) {
	byID := make(map[string]offlineFeature, len(currentFeatures))
	for _, f := range currentFeatures {
		byID[f.ID] = f
	}
	ids := splitList(cfg.features)
	var selected []offlineFeature
	if len(ids) == 0 {
		for _, f := range currentFeatures {
			if f.IsDefault {
				selected = append(selected, f)
			}
		}
		return selected, nil
	}
	for _, id := range ids {
		f, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("unknown feature %s", id)
		}
		selected = append(selected, f)
	}
	return selected, nil
}

func partitionFeatures(features []offlineFeature, pred func(offlineFeature) bool) (match, rest []offlineFeature) {
	for _, f := range features {
		if pred(f) {
			match = append(match, f)
		} else {
			rest = append(rest, f)
		}
	}
	return match, rest
}

func filterFeatures(features []offlineFeature, ids []string) []offlineFeature {
	var out []offlineFeature
	for _, f := range features {
		if slices.Contains(ids, f.ID) {
			out = append(out, f)
		}
	}
	return out
}

func resetCSV(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	if err := writeHeader(w); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func appendRowsCSV(path string, rows []resultRow) error {
	if err := ensureCSV(path); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	for _, row := range rows {
		if err := w.Write([]string{row.EditionID, row.FeatureID, row.FeatureName, row.SourceID, row.SourceRevision, row.SourceName, row.Value, row.Properties}); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func ensureCSV(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	return resetCSV(path)
}

func writeHeader(w *csv.Writer) error {
	return w.Write([]string{"edition_id", "feature_id", "feature_name", "source_id", "source_revision", "source_name", "value", "properties"})
}

func loadCompletedKeys(cfg cliConfig) (map[string]struct{}, error) {
	completed := make(map[string]struct{})
	if !cfg.resume {
		return completed, nil
	}
	path := checkpointPath(cfg)
	if path == "" {
		return completed, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return completed, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read checkpoint file: %w", err)
	}
	for _, key := range splitList(string(data)) {
		if key != "edition_id" && key != "key" {
			completed[key] = struct{}{}
		}
	}
	return completed, nil
}

func markCompletedKey(cfg cliConfig, key string) error {
	path := checkpointPath(cfg)
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	needsHeader := false
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		needsHeader = true
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if needsHeader {
		if _, err := f.WriteString("edition_id\n"); err != nil {
			return err
		}
	}
	_, err = f.WriteString(key + "\n")
	return err
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

func indexHeader(header []string) map[string]int {
	index := make(map[string]int, len(header))
	for i, name := range header {
		index[strings.TrimSpace(name)] = i
	}
	return index
}

func cell(row []string, header map[string]int, name string) string {
	i, ok := header[name]
	if !ok || i >= len(row) {
		return ""
	}
	return row[i]
}

func splitList(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	})
}

func uniqNonEmpty(values []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func trimFeatureValue(s string) string {
	return strings.TrimFunc(strings.TrimSpace(s), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}
