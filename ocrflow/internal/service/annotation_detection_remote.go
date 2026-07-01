package service

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/gpufarm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

const remoteDetectionJobName = "detect_annotation"

type AnnotationDetectionRemote struct {
	fileSysMgt *filesys.Manager
	rootDir    string
	apiURL     string
	apiToken   string
	submitter  gpufarm.Submitter
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
}

func NewAnnotationDetectionRemote(fileSysMgt *filesys.Manager, rootDir string, apiURL string, apiToken string, submitter gpufarm.Submitter) *AnnotationDetectionRemote {
	return &AnnotationDetectionRemote{
		fileSysMgt: fileSysMgt,
		rootDir:    rootDir,
		apiURL:     strings.TrimRight(apiURL, "/"),
		apiToken:   apiToken,
		submitter:  submitter,
	}
}

func (r *AnnotationDetectionRemote) Submit(req remoteDetectionRequest) error {
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

	remoteEnv, err := r.submitter.PreparePythonEnv(gpufarm.NewPythonEnvRequest(filepath.Join(r.rootDir, "jobs", remoteDetectionJobName)))
	if err != nil {
		return fmt.Errorf("prepare remote detection Python environment: %w", err)
	}

	for _, p := range req.Pages {
		imageName := pagesparser.PageToPNGFilename(p)
		if err := r.submitter.CopyTo(filepath.Join(req.ImageDir, imageName), path.Join(remoteEnv.RemoteRunDir, "assets", "images", imageName)); err != nil {
			return fmt.Errorf("copy image %s to GPU farm: %w", imageName, err)
		}
	}

	altoDir := r.fileSysMgt.DatasetAnnotationAltoDir(req.Annotation)
	if req.Mode != annotation.DetectionModeModelSegment {
		for _, p := range req.Pages {
			altoName := pagesparser.PageToXMLFilename(p)
			if err := r.submitter.CopyTo(filepath.Join(altoDir, altoName), path.Join(remoteEnv.RemoteRunDir, "assets", "alto", altoName)); err != nil {
				return fmt.Errorf("copy ALTO %s to GPU farm: %w", altoName, err)
			}
		}
	}

	remoteModelPath := ""
	if req.ModelPath != "" {
		remoteModelPath = path.Join(remoteEnv.RemoteRunDir, "assets", "models", filepath.Base(req.ModelPath))
		if err := r.submitter.CopyTo(req.ModelPath, remoteModelPath); err != nil {
			return fmt.Errorf("copy model to GPU farm: %w", err)
		}
	}

	tmp, err := futils.MkdirTemp("annotation-detect-remote-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	manifestPath := filepath.Join(tmp, "manifest.env")
	if err := os.WriteFile(manifestPath, []byte(r.detectionManifest(req, remoteEnv, remoteModelPath)), 0o600); err != nil {
		return fmt.Errorf("write detection manifest: %w", err)
	}
	if err := r.submitter.CopyTo(manifestPath, path.Join(remoteEnv.RemoteRunDir, "manifest.env")); err != nil {
		return fmt.Errorf("copy detection manifest to GPU farm: %w", err)
	}

	if _, err := r.submitter.Submit(remoteEnv); err != nil {
		return fmt.Errorf("submit GPU farm detection job: %w", err)
	}
	return nil
}

func (r *AnnotationDetectionRemote) detectionManifest(req remoteDetectionRequest, remoteEnv *gpufarm.RemoteEnv, remoteModelPath string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "export PROJECT_ROOT=%s\n", envexec.ShellQuote(remoteEnv.RemoteDir))
	fmt.Fprintf(&b, "export RUN_ID=%s\n", envexec.ShellQuote(remoteEnv.RunID))
	fmt.Fprintf(&b, "export RUN_DIR=%s\n", envexec.ShellQuote(remoteEnv.RemoteRunDir))
	fmt.Fprintf(&b, "export LOGS_DIR=%s\n", envexec.ShellQuote(remoteEnv.LogsDir))
	fmt.Fprintf(&b, "export MODE=%s\n", envexec.ShellQuote(string(req.Mode)))
	fmt.Fprintf(&b, "export IMAGE_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "assets", "images")))
	fmt.Fprintf(&b, "export ALTO_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "assets", "alto")))
	fmt.Fprintf(&b, "export OUTPUT_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "output", "alto")))
	fmt.Fprintf(&b, "export ARTIFACTS_DIR=%s\n", envexec.ShellQuote(path.Join(remoteEnv.RemoteRunDir, "artifacts")))
	fmt.Fprintf(&b, "export MODEL_PATH=%s\n", envexec.ShellQuote(remoteModelPath))
	fmt.Fprintf(&b, "export RESULT_UPLOAD_URL=%s\n", envexec.ShellQuote(r.resultUploadURL(req.Annotation)))
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
