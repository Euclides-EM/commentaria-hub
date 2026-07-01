package service

import (
	"errors"
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
	"github.com/MiaMish/elements-dh/ocrflow/pkg/gpufarm"
)

type ModelTrainingRemote struct {
	models      *Model
	fileSysMgt  *filesys.Manager
	datasets    *Dataset
	annotations *Annotation
	submitter   gpufarm.Submitter
	apiURL      string
	apiToken    string
	rootDir     string
}

type trainingRemoteAsset struct {
	LocalPath string
	AssetDir  string
}

type trainingRemoteRequest struct {
	Training      *model.ModelTraining
	TmpDir        string
	JobName       string
	BaseModelPath string
	Assets        []trainingRemoteAsset
	StatusDetails map[string]string
	Manifest      func(remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteAssetPaths []string) string
	AssetProgress func(done int, total int) string
}

func NewModelTrainingRemote(models *Model,
	fileSysMgt *filesys.Manager,
	datasets *Dataset,
	annotations *Annotation,
	rootDir string,
	apiURL string,
	apiToken string,
	submitter gpufarm.Submitter) *ModelTrainingRemote {
	return &ModelTrainingRemote{
		models:      models,
		fileSysMgt:  fileSysMgt,
		datasets:    datasets,
		annotations: annotations,
		rootDir:     rootDir,
		apiURL:      strings.TrimRight(apiURL, "/"),
		apiToken:    apiToken,
		submitter:   submitter,
	}
}

func (r *ModelTrainingRemote) Submit(training *model.ModelTraining, progress func(string)) (*model.ModelTraining, error) {
	if training == nil {
		return nil, fmt.Errorf("missing model training request")
	}
	if training.Model == nil {
		return nil, fmt.Errorf("model training request is missing model")
	}
	if r.submitter == nil {
		return nil, fmt.Errorf("model training submitter is not configured")
	}
	if len(training.Model.BaseAnnotations) == 0 {
		return nil, errors.New("model training requires at least one base annotation")
	}

	switch training.Model.Type {
	case common.OCRModelTypeOCR:
		return r.submitOCR(training, progress)
	case common.OCRModelTypeSegment:
		return r.submitYOLO(training, progress)
	default:
		return nil, fmt.Errorf("unsupported model training type: %s", training.Model.Type)
	}
}

func (r *ModelTrainingRemote) localBaseModelPath(modelID string, modelType common.OCRModelType, modelDescription string) (string, error) {
	baseModel, err := r.models.Get(modelID)
	if err != nil {
		return "", fmt.Errorf("get base model %s: %w", modelID, err)
	}
	if baseModel.Location != model.OCRModelLocationLocal {
		return "", fmt.Errorf("base model %s is not local", modelID)
	}
	if baseModel.Type != modelType {
		return "", fmt.Errorf("base model %s is not %s", modelID, modelDescription)
	}
	if baseModel.LocalPath == "" {
		return "", fmt.Errorf("base model %s has no local path", modelID)
	}
	p := r.fileSysMgt.ModelPath(baseModel)
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("stat base model %s: %w", p, err)
	}
	return p, nil
}

func (r *ModelTrainingRemote) submit(req trainingRemoteRequest, progress func(string)) (*model.ModelTraining, error) {
	if req.Training == nil || req.Training.Model == nil {
		return nil, fmt.Errorf("missing model training request")
	}
	if req.Manifest == nil {
		return nil, errors.New("missing training manifest builder")
	}

	progress("preparing remote training Python environment")
	remoteEnv, err := r.submitter.PreparePythonEnv(gpufarm.NewPythonEnvRequest(filepath.Join(r.rootDir, "jobs", req.JobName)))
	if err != nil {
		return nil, err
	}

	remoteBaseModelPath := ""
	if req.BaseModelPath != "" {
		remoteBaseModelPath = path.Join(remoteEnv.RemoteDir, "assets", "models", filepath.Base(req.BaseModelPath))
		progress("syncing base model to remote")
		exists, err := r.submitter.FileExists(remoteBaseModelPath)
		if err != nil {
			return nil, err
		}
		if !exists {
			if err := r.submitter.CopyTo(req.BaseModelPath, remoteBaseModelPath); err != nil {
				return nil, err
			}
		}
	}

	remoteAssetPaths := make([]string, 0, len(req.Assets))
	for i, asset := range req.Assets {
		if req.AssetProgress != nil {
			progress(req.AssetProgress(i, len(req.Assets)))
		}
		remoteAssetPath := path.Join(remoteEnv.RemoteDir, "assets", asset.AssetDir, filepath.Base(asset.LocalPath))
		if err := r.submitter.CopyTo(asset.LocalPath, remoteAssetPath); err != nil {
			return nil, err
		}
		remoteAssetPaths = append(remoteAssetPaths, remoteAssetPath)
		if req.AssetProgress != nil {
			progress(req.AssetProgress(len(remoteAssetPaths), len(req.Assets)))
		}
	}

	manifestPath := filepath.Join(req.TmpDir, "manifest.env")
	if err := os.WriteFile(manifestPath, []byte(req.Manifest(remoteEnv, remoteBaseModelPath, remoteAssetPaths)), 0o600); err != nil {
		return nil, fmt.Errorf("write training manifest: %w", err)
	}
	if err := r.submitter.CopyTo(manifestPath, path.Join(remoteEnv.RemoteRunDir, "manifest.env")); err != nil {
		return nil, err
	}

	progress("submitting training job")
	submission, err := r.submitter.Submit(remoteEnv)
	if err != nil {
		return nil, err
	}

	statusDetails := r.trainingStatusDetails(submission, remoteEnv, req.StatusDetails)
	return &model.ModelTraining{
		Status:        model.ModelTrainingStatusSubmitted,
		StatusDetails: statusDetails,
		Backend:       submission.Backend,
		GPUFarmHost:   submission.Host,
		RemoteRunDir:  remoteEnv.RemoteRunDir,
		Model:         req.Training.Model,
		Epochs:        req.Training.Epochs,
	}, nil
}

