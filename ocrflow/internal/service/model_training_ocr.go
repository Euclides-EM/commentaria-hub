package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/gpufarm"
)

// Remote layout under ${GPU_FARM_JOB_ROOT}/train_ocr:
//
//	train_ocr/
//	├── script.py
//	├── requirements.txt
//	├── job.sbatch
//	├── .venv/
//	├── assets/                         # shared across runs
//	│   ├── models/
//	│   │   └── <base-model>.mlmodel
//	│   └── zips/
//	│       └── <dataset-id>_<annotation-id>.zip
//	└── run_<YYMMDD-HHMMSS>-<suffix>/   # one directory per submission
//	    ├── manifest.env
//	    ├── logs/
//	    │   ├── kraken_train_<slurm-job-id>.out
//	    │   └── kraken_train_<slurm-job-id>.err
//	    ├── trained_models/
//	    │   ├── kraken_model_<epoch>.mlmodel
//	    │   └── kraken_model_best.mlmodel
//	    └── workspace/
//	        ├── pages_unzipped/
//	        │   └── <zip-stem>/
//	        │       ├── <page>.xml
//	        │       └── <page>.<image-ext>
//	        ├── alto_files.txt
//	        └── dataset.arrow
func (r *ModelTrainingRemote) submitOCR(training *model.ModelTraining, progress func(string)) (*model.ModelTraining, error) {
	mo := training.Model

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
		zipPath, err := r.stageOCRTrainingAnnotation(tmpDir, ref)
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
		baseModelPath, err = r.localBaseModelPath(mo.BaseModelID, common.OCRModelTypeOCR, "an OCR text model")
		if err != nil {
			return nil, err
		}
	}

	assets := make([]trainingRemoteAsset, 0, len(zipPaths))
	for _, zipPath := range zipPaths {
		assets = append(assets, trainingRemoteAsset{LocalPath: zipPath, AssetDir: "zips"})
	}
	return r.submit(trainingRemoteRequest{
		Training:      training,
		TmpDir:        tmpDir,
		JobName:       "train_ocr",
		DisplayName:   "OCR",
		BaseModelPath: baseModelPath,
		Assets:        assets,
		Manifest: func(remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteAssetPaths []string) string {
			return r.ocrTrainingManifest(training, remoteEnv, remoteBaseModelPath, remoteAssetPaths)
		},
		AssetProgress: func(done int, total int) string {
			return fmt.Sprintf("syncing OCR training annotation archives [%d/%d ZIP files]", done, total)
		},
	}, progress)
}

func (r *ModelTrainingRemote) stageOCRTrainingAnnotation(tmpDir string, ref *annotation.Reference) (string, error) {
	if ref.DatasetID == "" || ref.ID == "" {
		return "", fmt.Errorf("invalid base annotation reference")
	}
	ann, err := r.annotations.Get(ref.DatasetID, ref.ID)
	if err != nil {
		return "", fmt.Errorf("get base annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
	}
	if !ann.Ocred {
		return "", fmt.Errorf("base annotation %s:%s is not OCRed", ref.DatasetID, ref.ID)
	}
	ds, err := r.datasets.Get(ref.DatasetID)
	if err != nil {
		return "", fmt.Errorf("get dataset %s: %w", ref.DatasetID, err)
	}

	altoDir := r.fileSysMgt.DatasetAnnotationAltoDir(ann)
	imgDir := r.fileSysMgt.DatasetImagesDir(ds)
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

func (r *ModelTrainingRemote) ocrTrainingManifest(training *model.ModelTraining, remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteZipPaths []string) string {
	var b strings.Builder
	r.writeCommonManifest(&b, training, remoteEnv, remoteBaseModelPath)
	fmt.Fprintf(&b, "export MODEL_PREFIX=%s\n", envexec.ShellQuote("kraken_model"))
	b.WriteString("export ZIP_PATHS=(\n")
	for _, zipPath := range remoteZipPaths {
		fmt.Fprintf(&b, "  %s\n", envexec.ShellQuote(zipPath))
	}
	b.WriteString(")\n")
	return b.String()
}
