package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/app"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
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
	annotationID    string
	editionID       string
	featureID       string
	revisionName    string
	revisionDesc    string
	aiProvider      string
	aiModel         string
	prompt          string
	categorizer     string
	keys            string
	pushToOrigin    bool
	pushToOriginSet bool
	skipIf          string
	wait            time.Duration
	poll            time.Duration
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

	if _, err := ocrApp.Deps.FeatureSvc.GetFeatureInScope(defScope, cfg.featureID, nil); err != nil {
		log.Fatalf("feature lookup failed: %v", err)
	}

	rev := &feature.Revision{
		Scope:       defScope,
		FeatureID:   cfg.featureID,
		Prompt:      strings.TrimSpace(cfg.prompt),
		Categorizer: strings.TrimSpace(cfg.categorizer),
		AIProvider:  feature.AIProvider(strings.TrimSpace(cfg.aiProvider)),
		AIModel:     strings.TrimSpace(cfg.aiModel),
	}
	rev.Name = strings.TrimSpace(cfg.revisionName)
	rev.Description = strings.TrimSpace(cfg.revisionDesc)

	fmt.Printf("Creating revision for feature %s in scope %s\n", cfg.featureID, cfg.scope)
	createdRev, err := ocrApp.Deps.FeatureRevisionSvc.CreateFeatureRevision(cfg.featureID, rev)
	if err != nil {
		log.Fatalf("failed to create revision: %v", err)
	}

	exec := &feature.Execution{
		Scope: execScope,
		Keys:  keys,
		Apply: []feature.ExecutionApplyItem{{
			Feature:  cfg.featureID,
			Revision: createdRev.ID,
		}},
	}
	if cfg.pushToOriginSet || len(parseCSV(cfg.skipIf)) > 0 {
		skipIf, err := validateSkipIfList(cfg.skipIf)
		if err != nil {
			log.Fatalf("invalid skip policy: %v", err)
		}
		exec.Policy = &feature.ExecutionPolicy{
			PushToOrigin: cfg.pushToOrigin,
			SkipIf:       skipIf,
		}
	}

	fmt.Printf("Creating execution for %s\n", targetLabel)
	createdExec, err := ocrApp.Deps.FeatureExecutionSvc.CreateFeatureExecution(exec)
	if err != nil {
		log.Fatalf("failed to create execution: %v", err)
	}

	fmt.Printf("Revision ID: %s\n", createdRev.ID)
	fmt.Printf("Execution ID: %s\n", createdExec.ID)
	fmt.Printf("Execution status: %s\n", createdExec.Status)

	if cfg.scope == string(feature.ScopeTypeEditions) {
		fmt.Println("Edition execution is currently stubbed in the service layer and does not produce results yet.")
	}

	if cfg.wait > 0 {
		if err := waitForExecution(ocrApp, createdExec.ID, cfg.wait, cfg.poll); err != nil {
			log.Fatalf("execution wait failed: %v", err)
		}
	}
}

