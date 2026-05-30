package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/gpufarm"
)

type OCRModelTraining struct {
	models      *Model
	fileSysMgt  *filesys.Manager
	datasets    *Dataset
	annotations *Annotation
	submitter   gpufarm.Submitter

	rootDir string
}

// /pbs/home/m/mjoskowicz/jobs/train_ocr/
//├── script.py
//├── requirements.txt
//├── job.sbatch
//├── .venv/
//├── assets/
//│   ├── models/
//│   │   └── <base-model>.mlmodel
//│   └── zips/
//│       └── <uploaded-training-export>.zip
//├── run_<YYMMDD-HHMMSS>-<suffix>/
//│   ├── manifest.env
//│   ├── logs/
//│   ├── trained_models/
//│   └── workspace/
//│       ├── pages_unzipped/
//│       │   └── <zip-stem>/
//│       │       └── ...
//│       ├── alto_files.txt
//│       └── dataset.arrow

func NewOCRModelTraining(models *Model,
	fileSysMgt *filesys.Manager,
	datasets *Dataset,
	annotations *Annotation,
	rootDir string,
	submitter gpufarm.Submitter) *OCRModelTraining {
	return &OCRModelTraining{
		models:      models,
		fileSysMgt:  fileSysMgt,
		datasets:    datasets,
		annotations: annotations,
		submitter:   submitter,
		rootDir:     rootDir,
	}
}

