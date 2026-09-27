package service

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
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
	bundles     *GPUFarmManualBundles
}

type trainingRemoteAsset struct {
	LocalPath string
	AssetDir  string
}

type trainingRemoteRequest struct {
	Training      *model.ModelTraining
	AsyncJobID    string
	TmpDir        string
	JobName       string
	BaseModelPath string
	Assets        []trainingRemoteAsset
	StatusDetails map[string]string
	Manifest      func(runID string, baseModelRelPath string, assetRelPaths []string) string
	AssetProgress func(done int, total int) string
}

func NewModelTrainingRemote(models *Model,
	fileSysMgt *filesys.Manager,
	datasets *Dataset,
	annotations *Annotation,
	rootDir string,
	apiURL string,
	apiToken string,
	submitter gpufarm.Submitter,
	bundles *GPUFarmManualBundles) *ModelTrainingRemote {
	return &ModelTrainingRemote{
		models:      models,
		fileSysMgt:  fileSysMgt,
		datasets:    datasets,
		annotations: annotations,
		rootDir:     rootDir,
		apiURL:      strings.TrimRight(apiURL, "/"),
		apiToken:    apiToken,
		submitter:   submitter,
		bundles:     bundles,
	}
}

func (r *ModelTrainingRemote) Submit(training *model.ModelTraining, jobID string, progress func(string)) (*model.ModelTraining, error) {
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
		return r.submitOCR(training, jobID, progress)
	case common.OCRModelTypeSegment:
		return r.submitYOLO(training, jobID, progress)
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
	if req.Training.ManualRun {
		return r.submitManual(req, progress)
	}

	progress("preparing remote training Python environment")
	remoteEnv, err := r.submitter.PreparePythonEnv(gpufarm.NewPythonEnvRequest(filepath.Join(r.rootDir, "jobs", req.JobName)))
	if err != nil {
		return nil, err
	}
	submitted := false
	defer func() {
		if !submitted {
			if err := r.submitter.Discard(remoteEnv); err != nil {
				log.Printf("GPU farm abandoned training upload cleanup failed: training=%s run_dir=%s error=%v", req.Training.ID, remoteEnv.RemoteRunDir, err)
			}
		}
	}()

	copyTo := func(localPath string, relPath string) error {
		return r.submitter.CopyTo(localPath, path.Join(remoteEnv.RemoteDir, relPath))
	}
	baseModelRelPath, assetRelPaths, err := r.stageTrainingAssets(req, progress, copyTo, func(relPath string) (bool, error) {
		return r.submitter.FileExists(path.Join(remoteEnv.RemoteDir, relPath))
	})
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(req.TmpDir, "manifest.env")
	if err := os.WriteFile(manifestPath, []byte(req.Manifest(remoteEnv.RunID, baseModelRelPath, assetRelPaths)), 0o600); err != nil {
		return nil, fmt.Errorf("write training manifest: %w", err)
	}
	if err := copyTo(manifestPath, path.Join(remoteEnv.RunID, "manifest.env")); err != nil {
		return nil, err
	}

	progress("submitting training job")
	submission, err := r.submitter.Submit(remoteEnv)
	if err != nil {
		return nil, err
	}
	submitted = true

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

func (r *ModelTrainingRemote) submitManual(req trainingRemoteRequest, progress func(string)) (*model.ModelTraining, error) {
	runID := gpufarm.NewRunID()
	stageDir := filepath.Join(req.TmpDir, "manual_bundle")
	copyTo := func(localPath string, relPath string) error {
		return futils.CopyFile(localPath, filepath.Join(stageDir, filepath.FromSlash(relPath)))
	}
	baseModelRelPath, assetRelPaths, err := r.stageTrainingAssets(req, progress, copyTo, func(string) (bool, error) {
		return false, nil
	})
	if err != nil {
		return nil, err
	}

	manifestPath := filepath.Join(stageDir, runID, "manifest.env")
	if err := os.MkdirAll(filepath.Dir(manifestPath), 0o755); err != nil {
		return nil, fmt.Errorf("create manual training run directory: %w", err)
	}
	if err := os.WriteFile(manifestPath, []byte(req.Manifest(runID, baseModelRelPath, assetRelPaths)), 0o600); err != nil {
		return nil, fmt.Errorf("write training manifest: %w", err)
	}

	progress("publishing manual GPU farm bundle")
	message, err := r.bundles.Publish(filepath.Join(r.rootDir, "jobs", req.JobName), req.AsyncJobID, runID, stageDir)
	if err != nil {
		return nil, err
	}
	log.Printf("GPU farm training awaiting manual run: training=%s run_id=%s: %s", req.Training.ID, runID, message)

	statusDetails := map[string]string{
		"manual_run": message,
		"run_id":     runID,
	}
	if r.modelUploadURL() != "" {
		statusDetails["model_upload_url"] = r.modelUploadURL()
	}
	for k, v := range req.StatusDetails {
		statusDetails[k] = v
	}
	return &model.ModelTraining{
		Status:        model.ModelTrainingStatusSubmitted,
		StatusDetails: statusDetails,
		Backend:       gpuFarmManualBackend,
		Model:         req.Training.Model,
		Epochs:        req.Training.Epochs,
	}, nil
}

func (r *ModelTrainingRemote) stageTrainingAssets(req trainingRemoteRequest, progress func(string), copyTo func(localPath string, relPath string) error, exists func(relPath string) (bool, error)) (string, []string, error) {
	baseModelRelPath := ""
	if req.BaseModelPath != "" {
		baseModelRelPath = path.Join("assets", "models", filepath.Base(req.BaseModelPath))
		progress("syncing base model to remote")
		found, err := exists(baseModelRelPath)
		if err != nil {
			return "", nil, err
		}
		if !found {
			if err := copyTo(req.BaseModelPath, baseModelRelPath); err != nil {
				return "", nil, err
			}
		}
	}

	assetRelPaths := make([]string, 0, len(req.Assets))
	for i, asset := range req.Assets {
		if req.AssetProgress != nil {
			progress(req.AssetProgress(i, len(req.Assets)))
		}
		assetRelPath := path.Join("assets", asset.AssetDir, filepath.Base(asset.LocalPath))
		if err := copyTo(asset.LocalPath, assetRelPath); err != nil {
			return "", nil, err
		}
		assetRelPaths = append(assetRelPaths, assetRelPath)
		if req.AssetProgress != nil {
			progress(req.AssetProgress(len(assetRelPaths), len(req.Assets)))
		}
	}
	return baseModelRelPath, assetRelPaths, nil
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

func (r *ModelTrainingRemote) writeCommonManifest(b *strings.Builder, training *model.ModelTraining, runID string, baseModelRelPath string) {
	mo := training.Model
	fmt.Fprintf(b, "export RUN_DIR=\"$(pwd)\"\n")
	fmt.Fprintf(b, "export PROJECT_ROOT=\"$(dirname \"$RUN_DIR\")\"\n")
	fmt.Fprintf(b, "export RUN_ID=%s\n", envexec.ShellQuote(runID))
	fmt.Fprintf(b, "export BASE_MODEL_PATH=%s\n", manifestPathUnder("PROJECT_ROOT", baseModelRelPath))
	fmt.Fprintf(b, "export WORK_DIR=\"$RUN_DIR/workspace\"\n")
	fmt.Fprintf(b, "export OUTPUT_DIR=\"$RUN_DIR/trained_models\"\n")
	fmt.Fprintf(b, "export LOGS_DIR=\"$RUN_DIR/logs\"\n")
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
