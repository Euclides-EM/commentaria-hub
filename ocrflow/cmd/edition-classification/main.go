package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/app"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/joho/godotenv"
)

const (
	defaultFeatureID  = "m_classifier"
	defaultRevisionID = "6a0d47e3-f472-4b63-a6f5-67c693a0adf9"
)

type cliConfig struct {
	keys           string
	keysFile       string
	featureID      string
	revisionID     string
	outputCSV      string
	checkpointFile string
	dryRun         bool
	resume         bool
	skipExisting   bool
}

func main() {
	log.SetFlags(0)

	cfg := parseFlags()
	keys, err := loadKeys(cfg.keys, cfg.keysFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	if err := godotenv.Load(".env_private"); err != nil {
		log.Printf("failed to load the optional .env_private file, continuing without it: %v", err)
	}

	ocrApp, err := app.NewOCRFlowApp()
	if err != nil {
		log.Fatalf("error initializing: %v", err)
	}
	defer func() {
		if err := ocrApp.Close(); err != nil {
			log.Printf("error closing app: %v", err)
		}
	}()

	if err := validateEditionKeys(ocrApp, keys); err != nil {
		log.Fatal(err)
	}

	defScope := feature.NewEditionDefScope()
	feat, err := ocrApp.Deps.FeatureSvc.GetFeatureInScope(defScope, cfg.featureID, nil)
	if err != nil {
		log.Fatalf("feature lookup failed for %s: %v", cfg.featureID, err)
	}
	rev, err := ocrApp.Deps.FeatureRevisionSvc.GetFeatureRevisionInScope(defScope, cfg.featureID, cfg.revisionID)
	if err != nil {
		log.Fatalf("revision lookup failed for %s/%s: %v", cfg.featureID, cfg.revisionID, err)
	}

	completed, err := loadCompletedKeys(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if !cfg.resume && cfg.outputCSV != "" {
		if err := resetCSV(cfg.outputCSV); err != nil {
			log.Fatalf("initialize output CSV: %v", err)
		}
	}

	applyItems := []feature.ExecutionApplyItem{{
		Feature:  cfg.featureID,
		Revision: cfg.revisionID,
	}}
	policy := (*feature.ExecutionPolicy)(nil)
	if cfg.skipExisting {
		policy = &feature.ExecutionPolicy{
			SkipIf: []feature.ExecutionSkipIf{feature.ExecutionSkipIfRevisionExist},
		}
	}

	totalResults := 0
	completedThisRun := 0
	fmt.Printf("Running edition classification revision %s for %d selected keys\n", cfg.revisionID, len(keys))
	for i, key := range keys {
		if _, ok := completed[key]; ok {
			fmt.Printf("[%d/%d] skipping completed key %s\n", i+1, len(keys), key)
			continue
		}

		exec := &feature.Execution{
			Scope:  feature.NewEditionExecScope(),
			Keys:   []string{key},
			Apply:  applyItems,
			Policy: policy,
		}

		fmt.Printf("[%d/%d] running key %s\n", i+1, len(keys), key)
		results, err := ocrApp.Deps.FeatureExecutionSvc.ExecuteEphemeral(exec, []*feature.Revision{rev}, []*feature.Feature{feat})
		if err != nil {
			log.Fatalf("execution failed for key %s: %v\nResume with the same command; completed keys are recorded in %s", key, err, checkpointPath(cfg))
		}

		if cfg.outputCSV != "" {
			if err := appendResultsCSV(cfg.outputCSV, results); err != nil {
				log.Fatalf("append output CSV for key %s: %v", key, err)
			}
		}

		if !cfg.dryRun {
			if err := ocrApp.Deps.FeatureResultSvc.CreateResults(results, false); err != nil {
				log.Fatalf("store results for key %s: %v", key, err)
			}
		}

		if err := markCompletedKey(cfg, key); err != nil {
			log.Fatalf("write checkpoint for key %s: %v", key, err)
		}
		totalResults += len(results)
		completedThisRun++
		fmt.Printf("[%d/%d] completed key %s with %d result rows\n", i+1, len(keys), key, len(results))
	}

	if cfg.outputCSV != "" {
		fmt.Printf("Wrote result preview CSV: %s\n", cfg.outputCSV)
	}
	if cfg.dryRun {
		fmt.Printf("Dry run complete: %d keys completed in this run, %d result rows produced, none written to DB\n", completedThisRun, totalResults)
		return
	}
	fmt.Printf("Stored %d result rows for feature %s revision %s\n", totalResults, cfg.featureID, cfg.revisionID)
}

func parseFlags() cliConfig {
	var cfg cliConfig
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfg.keys, "keys", "", "selected edition keys, separated by commas or whitespace")
	fs.StringVar(&cfg.keysFile, "keys-file", "", "file containing selected edition keys, separated by commas or whitespace")
	fs.StringVar(&cfg.featureID, "feature", defaultFeatureID, "edition feature id to run")
	fs.StringVar(&cfg.revisionID, "revision", defaultRevisionID, "feature revision id to run")
	fs.StringVar(&cfg.outputCSV, "output-csv", "", "optional CSV preview path for produced results")
	fs.StringVar(&cfg.checkpointFile, "checkpoint-file", "", "path to resume checkpoint file; defaults to output CSV path plus .done")
	fs.BoolVar(&cfg.dryRun, "dry-run", false, "run the LLM and print/export results without storing them")
	fs.BoolVar(&cfg.resume, "resume", true, "skip keys already recorded in the checkpoint file")
	fs.BoolVar(&cfg.skipExisting, "skip-existing-revision", false, "skip keys whose current stored result already uses the selected revision")
	fs.BoolVar(&cfg.skipExisting, "skip-existing-v7", false, "deprecated alias for -skip-existing-revision")
	fs.Parse(os.Args[1:])
	return cfg
}

func loadKeys(rawKeys, keysFile string) ([]string, error) {
	var all []string
	all = append(all, splitKeys(rawKeys)...)

	if strings.TrimSpace(keysFile) != "" {
		data, err := os.ReadFile(keysFile)
		if err != nil {
			return nil, fmt.Errorf("read keys file: %w", err)
		}
		all = append(all, splitKeys(string(data))...)
	}

	keys := uniqNonEmpty(all)
	if len(keys) == 0 {
		return nil, errors.New("at least one selected edition key is required; use -keys or -keys-file")
	}
	return keys, nil
}

func splitKeys(raw string) []string {
	return strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
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

func validateEditionKeys(ocrApp *app.OCRFlowApp, keys []string) error {
	for _, key := range keys {
		if _, err := ocrApp.Deps.EditionSvc.GetEditionByID(key); err != nil {
			return fmt.Errorf("edition lookup failed for %s: %w", key, err)
		}
	}
	return nil
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
	if !cfg.resume {
		path := checkpointPath(cfg)
		if path != "" {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
		return completed, nil
	}

	path := checkpointPath(cfg)
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

func appendResultsCSV(path string, results []*feature.Result) error {
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
	for _, result := range results {
		if result == nil {
			continue
		}
		if len(result.Values) == 0 {
			if err := w.Write([]string{result.Key, result.FeatureID, result.Source.Id, result.Source.Revision, result.Source.Name, ""}); err != nil {
				return err
			}
			continue
		}
		for _, value := range result.Values {
			if err := w.Write([]string{result.Key, result.FeatureID, result.Source.Id, result.Source.Revision, result.Source.Name, value.Surface}); err != nil {
				return err
			}
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
