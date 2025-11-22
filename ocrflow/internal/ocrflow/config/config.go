package config

import (
	"github.com/caarlos0/env"
	"time"
)

type EnvConfig struct {
	DBPath string `env:"DB_PATH" envDefault:"./ocrflow.db"`
	// todo: change to embedded migrations ?
	MigrationsDir           string        `env:"MIGRATIONS_DIR" envDefault:"./migrations/ocrflow"`
	FacsimilesDir           string        `env:"FACSIMILES_DIR" envDefault:"./facsimiles"`
	HTTPAddr                string        `env:"HTTP_ADDR" envDefault:":8080"`
	GithubToken             string        `env:"GITHUB_TOKEN"`
	GithubDownloaderTimeout time.Duration `env:"GITHUB_DOWNLOADER_TIMEOUT" envDefault:"30s"`
}

func InitEnv() (*EnvConfig, error) {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		return nil, err
	}
	return &envConfig, nil
}
