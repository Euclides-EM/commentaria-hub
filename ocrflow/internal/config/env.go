package config

import (
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
)

type EnvConfig struct {
	HTTPAddr                string        `env:"HTTP_ADDR" envDefault:":8085"`
	GithubToken             string        `env:"GITHUB_TOKEN"`
	GithubDownloaderTimeout time.Duration `env:"GITHUB_DOWNLOADER_TIMEOUT" envDefault:"5m"`
	PythonExecutable        string        `env:"PYTHON_EXECUTABLE" envDefault:"python"`
	RoboflowAPIKey          string        `env:"ROBOFLOW_API_KEY"`
	EscriptoriumBasePath    string        `env:"ESCRIPTORIUM_BASE_PATH" envDefault:"http://localhost:8080/"`
	EscriptoriumUsername    string        `env:"ESCRIPTORIUM_USERNAME" envDefault:"admin"`
	EscriptoriumPassword    string        `env:"ESCRIPTORIUM_PASSWORD" envDefault:"admin"`
	CommentariaPath         string        `env:"COMMENTARIA_PATH" envDefault:"http://euclides.huma-num.fr/commentaria"`

	StoreDir      string `env:"STORE_DIR" envDefault:"./store"`
	BackupRootDir string `env:"BACKUP_ROOT_DIR" envDefault:"./full_backups"`

	FacsimilesGithubRepoUrl string `env:"FACSIMILES_GITHUB_REPO_URL" envDefault:"https://github.com/Euclides-EM/elements-facsimile"`
	FacsimilesDiagramsPath  string `env:"FACSIMILES_DIAGRAMS_PATH" envDefault:"docs/diagrams"`
	OpenAIAPIKey            string `env:"OPENAI_API_KEY"`
}

func InitEnv() (*EnvConfig, error) {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		return nil, err
	}
	return &envConfig, nil
}

func (ec *EnvConfig) ItemsMetadataStoreDir() string {
	return filepath.Join(ec.StoreDir, "items_metadata")
}

func (ec *EnvConfig) DiagramsDir() string {
	return filepath.Join(ec.StoreDir, "diagrams")
}

func (ec *EnvConfig) TrainingDir() string {
	return filepath.Join(ec.StoreDir, "training_data")
}

func (ec *EnvConfig) ModelsDir() string {
	return filepath.Join(ec.StoreDir, "models")
}

func (ec *EnvConfig) DataDir() string {
	return filepath.Join(ec.StoreDir, "data")
}

func (ec *EnvConfig) DBPath() string {
	return filepath.Join(ec.StoreDir, "ocrflow.db")
}

func (ec *EnvConfig) BackupDir() string {
	return filepath.Join(ec.BackupRootDir, "backups")
}

func (ec *EnvConfig) RestoreDir() string {
	return filepath.Join(ec.BackupRootDir, "restore")
}
