package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/app"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/joho/godotenv"
)

var allowedModelsByProvider = map[string][]string{
	string(feature.AIProviderOpenAI): {"gpt-5-mini"},
	string(feature.AIProviderOllama): {"gpt-oss:120b"},
}

type cliConfig struct {
	scope           string
	datasetID       string
	datasetSet      bool
	annotationID    string
	annotationSet   bool
	featureID       string
	revisionName    string
	revisionDesc    string
	aiProvider      string
	aiModel         string
	prompt          string
	keys            string
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

	featureID := strings.TrimSpace(cfg.featureID)
	if featureID == "" {
		featureID = idgen.GenerateID("fea")
	}
	revisionID := idgen.GenerateID("rev")
	feat := &feature.Feature{
		Scope:      defScope,
		IsList:     true,
		Color:      "#000000",
		Properties: nil,
	}
	feat.ID = featureID
	feat.Name = ephemeralFeatureName(cfg, featureID)

	rev := &feature.Revision{
		Scope:       defScope,
		FeatureID:   featureID,
		Prompt:      strings.TrimSpace(cfg.prompt),
		AIProvider:  feature.AIProvider(strings.TrimSpace(cfg.aiProvider)),
		AIModel:     strings.TrimSpace(cfg.aiModel),
	}
	rev.ID = revisionID
	rev.Name = strings.TrimSpace(cfg.revisionName)
	rev.Description = strings.TrimSpace(cfg.revisionDesc)

	exec := &feature.Execution{
		Scope: execScope,
		Keys:  keys,
		Apply: []feature.ExecutionApplyItem{{
			Feature:  featureID,
			Revision: revisionID,
		}},
	}
	fmt.Printf("Running ephemeral execution for %s\n", targetLabel)
	results, err := ocrApp.Deps.FeatureExecutionSvc.ExecuteEphemeral(exec, []*feature.Revision{rev}, []*feature.Feature{feat})
	if err != nil {
		log.Fatalf("ephemeral execution failed: %v", err)
	}

	if cfg.scope == string(feature.ScopeTypeEditions) {
		fmt.Println("Edition execution is currently stubbed in the service layer and does not produce results yet.")
	}
	printResults(results)
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
	fs.StringVar(&cfg.prompt, "prompt", "", "prompt-based revision definition")
	fs.StringVar(&cfg.keys, "keys", "", "comma-separated execution keys; default is inferred from target")

	fs.Parse(os.Args[1:])
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
	if strings.TrimSpace(cfg.prompt) == "" {
		cfg.prompt, err = promptNonEmpty(reader, "Prompt")
		if err != nil {
			return err
		}
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

func ephemeralFeatureName(cfg cliConfig, featureID string) string {
	target := cfg.datasetID
	if cfg.scope == string(feature.ScopeTypeEditions) {
		target = strings.ReplaceAll(cfg.keys, ",", "_")
	}
	return fmt.Sprintf("tmp-cli-%s-%s", target, featureID)
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

func printResults(results []*feature.Result) {
	if len(results) == 0 {
		fmt.Println("No results.")
		return
	}

	for _, result := range results {
		fmt.Printf("%s\n", result.Key)
		if len(result.Values) == 0 {
			fmt.Println("  (no values)")
			continue
		}
		for _, value := range result.Values {
			fmt.Printf("  %s\n", value.Surface)
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
