// gen_diagram_crops is a CLI to regenerate diagram crops metadata (directory list and per-edition JSON).
// The same logic runs automatically at app startup; use this script for one-off or dry-run from the repo.
package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/internal/config"
	"github.com/MiaMish/elements-dh/ocrflow/internal/diagramcrops"
	"github.com/joho/godotenv"
)

func main() {
	var dryRun bool
	flag.BoolVar(&dryRun, "dry-run", false, "print actions without writing files")
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		log.Fatalf("failed to locate repo root: %v", err)
	}

	envConfig, err := loadEnvConfig(repoRoot)
	if err != nil {
		log.Fatalf("failed to load env config: %v", err)
	}

	if err := diagramcrops.Generate(envConfig, diagramcrops.Options{DryRun: dryRun}); err != nil {
		log.Fatalf("failed to generate diagram crops: %v", err)
	}
}

func loadEnvConfig(repoRoot string) (*config.EnvConfig, error) {
	envFiles, err := resolveEnvFiles(repoRoot)
	if err != nil {
		return nil, err
	}

	for _, envFile := range envFiles {
		if err := godotenv.Overload(envFile); err != nil {
			return nil, err
		}
	}

	return config.InitEnv()
}

func resolveEnvFiles(repoRoot string) ([]string, error) {
	parentRoot := filepath.Dir(repoRoot)
	candidates := uniquePaths([]string{
		filepath.Join(repoRoot, ".env"),
		filepath.Join(repoRoot, ".env_private"),
		filepath.Join(parentRoot, ".env"),
		filepath.Join(parentRoot, ".env_private"),
	})

	var envFiles []string
	for _, envFile := range candidates {
		_, err := os.Stat(envFile)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, err
		}
		envFiles = append(envFiles, envFile)
	}

	return envFiles, nil
}

func uniquePaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func findRepoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		next := filepath.Dir(current)
		if next == current {
			return "", errors.New("go.mod not found")
		}
		current = next
	}
}
