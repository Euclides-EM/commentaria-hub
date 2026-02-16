package config

import (
	"time"

	"github.com/caarlos0/env"
)

type EnvConfig struct {
	DBPath string `env:"DB_PATH" envDefault:"./ocrflow.db"`
	// todo: change to embedded migrations ?
	MigrationsDir           string        `env:"MIGRATIONS_DIR" envDefault:"./migrations/ocrflow"`
	DataDir                 string        `env:"DATA_DIR" envDefault:"./store/data"`
	UIDir                   string        `env:"UI_DIR" envDefault:"./ui/dist"`
	HTTPAddr                string        `env:"HTTP_ADDR" envDefault:":8085"`
	GithubToken             string        `env:"GITHUB_TOKEN"`
	GithubDownloaderTimeout time.Duration `env:"GITHUB_DOWNLOADER_TIMEOUT" envDefault:"30s"`
	ModelsDir               string        `env:"MODELS_DIR" envDefault:"./store/models"`
	PythonExecutable        string        `env:"PYTHON_EXECUTABLE" envDefault:"python"`
	RoboflowAPIKey          string        `env:"ROBOFLOW_API_KEY"`
	TrainingDir             string        `env:"TRAINING_DIR" envDefault:"./store/training"`
	EscriptoriumBasePath    string        `env:"ESCRIPTORIUM_BASE_PATH" envDefault:"http://localhost:8080/"`
	EscriptoriumUsername    string        `env:"ESCRIPTORIUM_USERNAME" envDefault:"admin"`
	EscriptoriumPassword    string        `env:"ESCRIPTORIUM_PASSWORD" envDefault:"admin"`
	ItemsMetadataStoreDir   string        `env:"ITEMS_METADATA_STORE_DIR" envDefault:"./store/items_metadata"`
	FacsimilesGithubRepoUrl string        `env:"FACSIMILES_GITHUB_REPO_URL" envDefault:"https://github.com/ReallyLiri/elements-facsimile/blob/main/docs"`
}

func InitEnv() (*EnvConfig, error) {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		return nil, err
	}
	return &envConfig, nil
}
