package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/gpufarm"
)

func (r *ModelTrainingRemote) submitYOLO(training *model.ModelTraining, progress func(string)) (*model.ModelTraining, error) {
	mo := training.Model

	tmpDir, err := futils.MkdirTemp("ocrflow-yolo-training-*")
	if err != nil {
		return nil, fmt.Errorf("create YOLO training temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	datasetZipPath, imageCount, err := r.stageYOLOTrainingDataset(tmpDir, mo.BaseAnnotations)
	if err != nil {
		return nil, err
	}

	baseModelPath := ""
	if mo.BaseModelID != "" {
		baseModelPath, err = r.localBaseModelPath(mo.BaseModelID, common.OCRModelTypeSegment, "a segmentation model")
		if err != nil {
			return nil, err
		}
	}

	return r.submit(trainingRemoteRequest{
		Training:      training,
		TmpDir:        tmpDir,
		JobName:       "train_yolo",
		DisplayName:   "YOLO",
		BaseModelPath: baseModelPath,
		Assets: []trainingRemoteAsset{{
			LocalPath: datasetZipPath,
			AssetDir:  "datasets",
		}},
		StatusDetails: map[string]string{
			"training_images": fmt.Sprintf("%d", imageCount),
		},
		Manifest: func(remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteAssetPaths []string) string {
			return r.yoloTrainingManifest(training, remoteEnv, remoteBaseModelPath, remoteAssetPaths[0])
		},
		AssetProgress: func(done int, total int) string {
			return fmt.Sprintf("syncing YOLO training dataset archive [%d images]", imageCount)
		},
	}, progress)
}

type yoloSample struct {
	imagePath string
	labelPath string
	name      string
}

func (r *ModelTrainingRemote) stageYOLOTrainingDataset(tmpDir string, refs []*annotation.Reference) (string, int, error) {
	stageDir := filepath.Join(tmpDir, "dataset")
	var samples []yoloSample
	var classNames []string

	for _, ref := range refs {
		if ref == nil {
			continue
		}
		ann, err := r.annotations.Get(ref.DatasetID, ref.ID)
		if err != nil {
			return "", 0, fmt.Errorf("get base annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
		}
		if !ann.Segmented {
			return "", 0, fmt.Errorf("base annotation %s:%s is not segmented", ref.DatasetID, ref.ID)
		}
		yoloDir, err := r.getOrConvertToYOLO(ann, ref.DatasetID, ref.ID)
		if err != nil {
			return "", 0, err
		}
		names, err := readYoloLabelmap(yoloDir)
		if err != nil {
			return "", 0, fmt.Errorf("read YOLO labels for annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
		}
		if len(classNames) == 0 {
			classNames = names
		} else if strings.Join(classNames, "\n") != strings.Join(names, "\n") {
			return "", 0, fmt.Errorf("base annotation %s:%s has a different YOLO labelmap", ref.DatasetID, ref.ID)
		}

		annSamples, err := collectYoloSamples(yoloDir, ref.DatasetID, ref.ID)
		if err != nil {
			return "", 0, fmt.Errorf("collect YOLO samples for annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
		}
		samples = append(samples, annSamples...)
	}
	if len(samples) == 0 {
		return "", 0, fmt.Errorf("no YOLO image/label pairs found in base annotations")
	}
	if len(classNames) == 0 {
		return "", 0, fmt.Errorf("no YOLO classes found in base annotations")
	}

	sort.Slice(samples, func(i, k int) bool { return samples[i].name < samples[k].name })
	for idx, sample := range samples {
		split := yoloSplit(idx, len(samples))
		if err := futils.CopyFile(sample.imagePath, filepath.Join(stageDir, split, "images", sample.name+filepath.Ext(sample.imagePath))); err != nil {
			return "", 0, fmt.Errorf("copy YOLO image %s: %w", sample.imagePath, err)
		}
		if err := futils.CopyFile(sample.labelPath, filepath.Join(stageDir, split, "labels", sample.name+".txt")); err != nil {
			return "", 0, fmt.Errorf("copy YOLO label %s: %w", sample.labelPath, err)
		}
	}
	if len(samples) == 1 {
		sample := samples[0]
		if err := futils.CopyFile(sample.imagePath, filepath.Join(stageDir, "valid", "images", sample.name+filepath.Ext(sample.imagePath))); err != nil {
			return "", 0, fmt.Errorf("copy YOLO validation image %s: %w", sample.imagePath, err)
		}
		if err := futils.CopyFile(sample.labelPath, filepath.Join(stageDir, "valid", "labels", sample.name+".txt")); err != nil {
			return "", 0, fmt.Errorf("copy YOLO validation label %s: %w", sample.labelPath, err)
		}
	}

	if err := os.WriteFile(filepath.Join(stageDir, "data.yaml"), []byte(yoloDataYAML(classNames)), 0o644); err != nil {
		return "", 0, fmt.Errorf("write YOLO data.yaml: %w", err)
	}

	zipPath := filepath.Join(tmpDir, "yolo_training_dataset.zip")
	if err := futils.Zip(stageDir, zipPath); err != nil {
		return "", 0, fmt.Errorf("zip YOLO training dataset: %w", err)
	}
	return zipPath, len(samples), nil
}

func (r *ModelTrainingRemote) getOrConvertToYOLO(ann *annotation.Annotation, datasetID string, annotationID string) (string, error) {
	yoloDir := r.fileSysMgt.DatasetAnnotationYoloDir(ann)
	if _, err := os.Stat(yoloDir); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("stat YOLO dir for annotation %s:%s: %w", datasetID, annotationID, err)
	} else if err == nil {
		return yoloDir, nil
	}

	ds, err := r.datasets.Get(datasetID)
	if err != nil {
		return "", fmt.Errorf("get dataset %s for YOLO conversion: %w", datasetID, err)
	}
	if err := os.RemoveAll(yoloDir); err != nil {
		return "", fmt.Errorf("clear stale YOLO dir for annotation %s:%s: %w", datasetID, annotationID, err)
	}
	if err := formatcov.Alto2Yolo(r.fileSysMgt.DatasetImagesDir(ds), r.fileSysMgt.DatasetAnnotationAltoDir(ann), yoloDir, 0, "full"); err != nil {
		return "", fmt.Errorf("convert ALTO to YOLO for annotation %s:%s: %w", datasetID, annotationID, err)
	}
	return yoloDir, nil
}

func readYoloLabelmap(yoloDir string) ([]string, error) {
	raw, err := os.ReadFile(filepath.Join(yoloDir, "labelmap.txt"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, line := range strings.Split(string(raw), "\n") {
		name := strings.TrimSpace(line)
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

func collectYoloSamples(yoloDir string, datasetID string, annotationID string) ([]yoloSample, error) {
	imageDir := filepath.Join(yoloDir, "images")
	labelDir := filepath.Join(yoloDir, "labels")
	entries, err := os.ReadDir(imageDir)
	if err != nil {
		return nil, err
	}
	var samples []yoloSample
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue
		}
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		labelPath := filepath.Join(labelDir, stem+".txt")
		if _, err := os.Stat(labelPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		samples = append(samples, yoloSample{
			imagePath: filepath.Join(imageDir, entry.Name()),
			labelPath: labelPath,
			name:      datasetID + "_" + annotationID + "_" + stem,
		})
	}
	return samples, nil
}

func yoloSplit(idx int, total int) string {
	if total < 5 {
		if idx == total-1 && total > 1 {
			return "valid"
		}
		return "train"
	}
	if idx >= total*8/10 {
		return "valid"
	}
	return "train"
}

func yoloDataYAML(classNames []string) string {
	var b strings.Builder
	b.WriteString("train: train/images\n")
	b.WriteString("val: valid/images\n")
	b.WriteString(fmt.Sprintf("nc: %d\n", len(classNames)))
	b.WriteString("names:\n")
	for _, name := range classNames {
		fmt.Fprintf(&b, "- %q\n", name)
	}
	return b.String()
}

func (r *ModelTrainingRemote) yoloTrainingManifest(training *model.ModelTraining, remoteEnv *gpufarm.RemoteEnv, remoteBaseModelPath string, remoteDatasetZipPath string) string {
	var b strings.Builder
	r.writeCommonManifest(&b, training, remoteEnv, remoteBaseModelPath)
	fmt.Fprintf(&b, "export DATASET_ZIP_PATH=%s\n", envexec.ShellQuote(remoteDatasetZipPath))
	return b.String()
}