func (r *ModelTrainingRemote) trainingStatusDetails(submission *gpufarm.JobSubmission, remoteEnv *gpufarm.RemoteEnv, extra map[string]string) map[string]string {
	statusDetails := map[string]string{
		"submit_output":   submission.SubmitOutput,
		"monitor_command": fmt.Sprintf("ssh %s %s", submission.Host, envexec.ShellQuote("tail -f "+path.Join(remoteEnv.LogsDir, "*.*"))),
	}
	if r.modelUploadURL() != "" {
		statusDetails["model_upload_url"] = r.modelUploadURL()
	}
	if submission.SchedulerJobID != "" {
		statusDetails["scheduler_job_id"] = submission.SchedulerJobID
		if submission.Backend == "slurm" {
			statusDetails["slurm_job_id"] = submission.SchedulerJobID
		}
	}
	for k, v := range extra {
		statusDetails[k] = v
	}
	return statusDetails
}

func (r *ModelTrainingRemote) writeCommonManifest(b *strings.Builder, training *model.ModelTraining, remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string) {
	mo := training.Model
	fmt.Fprintf(b, "export PROJECT_ROOT=%s\n", envexec.ShellQuote(remoteEnv.RemoteDir))
	fmt.Fprintf(b, "export RUN_ID=%s\n", envexec.ShellQuote(remoteEnv.RunID))
	fmt.Fprintf(b, "export RUN_DIR=%s\n", envexec.ShellQuote(remoteEnv.RemoteRunDir))
	fmt.Fprintf(b, "export BASE_MODEL_PATH=%s\n", envexec.ShellQuote(remoteBaseModelPath))
	fmt.Fprintf(b, "export WORK_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "workspace")))
	fmt.Fprintf(b, "export OUTPUT_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "trained_models")))
	fmt.Fprintf(b, "export LOGS_DIR=%s\n", envexec.ShellQuote(remoteEnv.LogsDir))
	fmt.Fprintf(b, "export MODEL_UPLOAD_URL=%s\n", envexec.ShellQuote(r.modelUploadURL()))
	fmt.Fprintf(b, "export MODEL_UPLOAD_TOKEN=%s\n", envexec.ShellQuote(r.apiToken))
	fmt.Fprintf(b, "export MODEL_NAME=%s\n", envexec.ShellQuote(mo.Name))
	fmt.Fprintf(b, "export MODEL_DESCRIPTION=%s\n", envexec.ShellQuote(mo.Description))
	fmt.Fprintf(b, "export MODEL_BASE_MODEL_ID=%s\n", envexec.ShellQuote(mo.BaseModelID))
	fmt.Fprintf(b, "export MODEL_BASE_ANNOTATIONS=%s\n", envexec.ShellQuote(baseAnnotationRefs(mo.BaseAnnotations)))
	if training.Epochs > 0 {
		fmt.Fprintf(b, "export TRAIN_EPOCHS=%s\n", envexec.ShellQuote(fmt.Sprintf("%d", training.Epochs)))
	} else {
		fmt.Fprintf(b, "export TRAIN_EPOCHS=%s\n", envexec.ShellQuote(""))
	}
}

func (r *ModelTrainingRemote) modelUploadURL() string {
	if r.apiURL == "" {
		return ""
	}
	return r.apiURL + "/models_upload"
}

func baseAnnotationRefs(refs []*annotation.Reference) string {
	baseAnnotations := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref == nil {
			continue
		}
		baseAnnotations = append(baseAnnotations, ref.DatasetID+":"+ref.ID)
	}
	return strings.Join(baseAnnotations, ",")
}
