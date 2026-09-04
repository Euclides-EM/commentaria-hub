package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/transcriptioncorrector"
	"github.com/joho/godotenv"
)

type stringListFlag []string

func (f *stringListFlag) String() string { return strings.Join(*f, ",") }

func (f *stringListFlag) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("directory must not be empty")
	}
	*f = append(*f, value)
	return nil
}

type cliConfig struct {
	corrector       transcriptioncorrector.Config
	authToken       string
	openAIAPIKey    string
	ollamaBaseURL   string
	ollamaAuthToken string
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env_private")

	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		log.Fatal(err)
	}
	client := llm.NewClient(cfg.openAIAPIKey, cfg.ollamaBaseURL, cfg.ollamaAuthToken)
	if _, err := transcriptioncorrector.Run(cfg.corrector, client); err != nil {
		log.Fatal(err)
	}
}

func parseFlags(args []string) (cliConfig, error) {
	var cfg cliConfig
	var markdownDirs, altoDirs stringListFlag
	fs := flag.NewFlagSet("transcription-corrector", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Var(&markdownDirs, "markdown-dir", "input transcription directory; repeat for multiple transcriptions")
	fs.Var(&altoDirs, "alto-dir", "input directory containing page-NNNN ALTO XML files; repeat for multiple transcriptions")
	fs.StringVar(&cfg.corrector.ImagesDir, "images-dir", "", "directory containing page-NNNN images")
	fs.StringVar(&cfg.corrector.OutputDir, "output-dir", "", "directory for corrected markdown")
	fs.IntVar(&cfg.corrector.Rounds, "rounds", transcriptioncorrector.DefaultRounds, "number of correction rounds (minimum 1)")
	fs.BoolVar(&cfg.corrector.SkipExisting, "skip-existing", false, "skip pages with existing round output files")
	fs.StringVar(&cfg.corrector.Provider, "ai-provider", "", "LLM provider: openai, ollama, or claude-code")
	fs.StringVar(&cfg.corrector.Model, "ai-model", "", "LLM model name")
	fs.StringVar(&cfg.authToken, "auth-token", "", "provider auth token (OpenAI API key or Ollama bearer token)")
	fs.StringVar(&cfg.openAIAPIKey, "openai-api-key", "", "OpenAI API key; defaults to OPENAI_API_KEY")
	fs.StringVar(&cfg.ollamaBaseURL, "ollama-base-url", "", "Ollama server URL; defaults to OLLAMA_BASE_URL")
	fs.StringVar(&cfg.ollamaAuthToken, "ollama-auth-token", "", "Ollama bearer token; defaults to OLLAMA_AUTH_TOKEN")
	if err := fs.Parse(args); err != nil {
		return cliConfig{}, err
	}

	markdownDirs = append(markdownDirs, fs.Args()...)
	cfg.corrector.MarkdownDirs = compactStrings(markdownDirs)
	cfg.corrector.ALTODirs = compactStrings(altoDirs)
	cfg.corrector.Provider = strings.ToLower(strings.TrimSpace(cfg.corrector.Provider))
	cfg.corrector.Model = strings.TrimSpace(cfg.corrector.Model)
	cfg.corrector.ImagesDir = strings.TrimSpace(cfg.corrector.ImagesDir)
	cfg.corrector.OutputDir = strings.TrimSpace(cfg.corrector.OutputDir)
	if err := validateCLIConfig(cfg); err != nil {
		return cliConfig{}, err
	}
	resolveCredentials(&cfg)
	return cfg, nil
}

func validateCLIConfig(cfg cliConfig) error {
	if len(cfg.corrector.MarkdownDirs)+len(cfg.corrector.ALTODirs) == 0 {
		return errors.New("at least one -markdown-dir, positional markdown directory, or -alto-dir is required")
	}
	if cfg.corrector.ImagesDir == "" {
		return errors.New("-images-dir is required")
	}
	if cfg.corrector.OutputDir == "" {
		return errors.New("-output-dir is required")
	}
	if cfg.corrector.Rounds < 1 {
		return errors.New("-rounds must be at least 1")
	}
	if !slices.Contains([]string{llm.ProviderOpenAI, llm.ProviderOllama, llm.ProviderClaudeCode}, cfg.corrector.Provider) {
		return fmt.Errorf("unsupported -ai-provider %q (use openai, ollama, or claude-code)", cfg.corrector.Provider)
	}
	if cfg.corrector.Model == "" {
		return errors.New("-ai-model is required")
	}
	return nil
}

func resolveCredentials(cfg *cliConfig) {
	if cfg.openAIAPIKey == "" {
		if cfg.corrector.Provider == llm.ProviderOpenAI {
			cfg.openAIAPIKey = cfg.authToken
		}
		if cfg.openAIAPIKey == "" {
			cfg.openAIAPIKey = os.Getenv("OPENAI_API_KEY")
		}
	}
	if cfg.ollamaBaseURL == "" {
		cfg.ollamaBaseURL = os.Getenv("OLLAMA_BASE_URL")
	}
	if cfg.ollamaAuthToken == "" {
		if cfg.corrector.Provider == llm.ProviderOllama {
			cfg.ollamaAuthToken = cfg.authToken
		}
		if cfg.ollamaAuthToken == "" {
			cfg.ollamaAuthToken = os.Getenv("OLLAMA_AUTH_TOKEN")
		}
	}
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
