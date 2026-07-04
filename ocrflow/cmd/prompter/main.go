package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/app"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	"github.com/joho/godotenv"
)

var allowedModelsByProvider = map[string][]string{
	string(feature.AIProviderOpenAI): {"gpt-5-mini"},
	string(feature.AIProviderOllama): {"gpt-oss:120b"},
}

type cliConfig struct {
	scope         string
	datasetID     string
	datasetSet    bool
	annotationID  string
	annotationSet bool
	featureID     string
	revisionName  string
	revisionDesc  string
	aiProvider    string
	aiModel       string
	prompts       []string
	keys          string
}

func main() {
	log.SetFlags(0)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}
	if err := godotenv.Load(".env_private"); err != nil {
		log.Printf("failed to load the optional .env_private file, continuing without it: %v", err)
	}

	cfg := parseFlags()

	ocrApp, err := app.NewOCRFlowApp()
	if err != nil {
		log.Fatalf("error initializing: %v", err)
	}
	defer ocrApp.Close()

	if ocrApp.Deps == nil {
		log.Fatal("app dependencies are not available")
	}

	reader := bufio.NewReader(os.Stdin)

	if err := fillMissingInputs(reader, ocrApp, &cfg); err != nil {
		log.Fatalf("input error: %v", err)
	}

	defScope, execScope, keys, targetLabel, err := buildExecutionTarget(ocrApp, cfg)
	if err != nil {
		log.Fatalf("target error: %v", err)
	}

	var feats []*feature.Feature
	var revs []*feature.Revision
	var applyItems []feature.ExecutionApplyItem

	for i, rawInput := range cfg.prompts {
		featName := fmt.Sprintf("Ephemeral feature #%d", i+1)
		promptText := rawInput

		fID := strings.TrimSpace(cfg.featureID)
		if fID == "" || i > 0 {
			fID = idgen.GenerateID("fea")
		}
		rID := idgen.GenerateID("rev")

		f := &feature.Feature{
			Scope:       defScope,
			IsList:      true,
			FeatureName: featName,
			Color:       "#000000",
			Properties:  nil,
		}
		f.ID = fID
		f.Name = featName

		r := &feature.Revision{
			Scope:      defScope,
			FeatureID:  fID,
			Prompt:     promptText,
			AIProvider: feature.AIProvider(strings.TrimSpace(cfg.aiProvider)),
			AIModel:    strings.TrimSpace(cfg.aiModel),
		}
		r.ID = rID
		r.Name = strings.TrimSpace(cfg.revisionName)
		r.Description = strings.TrimSpace(cfg.revisionDesc)

		feats = append(feats, f)
		revs = append(revs, r)
		applyItems = append(applyItems, feature.ExecutionApplyItem{
			Feature:  fID,
			Revision: rID,
		})
	}

	exec := &feature.Execution{
		Scope: execScope,
		Keys:  keys,
		Apply: applyItems,
	}
	fmt.Printf("Running ephemeral execution for %s\n", targetLabel)
	results, err := ocrApp.Deps.FeatureExecutionSvc.ExecuteEphemeral(exec, revs, feats)
	if err != nil {
		log.Fatalf("ephemeral execution failed: %v", err)
	}

	printResults(results, feats)
}

func parseFlags() cliConfig {
	var cfg cliConfig

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfg.scope, "scope", "", "feature scope: dataset or editions")
	fs.StringVar(&cfg.datasetID, "dataset", "", "dataset ID for dataset scope")
	fs.StringVar(&cfg.annotationID, "annotation", "", "annotation ID for dataset scope")
	fs.StringVar(&cfg.featureID, "feature-id", "", "ephemeral feature ID override")
	fs.StringVar(&cfg.revisionName, "revision-name", "", "ephemeral revision name")
	fs.StringVar(&cfg.revisionDesc, "revision-description", "", "ephemeral revision description")
	fs.StringVar(&cfg.aiProvider, "ai-provider", "", "AI provider: openai or ollama")
	fs.StringVar(&cfg.aiModel, "ai-model", "", "AI model")
	var promptFlag string
	fs.StringVar(&promptFlag, "prompt", "", "prompt-based revision definition")
	fs.StringVar(&cfg.keys, "keys", "", "comma-separated execution keys; default is inferred from target")

	fs.Parse(os.Args[1:])
	if strings.TrimSpace(promptFlag) != "" {
		cfg.prompts = append(cfg.prompts, promptFlag)
	}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "dataset":
			cfg.datasetSet = true
		case "annotation":
			cfg.annotationSet = true
		}
	})
	return cfg
}