func parseFlags() cliConfig {
	var cfg cliConfig
	var pushToOriginValue bool
	var pushToOriginProvided bool

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&cfg.scope, "scope", "", "feature scope: dataset or editions")
	fs.StringVar(&cfg.datasetID, "dataset", "", "dataset ID for dataset scope")
	fs.StringVar(&cfg.annotationID, "annotation", "", "annotation ID for dataset scope")
	fs.StringVar(&cfg.editionID, "edition", "", "edition ID for editions scope")
	fs.StringVar(&cfg.featureID, "feature", "", "feature ID")
	fs.StringVar(&cfg.revisionName, "revision-name", "", "revision name")
	fs.StringVar(&cfg.revisionDesc, "revision-description", "", "revision description")
	fs.StringVar(&cfg.aiProvider, "ai-provider", "", "AI provider: openai or ollama")
	fs.StringVar(&cfg.aiModel, "ai-model", "", "AI model")
	fs.StringVar(&cfg.prompt, "prompt", "", "prompt-based revision definition")
	fs.StringVar(&cfg.categorizer, "categorizer", "", "categorizer-based revision definition")
	fs.StringVar(&cfg.keys, "keys", "", "comma-separated execution keys; default is inferred from target")
	fs.StringVar(&cfg.skipIf, "skip-if", "", "comma-separated skip policy values")
	fs.DurationVar(&cfg.wait, "wait", 30*time.Second, "time to wait for execution completion; 0 disables waiting")
	fs.DurationVar(&cfg.poll, "poll", 1*time.Second, "poll interval while waiting")
	fs.Func("push-to-origin", "set execution push_to_origin policy", func(v string) error {
		parsed, err := parseBoolString(v)
		if err != nil {
			return err
		}
		pushToOriginValue = parsed
		pushToOriginProvided = true
		return nil
	})

	fs.Parse(os.Args[1:])

	cfg.pushToOrigin = pushToOriginValue
	cfg.pushToOriginSet = pushToOriginProvided
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
		if cfg.datasetID == "" {
			cfg.datasetID, err = promptNonEmpty(reader, "Dataset ID")
			if err != nil {
				return err
			}
		}
		if _, err := ocrApp.Deps.DatasetSvc.Get(cfg.datasetID); err != nil {
			return fmt.Errorf("dataset lookup failed: %w", err)
		}
		if cfg.annotationID == "" {
			cfg.annotationID, err = promptNonEmpty(reader, "Annotation ID")
			if err != nil {
				return err
			}
		}
		if _, err := ocrApp.Deps.AnnotationSvc.Get(cfg.datasetID, cfg.annotationID); err != nil {
			return fmt.Errorf("annotation lookup failed: %w", err)
		}
	case string(feature.ScopeTypeEditions):
		if cfg.editionID == "" {
			cfg.editionID, err = promptNonEmpty(reader, "Edition ID")
			if err != nil {
				return err
			}
		}
		if _, err := ocrApp.Deps.EditionSvc.GetEditionByID(cfg.editionID); err != nil {
			return fmt.Errorf("edition lookup failed: %w", err)
		}
	default:
		return fmt.Errorf("invalid scope %q", cfg.scope)
	}

	if cfg.featureID == "" {
		cfg.featureID, err = promptNonEmpty(reader, "Feature ID")
		if err != nil {
			return err
		}
	}
	if cfg.revisionName == "" {
		cfg.revisionName, err = promptNonEmpty(reader, "Revision name")
		if err != nil {
			return err
		}
	}
	if cfg.revisionDesc == "" {
		cfg.revisionDesc, err = prompt(reader, "Revision description", "")
		if err != nil {
			return err
		}
	}
	if cfg.aiProvider == "" {
		cfg.aiProvider, err = promptChoice(reader, "AI provider (openai/ollama)", []string{string(feature.AIProviderOpenAI), string(feature.AIProviderOllama)})
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
	if strings.TrimSpace(cfg.prompt) == "" && strings.TrimSpace(cfg.categorizer) == "" {
		mode, err := promptChoice(reader, "Revision type (prompt/categorizer)", []string{"prompt", "categorizer"})
		if err != nil {
			return err
		}
		if mode == "prompt" {
			cfg.prompt, err = promptNonEmpty(reader, "Prompt")
			if err != nil {
				return err
			}
		} else {
			props := ocrApp.Deps.FeaturePropertySvc.ListFeaturePropertyKeys()
			cfg.categorizer, err = promptChoice(reader, "Categorizer", props)
			if err != nil {
				return err
			}
		}
	}
	if strings.TrimSpace(cfg.prompt) != "" && strings.TrimSpace(cfg.categorizer) != "" {
		return errors.New("prompt and categorizer are mutually exclusive")
	}
	if cfg.keys == "" {
		cfg.keys, err = prompt(reader, "Execution keys (comma-separated, blank to use inferred default)", "")
		if err != nil {
			return err
		}
	}
	if cfg.skipIf == "" {
		cfg.skipIf, err = prompt(reader, "Skip policy values (comma-separated: feature_exist, revision_exist, human_reviewed)", "")
		if err != nil {
			return err
		}
	}
	if !cfg.pushToOriginSet {
		val, err := prompt(reader, "Push results to origin? (true/false, default false)", "false")
		if err != nil {
			return err
		}
		parsed, err := parseBoolString(val)
		if err != nil {
			return err
		}
		cfg.pushToOrigin = parsed
		cfg.pushToOriginSet = true
	}

	cfg.aiModel = strings.TrimSpace(cfg.aiModel)
	if err := validateAIModel(cfg.aiProvider, cfg.aiModel); err != nil {
		return err
	}
	if _, err := validateSkipIfList(cfg.skipIf); err != nil {
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
			keys = []string{cfg.editionID}
		}
		return feature.NewEditionDefScope(), feature.NewEditionExecScope(), keys, fmt.Sprintf("edition %s", cfg.editionID), nil
	default:
		return feature.DefScope{}, feature.ExecScope{}, nil, "", fmt.Errorf("invalid scope %q", cfg.scope)
	}
}

func waitForExecution(ocrApp *app.OCRFlowApp, executionID string, timeout, poll time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		exec, err := ocrApp.Deps.FeatureExecutionSvc.GetFeatureExecution(executionID)
		if err != nil {
			return err
		}
		if exec.Status != feature.ExecutionStatusInProgress && exec.Status != feature.ExecutionStatusCanceling {
			fmt.Printf("Final execution status: %s\n", exec.Status)
			if exec.StatusReason != "" {
				fmt.Printf("Status reason: %s\n", exec.StatusReason)
			}
			return nil
		}
		if time.Now().After(deadline) {
			fmt.Printf("Execution still in progress after %s\n", timeout)
			return nil
		}
		time.Sleep(poll)
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

func validateSkipIfList(v string) ([]feature.ExecutionSkipIf, error) {
	raw := parseCSV(v)
	out := make([]feature.ExecutionSkipIf, 0, len(raw))
	for _, item := range raw {
		switch feature.ExecutionSkipIf(item) {
		case feature.ExecutionSkipIfFeatureExist, feature.ExecutionSkipIfRevisionExist, feature.ExecutionSkipIfHumanReviewed:
			out = append(out, feature.ExecutionSkipIf(item))
		default:
			return nil, fmt.Errorf("invalid skip_if value %q", item)
		}
	}
	return out, nil
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

func parseBoolString(v string) (bool, error) {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "true", "t", "1", "yes", "y":
		return true, nil
	case "false", "f", "0", "no", "n", "":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", v)
	}
}
