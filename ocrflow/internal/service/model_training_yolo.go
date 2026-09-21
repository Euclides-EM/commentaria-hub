package service

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	imagePath  string
	labelPath  string
	name       string
	datasetID  string
	pageID     string
	classNames []string
	split      string
}

const yoloSplitSeed int64 = 42

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
		classNames = appendUniqueYoloClasses(classNames, names)

		annSamples, err := collectYoloSamples(yoloDir, ref.DatasetID, ref.ID)
		if err != nil {
			return "", 0, fmt.Errorf("collect YOLO samples for annotation %s:%s: %w", ref.DatasetID, ref.ID, err)
		}
		for idx := range annSamples {
			annSamples[idx].classNames = names
		}
		samples = append(samples, annSamples...)
	}
	if len(samples) == 0 {
		return "", 0, fmt.Errorf("no YOLO image/label pairs found in base annotations")
	}
	if len(classNames) == 0 {
		return "", 0, fmt.Errorf("no YOLO classes found in base annotations")
	}

	assignYoloSplits(samples, yoloSplitSeed)
	sort.Slice(samples, func(i, k int) bool { return samples[i].name < samples[k].name })
	classIDs := make(map[string]int, len(classNames))
	for idx, name := range classNames {
		classIDs[name] = idx
	}
	for _, sample := range samples {
		if err := futils.CopyFile(sample.imagePath, filepath.Join(stageDir, sample.split, "images", sample.name+filepath.Ext(sample.imagePath))); err != nil {
			return "", 0, fmt.Errorf("copy YOLO image %s: %w", sample.imagePath, err)
		}
		if err := remapYoloLabelFile(sample.labelPath, filepath.Join(stageDir, sample.split, "labels", sample.name+".txt"), sample.classNames, classIDs); err != nil {
			return "", 0, fmt.Errorf("remap YOLO label %s: %w", sample.labelPath, err)
		}
	}
	if !hasYoloSplit(samples, "valid") {
		// With a single unique page there is no leak-free split. Duplicate it so
		// Ultralytics still receives the required validation directory.
		sample := samples[0]
		if err := futils.CopyFile(sample.imagePath, filepath.Join(stageDir, "valid", "images", sample.name+filepath.Ext(sample.imagePath))); err != nil {
			return "", 0, fmt.Errorf("copy YOLO validation image %s: %w", sample.imagePath, err)
		}
		if err := remapYoloLabelFile(sample.labelPath, filepath.Join(stageDir, "valid", "labels", sample.name+".txt"), sample.classNames, classIDs); err != nil {
			return "", 0, fmt.Errorf("remap YOLO validation label %s: %w", sample.labelPath, err)
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

func appendUniqueYoloClasses(dst []string, src []string) []string {
	seen := make(map[string]struct{}, len(dst)+len(src))
	for _, name := range dst {
		seen[name] = struct{}{}
	}
	for _, name := range src {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		dst = append(dst, name)
	}
	return dst
}

// assignYoloSplits performs a reproducible 80/10/10 split within each source
// dataset. Samples for the same canonical page stay together so alternate
// annotations of a page cannot leak across train, validation, and test sets.
func assignYoloSplits(samples []yoloSample, seed int64) {
	type pageGroup struct {
		key     string
		indices []int
	}

	datasetPages := make(map[string]map[string][]int)
	for idx := range samples {
		pages := datasetPages[samples[idx].datasetID]
		if pages == nil {
			pages = make(map[string][]int)
			datasetPages[samples[idx].datasetID] = pages
		}
		pages[samples[idx].pageID] = append(pages[samples[idx].pageID], idx)
	}

	datasetIDs := make([]string, 0, len(datasetPages))
	for datasetID := range datasetPages {
		datasetIDs = append(datasetIDs, datasetID)
	}
	sort.Strings(datasetIDs)
	rng := rand.New(rand.NewSource(seed))
	var allGroups []pageGroup
	for _, datasetID := range datasetIDs {
		pages := datasetPages[datasetID]
		groups := make([]pageGroup, 0, len(pages))
		for pageID, indices := range pages {
			groups = append(groups, pageGroup{key: pageID, indices: indices})
		}
		sort.Slice(groups, func(i, j int) bool { return groups[i].key < groups[j].key })
		rng.Shuffle(len(groups), func(i, j int) { groups[i], groups[j] = groups[j], groups[i] })
		allGroups = append(allGroups, groups...)

		trainCount, validCount := yoloSplitCounts(len(groups))
		for groupIdx, group := range groups {
			split := "test"
			if groupIdx < trainCount {
				split = "train"
			} else if groupIdx < trainCount+validCount {
				split = "valid"
			}
			for _, sampleIdx := range group.indices {
				samples[sampleIdx].split = split
			}
		}
	}

	// Tiny per-dataset groups can leave a global split empty. Rebalance whole
	// page groups deterministically while retaining at least one training page.
	ensureSplit := func(target string, minimumGroups int) {
		if len(allGroups) < minimumGroups || hasYoloSplit(samples, target) {
			return
		}
		for idx := len(allGroups) - 1; idx >= 0; idx-- {
			group := allGroups[idx]
			if samples[group.indices[0]].split != "train" {
				continue
			}
			for _, sampleIdx := range group.indices {
				samples[sampleIdx].split = target
			}
			return
		}
	}
	ensureSplit("valid", 2)
	ensureSplit("test", 3)
}

func hasYoloSplit(samples []yoloSample, split string) bool {
	for idx := range samples {
		if samples[idx].split == split {
			return true
		}
	}
	return false
}

func yoloSplitCounts(total int) (train int, valid int) {
	switch total {
	case 0:
		return 0, 0
	case 1:
		return 1, 0
	case 2:
		return 1, 1
	}
	train = total * 8 / 10
	remaining := total - train
	valid = remaining / 2
	test := remaining - valid
	if valid == 0 {
		valid++
		train--
	}
	if test == 0 {
		train--
	}
	return train, valid
}

func remapYoloLabelFile(src string, dst string, sourceClasses []string, targetClassIDs map[string]int) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	lines := strings.Split(string(raw), "\n")
	for idx, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		sourceID, err := strconv.Atoi(fields[0])
		if err != nil {
			return fmt.Errorf("line %d has invalid class ID %q", idx+1, fields[0])
		}
		if sourceID < 0 || sourceID >= len(sourceClasses) {
			return fmt.Errorf("line %d class ID %d is outside label map with %d classes", idx+1, sourceID, len(sourceClasses))
		}
		targetID, ok := targetClassIDs[sourceClasses[sourceID]]
		if !ok {
			return fmt.Errorf("line %d class %q is missing from merged label map", idx+1, sourceClasses[sourceID])
		}
		fields[0] = strconv.Itoa(targetID)
		lines[idx] = strings.Join(fields, " ")
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(strings.Join(lines, "\n")), 0o644)
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
				datasetID: datasetID,
				pageID:    stem,
			})
		}
	}
	return samples, nil
}

func yoloDataYAML(classNames []string) string {
	var b strings.Builder
	b.WriteString("train: train/images\n")
	b.WriteString("val: valid/images\n")
	b.WriteString("test: test/images\n")
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