func fillMissingInputs(reader *bufio.Reader, ocrApp *app.OCRFlowApp, cfg *cliConfig) error {
	var err error

	cfg.scope = normalizeScope(cfg.scope)
	if cfg.scope == "" {
		cfg.scope, err = promptChoice(reader, "Scope (dataset/editions)", []string{string(feature.ScopeTypeDataset), string(feature.ScopeTypeEditions)})
		if err != nil {
			return err
		}
	}

	switch cfg.scope {
	case string(feature.ScopeTypeDataset):
		if !cfg.datasetSet {
			cfg.datasetID, err = prompt(reader, "Dataset ID", "tps")
			if err != nil {
				return err
			}
		}
		if strings.TrimSpace(cfg.datasetID) == "" {
			cfg.datasetID = "tps"
		}
		if _, err := ocrApp.Deps.DatasetSvc.Get(cfg.datasetID); err != nil {
			return fmt.Errorf("dataset lookup failed: %w", err)
		}
		if !cfg.annotationSet {
			defaultAnnotationID := ""
			if cfg.datasetID == "tps" {
				defaultAnnotationID = "ann_1"
			}
			cfg.annotationID, err = prompt(reader, "Annotation ID", defaultAnnotationID)
			if err != nil {
				return err
			}
		}
		if strings.TrimSpace(cfg.annotationID) == "" {
			return fmt.Errorf("annotation ID is required")
		}
		if _, err := ocrApp.Deps.AnnotationSvc.Get(cfg.datasetID, cfg.annotationID); err != nil {
			return fmt.Errorf("annotation lookup failed: %w", err)
		}
	case string(feature.ScopeTypeEditions):
		if cfg.keys == "" {
			cfg.keys, err = promptNonEmpty(reader, "Edition IDs (comma-separated)")
			if err != nil {
				return err
			}
		}
		keys := parseCSV(cfg.keys)
		if len(keys) == 0 {
			return fmt.Errorf("edition IDs are required")
		}
		for _, editionID := range keys {
			if _, err := ocrApp.Deps.EditionSvc.GetEditionByID(editionID); err != nil {
				return fmt.Errorf("edition lookup failed for %s: %w", editionID, err)
			}
		}
	default:
		return fmt.Errorf("invalid scope %q", cfg.scope)
	}

	if cfg.aiProvider == "" {
		cfg.aiProvider, err = promptChoice(reader, "AI provider (openai/ollama)", []string{string(feature.AIProviderOllama), string(feature.AIProviderOpenAI)})
		if err != nil {
			return err
		}
	}
	cfg.aiProvider = strings.TrimSpace(strings.ToLower(cfg.aiProvider))
	if cfg.aiModel == "" {
		cfg.aiModel, err = promptChoice(reader, "AI model", allowedModelsByProvider[cfg.aiProvider])
		if err != nil {
			return err
		}
	}
	if len(cfg.prompts) == 0 {
		fmt.Println("Enter prompt text. Multiline input is supported. Finish with a line containing only END:")
		v, err := promptBlock(reader)
		if err != nil {
			return err
		}
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("at least one prompt is required")
		}
		cfg.prompts = append(cfg.prompts, v)
	}
	if cfg.keys == "" {
		cfg.keys, err = prompt(reader, "Execution keys (comma-separated, blank to use inferred default)", "")
		if err != nil {
			return err
		}
	}
	cfg.aiModel = strings.TrimSpace(cfg.aiModel)
	if err := validateAIModel(cfg.aiProvider, cfg.aiModel); err != nil {
		return err
	}
	return nil
}

