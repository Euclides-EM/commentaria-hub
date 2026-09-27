package service

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
)

const remoteDetectionJobName = "detect_annotation"

type AnnotationDetectionRemote struct {
	fileSysMgt *filesys.Manager
	rootDir    string
	apiURL     string
	apiToken   string
	submitter  gpufarm.Submitter

	manualBundles *GPUFarmManualBundles
}

type remoteDetectionRequest struct {
	Mode              annotation.DetectionMode
	ImageDir          string
	Annotation        *annotation.Annotation
	Pages             []int
	IncludeCategories []string
	IgnoreCategories  []string
	Model             *model.Model
	ModelPath         string
	ManualRun         bool
}

func NewAnnotationDetectionRemote(fileSysMgt *filesys.Manager, rootDir string, apiURL string, apiToken string, submitter gpufarm.Submitter, manualBundles *GPUFarmManualBundles) *AnnotationDetectionRemote {
	return &AnnotationDetectionRemote{
		fileSysMgt: fileSysMgt,
		rootDir:    rootDir,
		apiURL:     strings.TrimRight(apiURL, "/"),
		apiToken:   apiToken,
		submitter:  submitter,

		manualBundles: manualBundles,
	}
}

func (r *AnnotationDetectionRemote) Submit(req remoteDetectionRequest, onSubmitted func(string)) error {
	if r == nil || r.submitter == nil {
		return fmt.Errorf("GPU farm detection submitter is not configured")
	}
	if req.Annotation == nil {
		return fmt.Errorf("missing annotation for GPU farm detection")
	}
	if len(req.Pages) == 0 {
		return nil
	}
	if r.apiURL == "" {
		return fmt.Errorf("API_URL is required for GPU farm detection result upload")
	}
	if req.Model != nil && req.Model.Location != model.OCRModelLocationLocal {
		return fmt.Errorf("GPU farm detection currently supports local models only, got %s", req.Model.Location)
	}

	if req.ManualRun {
		return r.submitManual(req, onSubmitted)
	}

	remoteEnv, err := r.submitter.PreparePythonEnv(gpufarm.NewPythonEnvRequest(filepath.Join(r.rootDir, "jobs", remoteDetectionJobName)))
	if err != nil {
		return fmt.Errorf("prepare remote detection Python environment: %w", err)
	}
	submitted := false
	defer func() {
		if !submitted {
			if err := r.submitter.Discard(remoteEnv); err != nil {
				log.Printf("GPU farm abandoned detection upload cleanup failed: annotation=%s run_dir=%s error=%v", req.Annotation.ID, remoteEnv.RemoteRunDir, err)
			}
		}
	}()

	log.Printf("GPU farm detection upload started: annotation=%s mode=%s pages=%d run_id=%s run_dir=%s", req.Annotation.ID, req.Mode, len(req.Pages), remoteEnv.RunID, remoteEnv.RemoteRunDir)
	if err := r.stageInputs(req, remoteEnv.RunID, func(localPath string, relPath string) error {
		return r.submitter.CopyTo(localPath, path.Join(remoteEnv.RemoteRunDir, relPath))
	}); err != nil {
		return err
	}

	submission, err := r.submitter.Submit(remoteEnv)
	if err != nil {
		return fmt.Errorf("submit GPU farm detection job: %w", err)
	}
	submitted = true
	log.Printf("GPU farm detection submitted: annotation=%s mode=%s backend=%s scheduler_job_id=%s run_id=%s", req.Annotation.ID, req.Mode, submission.Backend, submission.SchedulerJobID, remoteEnv.RunID)
	if submission.Host != "" && submission.SchedulerJobID != "" {
		stdoutPath := path.Join(remoteEnv.LogsDir, "annotation_detect_"+submission.SchedulerJobID+".out")
		stderrPath := path.Join(remoteEnv.LogsDir, "annotation_detect_"+submission.SchedulerJobID+".err")
		followCommand := detectionFollowCommand(submission.Host, stdoutPath, stderrPath)
		log.Printf("Follow GPU farm detection logs with: %s", followCommand)
		if onSubmitted != nil {
			onSubmitted(followCommand)
		}
	}
	return nil
}

