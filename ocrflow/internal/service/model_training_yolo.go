package service

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/common"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/envexec"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/formatcov"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/gpufarm"
)

func (r *ModelTrainingRemote) submitYOLO(training *model.ModelTraining, progress func(string)) (*model.ModelTraining, error) {
	mo := training.Model

	tmpDir, err := futils.MkdirTemp("yolo-training")
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
		yoloDir := filepath.Join(tmpDir, "source_yolo", ref.DatasetID+"_"+ref.ID)
		err = r.convertToTemporaryYOLO(ann, ref.DatasetID, ref.ID, yoloDir)
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

func (r *ModelTrainingRemote) convertToTemporaryYOLO(ann *annotation.Annotation, datasetID string, annotationID string, yoloDir string) error {
	ds, err := r.datasets.Get(datasetID)
	if err != nil {
		return fmt.Errorf("get dataset %s for YOLO conversion: %w", datasetID, err)
	}
	if err := os.RemoveAll(yoloDir); err != nil {
		return fmt.Errorf("clear temporary YOLO dir for annotation %s:%s: %w", datasetID, annotationID, err)
	}
	if err := formatcov.Alto2Yolo(r.fileSysMgt.DatasetImagesDir(ds), r.fileSysMgt.DatasetAnnotationAltoDir(ann), yoloDir, 0, "full"); err != nil {
		return fmt.Errorf("convert ALTO to temporary YOLO for annotation %s:%s: %w", datasetID, annotationID, err)
	}
	if err := validateYoloTrainingDir(yoloDir); err != nil {
		return fmt.Errorf("validate temporary YOLO for annotation %s:%s: %w", datasetID, annotationID, err)
	}
	return nil
}

func readYoloLabelmap(yoloDir string) ([]string, error) {
	return formatcov.LoadYoloLabelmap(yoloDir)
}

func validateYoloTrainingDir(yoloDir string) error {
	return formatcov.ValidateYoloDataset(yoloDir)
}

func collectYoloSamples(yoloDir string, datasetID string, annotationID string) ([]yoloSample, error) {
	var samples []yoloSample
	for _, subDir := range formatcov.SubDirs {
		labelDir := filepath.Join(yoloDir, subDir, "labels")
		entries, err := os.ReadDir(labelDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".txt" || entry.Name() == "labelmap.txt" {
				continue
			}
			stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			labelPath := filepath.Join(labelDir, entry.Name())
			imagePath, err := formatcov.FindYoloImage(labelPath)
			if err != nil {
				return nil, err
			}
			nameParts := []string{datasetID, annotationID}
			if subDir != "" {
				nameParts = append(nameParts, subDir)
			}
			nameParts = append(nameParts, stem)
			samples = append(samples, yoloSample{
				imagePath: imagePath,
				labelPath: labelPath,
				name:      strings.Trim(strings.Join(nameParts, "_"), "_"),
			})
		}
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