func buildExecutionTarget(ocrApp *app.OCRFlowApp, cfg cliConfig) (feature.DefScope, feature.ExecScope, []string, string, error) {
	switch cfg.scope {
	case string(feature.ScopeTypeDataset):
		ann, err := ocrApp.Deps.AnnotationSvc.Get(cfg.datasetID, cfg.annotationID)
		if err != nil {
			return feature.DefScope{}, feature.ExecScope{}, nil, "", err
		}
		keys := parseCSV(cfg.keys)
		if len(keys) == 0 {
			keys, err = pagesparser.Range(ann.Pages)
			if err != nil {
				return feature.DefScope{}, feature.ExecScope{}, nil, "", fmt.Errorf("infer annotation keys: %w", err)
			}
		}
		if len(keys) == 0 {
			return feature.DefScope{}, feature.ExecScope{}, nil, "", errors.New("no execution keys available for annotation")
		}
		return feature.NewDatasetDefScope(cfg.datasetID), feature.NewDatasetExecScope(cfg.datasetID, cfg.annotationID), keys, fmt.Sprintf("annotation %s in dataset %s", cfg.annotationID, cfg.datasetID), nil
	case string(feature.ScopeTypeEditions):
		keys := parseCSV(cfg.keys)
		if len(keys) == 0 {
			return feature.DefScope{}, feature.ExecScope{}, nil, "", errors.New("edition IDs are required")
		}
		return feature.NewEditionDefScope(), feature.NewEditionExecScope(), keys, fmt.Sprintf("edition(s) %s", strings.Join(keys, ",")), nil
	default:
		return feature.DefScope{}, feature.ExecScope{}, nil, "", fmt.Errorf("invalid scope %q", cfg.scope)
	}
}

func normalizeScope(v string) string {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case string(feature.ScopeTypeDataset):
		return string(feature.ScopeTypeDataset)
	case string(feature.ScopeTypeEditions):
		return string(feature.ScopeTypeEditions)
	default:
		return ""
	}
}

func parseCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validateAIModel(provider, model string) error {
	allowed, ok := allowedModelsByProvider[provider]
	if !ok {
		return fmt.Errorf("invalid ai_provider %q", provider)
	}
	for _, candidate := range allowed {
		if model == candidate {
			return nil
		}
	}
	return fmt.Errorf("invalid ai_model %q for provider %q", model, provider)
}

func printResults(results []*feature.Result, feats []*feature.Feature) {
	if len(results) == 0 {
		fmt.Println("No results.")
		return
	}

	nameByFeatureID := make(map[string]string, len(feats))
	for _, f := range feats {
		nameByFeatureID[f.ID] = f.FeatureName
	}

	keyOrder := make([]string, 0)
	byKey := make(map[string][]*feature.Result)
	for _, result := range results {
		if _, seen := byKey[result.Key]; !seen {
			keyOrder = append(keyOrder, result.Key)
		}
		byKey[result.Key] = append(byKey[result.Key], result)
	}

	for _, key := range keyOrder {
		fmt.Printf("%s\n", key)
		for _, result := range byKey[key] {
			fmt.Printf("  [%s]\n", nameByFeatureID[result.FeatureID])
			if len(result.Values) == 0 {
				fmt.Println("    (no values)")
				continue
			}
			for _, value := range result.Values {
				fmt.Printf("    %s\n", value.Surface)
			}
		}
	}
}

func prompt(reader *bufio.Reader, label, defaultValue string) (string, error) {
	if defaultValue == "" {
		fmt.Printf("%s: ", label)
	} else {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	}
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return defaultValue, nil
	}
	return text, nil
}

func promptBlock(reader *bufio.Reader) (string, error) {
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if len(lines) == 0 {
				return "", err
			}
			lines = append(lines, strings.TrimRight(line, "\r\n"))
			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "END" {
			break
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), nil
}

func promptNonEmpty(reader *bufio.Reader, label string) (string, error) {
	for {
		v, err := prompt(reader, label, "")
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(v) != "" {
			return v, nil
		}
		fmt.Println("Value is required.")
	}
}

func promptChoice(reader *bufio.Reader, label string, options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no options available")
	}
	for {
		v, err := prompt(reader, fmt.Sprintf("%s (%s)", label, strings.Join(options, "/")), options[0])
		if err != nil {
			return "", err
		}
		v = strings.TrimSpace(v)
		for _, option := range options {
			if v == option {
				return v, nil
			}
		}
		fmt.Printf("Invalid value. Allowed: %s\n", strings.Join(options, ", "))
	}
}
