package config

import (
	"github.com/caarlos0/env"
)

type EnvConfig struct {
	DBPath        string `env:"DB_PATH"`
	MigrationsDir string `env:"MIGRATIONS_DIR"`
	HTTPAddr      string `env:"HTTP_ADDR" envDefault:":8080"`
}

func InitEnv() (*EnvConfig, error) {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		return nil, err
	}
	return &envConfig, nil
}
