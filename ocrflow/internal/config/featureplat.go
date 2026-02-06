package config

import (
	"github.com/caarlos0/env"
)

type FeatureAppEnvConfig struct {
	DBPath string `env:"DB_PATH" envDefault:"./features.db"`
	// todo: change to embedded migrations ?
	MigrationsDir string `env:"MIGRATIONS_DIR" envDefault:"./migrations/featureapp"`
	HTTPAddr      string `env:"HTTP_ADDR" envDefault:":8086"`
}

func InitFeatureAppEnvConfig() (*FeatureAppEnvConfig, error) {
	var c FeatureAppEnvConfig
	if err := env.Parse(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