func (r *AnnotationDetectionRemote) submitManual(req remoteDetectionRequest, onSubmitted func(string)) error {
	runID := gpufarm.NewRunID()
	stageDir, err := futils.MkdirTemp("annotation-detect-manual-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(stageDir)

	if err := r.stageInputs(req, runID, func(localPath string, relPath string) error {
		return futils.CopyFile(localPath, filepath.Join(stageDir, runID, filepath.FromSlash(relPath)))
	}); err != nil {
		return err
	}
	message, err := r.manualBundles.Publish(filepath.Join(r.rootDir, "jobs", remoteDetectionJobName), runID, stageDir)
	if err != nil {
		return err
	}
	log.Printf("GPU farm detection awaiting manual run: annotation=%s mode=%s run_id=%s: %s", req.Annotation.ID, req.Mode, runID, message)
	if onSubmitted != nil {
		onSubmitted(message)
	}
	return nil
}

func (r *AnnotationDetectionRemote) stageInputs(req remoteDetectionRequest, runID string, copyTo func(localPath string, relPath string) error) error {
	for i, p := range req.Pages {
		imageName := pagesparser.PageToPNGFilename(p)
		if err := copyTo(filepath.Join(req.ImageDir, imageName), path.Join("assets", "images", imageName)); err != nil {
			return fmt.Errorf("copy image %s to GPU farm: %w", imageName, err)
		}
		if (i+1)%25 == 0 || i+1 == len(req.Pages) {
			log.Printf("GPU farm detection image upload progress: annotation=%s run_id=%s uploaded=%d total=%d", req.Annotation.ID, runID, i+1, len(req.Pages))
		}
	}

	altoDir := r.fileSysMgt.DatasetAnnotationAltoDir(req.Annotation)
	if req.Mode != annotation.DetectionModeModelSegment {
		for _, p := range req.Pages {
			altoName := pagesparser.PageToXMLFilename(p)
			if err := copyTo(filepath.Join(altoDir, altoName), path.Join("assets", "alto", altoName)); err != nil {
				return fmt.Errorf("copy ALTO %s to GPU farm: %w", altoName, err)
			}
		}
	}

	modelRelPath := ""
	if req.ModelPath != "" {
		modelRelPath = path.Join("assets", "models", filepath.Base(req.ModelPath))
		if err := copyTo(req.ModelPath, modelRelPath); err != nil {
			return fmt.Errorf("copy model to GPU farm: %w", err)
		}
	}

	tmp, err := futils.MkdirTemp("annotation-detect-remote-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	manifestPath := filepath.Join(tmp, "manifest.env")
	if err := os.WriteFile(manifestPath, []byte(r.detectionManifest(req, runID, modelRelPath)), 0o600); err != nil {
		return fmt.Errorf("write detection manifest: %w", err)
	}
	if err := copyTo(manifestPath, "manifest.env"); err != nil {
		return fmt.Errorf("copy detection manifest to GPU farm: %w", err)
	}
	return nil
}

func detectionFollowCommand(host, stdoutPath, stderrPath string) string {
	// Keep each SSH argument separate. Quoting the complete remote command as
	// well as its paths creates nested shell quotes that are escaped in JSON and
	// easy to copy literally as part of the filenames.
	return fmt.Sprintf("ssh %s tail -n 100 -F %s %s",
		envexec.ShellQuote(host),
		envexec.ShellQuote(stdoutPath),
		envexec.ShellQuote(stderrPath),
	)
}

func (r *AnnotationDetectionRemote) detectionManifest(req remoteDetectionRequest, runID string, modelRelPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export RUN_DIR=\"$(pwd)\"\n")
	fmt.Fprintf(&b, "export PROJECT_ROOT=\"$(dirname \"$RUN_DIR\")\"\n")
	fmt.Fprintf(&b, "export RUN_ID=%s\n", envexec.ShellQuote(runID))
	fmt.Fprintf(&b, "export LOGS_DIR=\"$RUN_DIR/logs\"\n")
	fmt.Fprintf(&b, "export MODE=%s\n", envexec.ShellQuote(string(req.Mode)))
	fmt.Fprintf(&b, "export IMAGE_DIR=\"$RUN_DIR/assets/images\"\n")
	fmt.Fprintf(&b, "export ALTO_DIR=\"$RUN_DIR/assets/alto\"\n")
	fmt.Fprintf(&b, "export OUTPUT_DIR=\"$RUN_DIR/output/alto\"\n")
	fmt.Fprintf(&b, "export ARTIFACTS_DIR=\"$RUN_DIR/artifacts\"\n")
	fmt.Fprintf(&b, "export MODEL_PATH=%s\n", manifestPathUnder("RUN_DIR", modelRelPath))
	fmt.Fprintf(&b, "export RESULT_UPLOAD_URL=%s\n", envexec.ShellQuote(r.resultUploadURL(req.Annotation)))
	fmt.Fprintf(&b, "export RESULT_FAILURE_URL=%s\n", envexec.ShellQuote(r.resultFailureURL(req.Annotation)))
	fmt.Fprintf(&b, "export RESULT_UPLOAD_TOKEN=%s\n", envexec.ShellQuote(r.apiToken))
	fmt.Fprintf(&b, "export INCLUDE_CATEGORIES=%s\n", envexec.ShellQuote(strings.Join(req.IncludeCategories, "\n")))
	fmt.Fprintf(&b, "export IGNORE_CATEGORIES=%s\n", envexec.ShellQuote(strings.Join(req.IgnoreCategories, "\n")))
	if req.Model != nil {
		fmt.Fprintf(&b, "export MODEL_TYPE=%s\n", envexec.ShellQuote(string(req.Model.Type)))
	}
	return b.String()
}

func (r *AnnotationDetectionRemote) resultUploadURL(ann *annotation.Annotation) string {
	return fmt.Sprintf("%s/datasets/%s/annotations/%s/detection_upload", r.apiURL, ann.DatasetID, ann.ID)
}

func (r *AnnotationDetectionRemote) resultFailureURL(ann *annotation.Annotation) string {
	return fmt.Sprintf("%s/datasets/%s/annotations/%s/detection_failure", r.apiURL, ann.DatasetID, ann.ID)
}
