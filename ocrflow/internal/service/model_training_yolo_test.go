package service

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestReadYoloLabelmap(t *testing.T) {
	t.Parallel()

	t.Run("labelmap.txt", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "labelmap.txt"), []byte("MainZone\n  DropCapitalZone  \n\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := readYoloLabelmap(dir)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"MainZone", "DropCapitalZone"}
		if !slices.Equal(got, want) {
			t.Fatalf("labels = %v, want %v", got, want)
		}
	})

	t.Run("data.yaml", func(t *testing.T) {
		dir := t.TempDir()
		data := "names: ['MainZone', 'DropCapitalZone']\n"
		if err := os.WriteFile(filepath.Join(dir, "data.yaml"), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := readYoloLabelmap(dir)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"MainZone", "DropCapitalZone"}
		if !slices.Equal(got, want) {
			t.Fatalf("labels = %v, want %v", got, want)
		}
	})

	t.Run("missing metadata", func(t *testing.T) {
		if _, err := readYoloLabelmap(t.TempDir()); err == nil {
			t.Fatal("expected missing metadata to return an error")
		}
	})
}

func TestValidateYoloTrainingDirRequiresSamples(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "labelmap.txt"), []byte("MainZone\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateYoloTrainingDir(dir); err == nil {
		t.Fatal("metadata without image/label pairs reported as usable")
	}

	if err := os.MkdirAll(filepath.Join(dir, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "images", "page-0001.png"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := validateYoloTrainingDir(dir); err != nil {
		t.Fatalf("complete flat YOLO directory rejected: %v", err)
	}
}

func TestCollectYoloSamplesFromSplitLayout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "train", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "train", "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "train", "images", "page-0001.jpg"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "train", "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	samples, err := collectYoloSamples(dir, "ds_1", "ann_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 {
		t.Fatalf("got %d samples, want 1", len(samples))
	}
	if samples[0].name != "ds_1_ann_1_train_page-0001" {
		t.Fatalf("sample name = %q", samples[0].name)
	}
}

func TestCollectYoloSamplesRejectsMissingBundleImage(t *testing.T) {
	t.Parallel()

	yoloDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(yoloDir, "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(yoloDir, "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := collectYoloSamples(yoloDir, "ds_1", "ann_1"); err == nil {
		t.Fatal("expected a label without a matching bundle image to fail")
	}
}

func TestCollectYoloSamplesFromValLayout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "val", "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "val", "labels"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "val", "images", "page-0001.jpg"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "val", "labels", "page-0001.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	samples, err := collectYoloSamples(dir, "ds_1", "ann_1")
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 1 {
		t.Fatalf("got %d samples, want 1", len(samples))
	}
}

func TestAppendUniqueYoloClasses(t *testing.T) {
	t.Parallel()

	got := appendUniqueYoloClasses(
		[]string{"MainZone", "RunningTitleZone"},
		[]string{"RunningTitleZone", "MarginTextZone", "MainZone-Head--Section"},
	)
	want := []string{"MainZone", "RunningTitleZone", "MarginTextZone", "MainZone-Head--Section"}
	if !slices.Equal(got, want) {
		t.Fatalf("classes = %v, want %v", got, want)
	}
}

