package config

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/samber/lo"
)

type EnvConfig struct {
	HTTPAddr                 string        `env:"HTTP_ADDR" envDefault:":8085"`
	GithubToken              string        `env:"GITHUB_TOKEN"`
	GithubDownloaderTimeout  time.Duration `env:"GITHUB_DOWNLOADER_TIMEOUT" envDefault:"5m"`
	DatasetCreateMaxParallel int           `env:"DATASET_CREATE_MAX_PARALLEL"`
	DatasetCreateQueueWait   time.Duration `env:"DATASET_CREATE_QUEUE_WAIT_TIMEOUT"`
	PythonExecutable         string        `env:"PYTHON_EXECUTABLE" envDefault:"python"`
	RoboflowAPIKey           string        `env:"ROBOFLOW_API_KEY"`
	EscriptoriumBasePath     string        `env:"ESCRIPTORIUM_BASE_PATH" envDefault:"http://localhost:8080/"`
	EscriptoriumUsername     string        `env:"ESCRIPTORIUM_USERNAME" envDefault:"admin"`
	EscriptoriumPassword     string        `env:"ESCRIPTORIUM_PASSWORD" envDefault:"admin"`
	CommentariaPath          string        `env:"COMMENTARIA_PATH" envDefault:"https://euclides.huma-num.fr/commentaria"`
	AllowedOriginsCORS       string        `env:"ALLOWED_ORIGINS_CORS" envDefault:""`
	LogsSystemdUnit          string        `env:"LOGS_SYSTEMD_UNIT" envDefault:"commentaria-hub-api"`
	LogsTailDefaultLines     int           `env:"LOGS_TAIL_DEFAULT_LINES" envDefault:"200"`
	LogsTailMaxLines         int           `env:"LOGS_TAIL_MAX_LINES" envDefault:"2000"`
	GPUFarmHost              string        `env:"GPU_FARM_HOST" envDefault:""`
	GPUFarmJobRoot           string        `env:"GPU_FARM_JOB_ROOT" envDefault:""`
	APIURL                   string        `env:"API_URL" envDefault:""`

	RootDir          string `env:"ROOT_DIR" envDefault:"../"`
	StoreDir         string `env:"STORE_DIR" envDefault:"./store"`
	TempDir          string `env:"OCRFLOW_TEMP_DIR"`
	BackupRootDir    string `env:"BACKUP_ROOT_DIR" envDefault:"./full_backups"`
	BackupMaxToStore int    `env:"BACKUP_MAX_TO_STORE" envDefault:"5"`

	FacsimilesPDFDir         string `env:"FACSIMILES_PDF_DIR" envDefault:""`
	FacsimilesDiagramsPath   string `env:"FACSIMILES_DIAGRAMS_PATH" envDefault:"docs/diagrams"`
	FacsimilesDiagramsURL    string `env:"FACSIMILES_DIAGRAMS_URL" envDefault:""`
	FacsimilesGDriveFolderID string `env:"FACSIMILES_GDRIVE_FOLDER_ID" envDefault:""`
	FacsimilesRemoteAPIURL   string `env:"FACSIMILES_REMOTE_API_URL" envDefault:""`
	OpenAIAPIKey             string `env:"OPENAI_API_KEY"`
	OllamaBaseURL            string `env:"OLLAMA_BASE_URL" envDefault:""`
	OllamaAuthToken          string `env:"OLLAMA_AUTH_TOKEN" envDefault:""`

	SkipDiagramCropsGeneration bool     `env:"SKIP_DIAGRAM_CROPS_GENERATION" envDefault:"false"`
	OptMigrations              []string `env:"OPT_MIGRATIONS" envDefault:""`

	RcloneRemoteName     string `env:"RCLONE_REMOTE_NAME" envDefault:"G"`
	BackupGDriveFolderID string `env:"BACKUP_GDRIVE_FOLDER_ID" envDefault:""`
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

func (ec *EnvConfig) DefaultModelsDir() string {
	return filepath.Join(ec.RootDir, "ocrflow", "store", "default_models")
}

func (ec *EnvConfig) TmpDir() string {
	return ec.TempDir
}

func (ec *EnvConfig) AllowedOriginsCORSList() []string {
	corsOrigins := ec.defaultAllowedOriginsCORS()
	if ec.AllowedOriginsCORS != "" {
		lo.ForEach(strings.Split(ec.AllowedOriginsCORS, ","), func(origin string, _ int) {
			corsOrigins = append(corsOrigins, strings.TrimSpace(origin))
		})
	}
	return corsOrigins
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

func (ec *EnvConfig) SyncBackupToRemoteByDefault() bool {
	return ec.BackupGDriveFolderID != ""
}
