package service

import (
	"os"
	"path/filepath"
	"slices"
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