func TestRemapYoloLabelFile(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "source.txt")
	dst := filepath.Join(t.TempDir(), "labels", "remapped.txt")
	if err := os.WriteFile(src, []byte("0 0.5 0.5 0.2 0.2\n2 0.1 0.2 0.3 0.4 0.5 0.6\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	sourceClasses := []string{"RunningTitleZone", "MainZone", "MarginTextZone"}
	targetClassIDs := map[string]int{
		"MainZone":         0,
		"MarginTextZone":   1,
		"RunningTitleZone": 2,
	}
	if err := remapYoloLabelFile(src, dst, sourceClasses, targetClassIDs); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	want := "2 0.5 0.5 0.2 0.2\n1 0.1 0.2 0.3 0.4 0.5 0.6\n"
	if string(raw) != want {
		t.Fatalf("remapped labels = %q, want %q", raw, want)
	}
}

func TestRemapYoloLabelFileRejectsUnknownClassID(t *testing.T) {
	t.Parallel()

	src := filepath.Join(t.TempDir(), "source.txt")
	if err := os.WriteFile(src, []byte("3 0.5 0.5 0.2 0.2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := remapYoloLabelFile(src, filepath.Join(t.TempDir(), "output.txt"), []string{"MainZone"}, map[string]int{"MainZone": 0})
	if err == nil {
		t.Fatal("expected out-of-range class ID to fail")
	}
}

func TestAssignYoloSplitsIsDeterministicAndStratified(t *testing.T) {
	t.Parallel()

	buildSamples := func() []yoloSample {
		var samples []yoloSample
		for idx := 0; idx < 20; idx++ {
			samples = append(samples, yoloSample{datasetID: "ds_a", pageID: fmt.Sprintf("page-%04d", idx), name: fmt.Sprintf("a-%04d", idx)})
		}
		for idx := 0; idx < 10; idx++ {
			samples = append(samples, yoloSample{datasetID: "ds_b", pageID: fmt.Sprintf("page-%04d", idx), name: fmt.Sprintf("b-%04d", idx)})
		}
		return samples
	}

	first := buildSamples()
	second := buildSamples()
	assignYoloSplits(first, 42)
	assignYoloSplits(second, 42)

	counts := make(map[string]map[string]int)
	for idx := range first {
		if first[idx].split != second[idx].split {
			t.Fatalf("split is not deterministic for %s: %s != %s", first[idx].name, first[idx].split, second[idx].split)
		}
		if counts[first[idx].datasetID] == nil {
			counts[first[idx].datasetID] = make(map[string]int)
		}
		counts[first[idx].datasetID][first[idx].split]++
	}
	for datasetID, want := range map[string]map[string]int{
		"ds_a": {"train": 16, "valid": 2, "test": 2},
		"ds_b": {"train": 8, "valid": 1, "test": 1},
	} {
		if !maps.Equal(counts[datasetID], want) {
			t.Fatalf("%s split counts = %v, want %v", datasetID, counts[datasetID], want)
		}
	}
}

func TestAssignYoloSplitsKeepsDuplicatePagesTogether(t *testing.T) {
	t.Parallel()

	samples := []yoloSample{
		{datasetID: "ds_a", pageID: "page-0001", name: "ann_a_page-0001"},
		{datasetID: "ds_a", pageID: "page-0001", name: "ann_b_page-0001"},
		{datasetID: "ds_a", pageID: "page-0002", name: "ann_a_page-0002"},
		{datasetID: "ds_a", pageID: "page-0003", name: "ann_a_page-0003"},
	}
	assignYoloSplits(samples, 42)
	if samples[0].split != samples[1].split {
		t.Fatalf("duplicate page leaked across splits: %s != %s", samples[0].split, samples[1].split)
	}
}

func TestAssignYoloSplitsFillsGlobalSplitsForSingletonDatasets(t *testing.T) {
	t.Parallel()

	samples := []yoloSample{
		{datasetID: "ds_a", pageID: "page-0001"},
		{datasetID: "ds_b", pageID: "page-0001"},
		{datasetID: "ds_c", pageID: "page-0001"},
	}
	assignYoloSplits(samples, 42)
	for _, split := range []string{"train", "valid", "test"} {
		if !hasYoloSplit(samples, split) {
			t.Fatalf("missing %s split: %+v", split, samples)
		}
	}
}

func TestYoloDataYAMLIncludesAllSplits(t *testing.T) {
	t.Parallel()

	got := yoloDataYAML([]string{"MainZone"})
	for _, want := range []string{"train: train/images", "val: valid/images", "test: test/images"} {
		if !strings.Contains(got, want) {
			t.Fatalf("data.yaml missing %q: %s", want, got)
		}
	}
}