func (j *OCRModelTraining) Submit(mo *model.Model, progress func(string)) (*model.ModelTraining, error) {
	if err := j.verify(mo); err != nil {
		return nil, err
	}

	tmpDir, err := futils.MkdirTemp("ocrflow-model-training-*")
	if err != nil {
		return nil, fmt.Errorf("create training temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	var zipPaths []string
	for _, ref := range mo.BaseAnnotations {
		if ref == nil {
			continue
		}
		zipPath, err := j.stageTrainingAnnotation(tmpDir, ref)
		if err != nil {
			return nil, err
		}
		zipPaths = append(zipPaths, zipPath)
	}
	if len(zipPaths) == 0 {
		return nil, fmt.Errorf("model training requires at least one valid base annotation")
	}

	baseModelPath := ""
	if mo.BaseModelID != "" {
		var err error
		baseModelPath, err = j.localBaseModelPath(mo.BaseModelID)
		if err != nil {
			return nil, err
		}
	}

	result, err := j.submitModelTraining(tmpDir, zipPaths, baseModelPath, progress)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (j *OCRModelTraining) verify(mo *model.Model) error {
	if mo == nil {
		return fmt.Errorf("missing model")
	}
	if j.submitter == nil {
		return fmt.Errorf("model training submitter is not configured")
	}
	if len(mo.BaseAnnotations) == 0 {
		return fmt.Errorf("model training requires at least one base annotation")
	}
	return nil
}

func (j *OCRModelTraining) submitModelTraining(tmpDir string, zipPaths []string, baseModelPath string, progress func(string)) (*model.ModelTraining, error) {
	progress("preparing remote OCR training Python environment")
	remoteEnv, err := j.submitter.PreparePythonEnv(gpufarm.NewPythonEnvRequest(filepath.Join(j.rootDir, "jobs", "train_ocr")))
	if err != nil {
		return nil, err
	}

	remoteAssetsModels := path.Join(remoteEnv.RemoteDir, "assets", "models")
	remoteAssetsZips := path.Join(remoteEnv.RemoteDir, "assets", "zips")

	remoteBaseModelPath := ""
	if baseModelPath != "" {
		remoteBaseModelPath = path.Join(remoteAssetsModels, filepath.Base(baseModelPath))
		progress("syncing base OCR model")
		exists, err := j.submitter.FileExists(remoteBaseModelPath)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := j.submitter.CopyTo(baseModelPath, remoteBaseModelPath); err != nil {
				return nil, err
			}
		}
	}

	var remoteZipPaths []string
	progress(fmt.Sprintf("syncing OCR training annotation archives [%d/%d ZIP files]", len(remoteZipPaths), len(zipPaths)))
	for _, zipPath := range zipPaths {
		remoteZipPath := path.Join(remoteAssetsZips, filepath.Base(zipPath))
		if err := j.submitter.CopyTo(zipPath, remoteZipPath); err != nil {
			return nil, err
		}
		remoteZipPaths = append(remoteZipPaths, remoteZipPath)
		progress(fmt.Sprintf("syncing OCR training annotation archives [%d/%d ZIP files]", len(remoteZipPaths), len(zipPaths)))
	}

	manifestPath := filepath.Join(tmpDir, "manifest.env")

	if err := os.WriteFile(manifestPath, []byte(trainingManifest(remoteEnv, remoteBaseModelPath, remoteZipPaths)), 0o600); err != nil {
		return nil, fmt.Errorf("write training manifest: %w", err)
	}
	if err := j.submitter.CopyTo(manifestPath, path.Join(remoteEnv.RemoteRunDir, "manifest.env")); err != nil {
		return nil, err
	}

	progress("submitting OCR training job")
	submission, err := j.submitter.Submit(remoteEnv)
	if err != nil {
		return nil, err
	}

	statusDetails := map[string]string{
		"submit_output":   submission.SubmitOutput,
		"monitor_command": fmt.Sprintf("ssh %s %s", submission.Host, envexec.ShellQuote("tail -f "+path.Join(remoteEnv.LogsDir, "*.*"))),
	}
	if submission.SchedulerJobID != "" {
		statusDetails["scheduler_job_id"] = submission.SchedulerJobID
		if submission.Backend == "slurm" {
			statusDetails["slurm_job_id"] = submission.SchedulerJobID
		}
	}
	return &model.ModelTraining{
		Status:        model.ModelTrainingStatusSubmitted,
		StatusDetails: statusDetails,
		Backend:       submission.Backend,
		GPUFarmHost:   submission.Host,
		RemoteRunDir:  remoteEnv.RemoteRunDir,
	}, nil
}

func (j *OCRModelTraining) stageTrainingAnnotation(tmpDir string, ref *annotation.Reference) (string, error) {
	if ref.DatasetID == "" || ref.ID == "" {
		return "", fmt.Errorf("invalid base annotation reference")
	}
	ann, err := j.annotations.Get(ref.DatasetID, ref.ID)
	if err != nil {
		return "", fmt.Errorf("get base annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
	}
	if !ann.Ocred {
		return "", fmt.Errorf("base annotation %s:%s is not OCRed", ref.DatasetID, ref.ID)
	}
	ds, err := j.datasets.Get(ref.DatasetID)
	if err != nil {
		return "", fmt.Errorf("get dataset %s: %w", ref.DatasetID, err)
	}

	altoDir := j.fileSysMgt.DatasetAnnotationAltoDir(ann)
	imgDir := j.fileSysMgt.DatasetImagesDir(ds)
	stageDir := filepath.Join(tmpDir, fmt.Sprintf("%s_%s", ref.DatasetID, ref.ID))
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return "", fmt.Errorf("create annotation training stage dir: %w", err)
	}

	entries, err := os.ReadDir(altoDir)
	if err != nil {
		return "", fmt.Errorf("read ALTO dir for annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
	}

	copied := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".xml" || strings.EqualFold(entry.Name(), "mets.xml") {
			continue
		}
		altoPath := filepath.Join(altoDir, entry.Name())
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		imagePath, err := findTrainingImage(imgDir, stem)
		if err != nil {
			return "", fmt.Errorf("find image for %s: %w", altoPath, err)
		}
		if err := futils.CopyFile(altoPath, filepath.Join(stageDir, entry.Name())); err != nil {
			return "", fmt.Errorf("copy ALTO %s: %w", altoPath, err)
		}
		if err := futils.CopyFile(imagePath, filepath.Join(stageDir, filepath.Base(imagePath))); err != nil {
			return "", fmt.Errorf("copy image %s: %w", imagePath, err)
		}
		copied++
	}
	if copied == 0 {
		return "", fmt.Errorf("no ALTO XML files found for annotation %s:%s", ref.DatasetID, ref.ID)
	}

	zipPath := filepath.Join(tmpDir, fmt.Sprintf("%s_%s.zip", ref.DatasetID, ref.ID))
	if err := futils.Zip(stageDir, zipPath); err != nil {
		return "", fmt.Errorf("zip staged annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
	}
	return zipPath, nil
}

func findTrainingImage(imgDir string, stem string) (string, error) {
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".tif", ".tiff"} {
		p := filepath.Join(imgDir, stem+ext)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("no matching image found in %s", imgDir)
}

func (j *OCRModelTraining) localBaseModelPath(modelID string) (string, error) {
	baseModel, err := j.models.Get(modelID)
	if err != nil {
		return "", fmt.Errorf("get base model %s: %w", modelID, err)
	}
	if baseModel.Location != model.OCRModelLocationLocal {
		return "", fmt.Errorf("base model %s is not local", modelID)
	}
	if baseModel.Type != common.OCRModelTypeOCR {
		return "", fmt.Errorf("base model %s is not an OCR text model", modelID)
	}
	if baseModel.LocalPath == "" {
		return "", fmt.Errorf("base model %s has no local path", modelID)
	}
	p := j.fileSysMgt.ModelPath(baseModel)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("stat base model %s: %w", p, err)
	}
	return p, nil
}

func trainingManifest(remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteZipPaths []string) string {
	remoteRoot := remoteEnv.RemoteDir
	remoteRunDir := remoteEnv.RemoteRunDir
	logsDir := remoteEnv.LogsDir

	var b strings.Builder
	fmt.Fprintf(&b, "export PROJECT_ROOT=%s\n", envexec.ShellQuote(remoteRoot))
	fmt.Fprintf(&b, "export RUN_ID=%s\n", envexec.ShellQuote(remoteEnv.RunID))
	fmt.Fprintf(&b, "export RUN_DIR=%s\n", envexec.ShellQuote(remoteRunDir))
	fmt.Fprintf(&b, "export BASE_MODEL_PATH=%s\n", envexec.ShellQuote(remoteBaseModelPath))
	fmt.Fprintf(&b, "export MODEL_PREFIX=%s\n", envexec.ShellQuote("kraken_model"))
	fmt.Fprintf(&b, "export WORK_DIR=%s\n", envexec.ShellQuote(path.Join(remoteRunDir, "workspace")))
	fmt.Fprintf(&b, "export OUTPUT_DIR=%s\n", envexec.ShellQuote(path.Join(remoteRunDir, "trained_models")))
	fmt.Fprintf(&b, "export LOGS_DIR=%s\n", envexec.ShellQuote(logsDir))
	b.WriteString("export ZIP_PATHS=(\n")
	for _, zipPath := range remoteZipPaths {
		fmt.Fprintf(&b, "  %s\n", envexec.ShellQuote(zipPath))
	}
	b.WriteString(")\n")
	return b.String()
}
