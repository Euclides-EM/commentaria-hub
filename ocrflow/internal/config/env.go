package config

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/samber/lo"
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
	CommentariaPath         string        `env:"COMMENTARIA_PATH" envDefault:"https://euclides.huma-num.fr/commentaria"`
	AllowedOriginsCORS      string        `env:"ALLOWED_ORIGINS_CORS" envDefault:""`

	RootDir       string `env:"ROOT_DIR" envDefault:"./"`
	StoreDir      string `env:"STORE_DIR" envDefault:"./store"`
	TempDir       string `env:"OCRFLOW_TEMP_DIR"`
	BackupRootDir string `env:"BACKUP_ROOT_DIR" envDefault:"./full_backups"`

	FacsimilesGithubRepoUrl string `env:"FACSIMILES_GITHUB_REPO_URL" envDefault:"https://github.com/Euclides-EM/elements-facsimile"`
	FacsimilesDiagramsPath  string `env:"FACSIMILES_DIAGRAMS_PATH" envDefault:"docs/diagrams"`
	OpenAIAPIKey            string `env:"OPENAI_API_KEY"`

	SkipDiagramCropsGeneration bool     `env:"SKIP_DIAGRAM_CROPS_GENERATION" envDefault:"false"`
	OptMigrations              []string `env:"OPT_MIGRATIONS" envDefault:""`
}

func InitEnv() (*EnvConfig, error) {
	var envConfig EnvConfig
	if err := env.Parse(&envConfig); err != nil {
		return nil, err
	}
	return &envConfig, nil
}

func (ec *EnvConfig) OptionalMigrations() []string {
	if len(ec.OptMigrations) == 0 {
		return nil
	}
	return lo.Filter(lo.Map(ec.OptMigrations, func(mig string, _ int) string {
		return strings.TrimSpace(mig)
	}), func(mig string, _ int) bool {
		return mig != ""
	})
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

func (ec *EnvConfig) TmpDir() string {
	return ec.TempDir
}

func (ec *EnvConfig) AllowedOriginsCORSList() []string {
	if ec.AllowedOriginsCORS == "" {
		return ec.defaultAllowedOriginsCORS()
	}
	return lo.Map(strings.Split(ec.AllowedOriginsCORS, ","), func(origin string, _ int) string {
		return strings.TrimSpace(origin)
	})
}

func (ec *EnvConfig) defaultAllowedOriginsCORS() []string {
	domains := []string{
		"euclides.huma-num.fr",
		"elements-resource-box.netlify.app",
		"localhost",
		"127.0.0.1",
	}
	localhostPorts := []string{"3000", "5173", "5174", "5180", "5181", "5190", "5191", "8080"}
	var l []string
	for _, domain := range domains {
		l = append(l, "http://"+domain)
		l = append(l, "https://"+domain)
	}
	for _, port := range localhostPorts {
		l = append(l, "http://localhost:"+port)
		l = append(l, "http://127.0.0.1:"+port)
	}
	return l
}
