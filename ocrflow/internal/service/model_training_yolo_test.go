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
