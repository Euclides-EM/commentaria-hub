package service

import (
	"path/filepath"
	"testing"
)

func TestNewVCSMgtKeepsRepoPathAndRelativeWatchedPaths(t *testing.T) {
	root := t.TempDir()
	itemsMetadataDir := filepath.Join(root, "items_metadata")
	titlePageImgDir := filepath.Join(root, "data", "titlepages", "imgs")

	vcs := NewVCSMgt(root, itemsMetadataDir, titlePageImgDir)

	if vcs.repoPath != root {
		t.Fatalf("repoPath = %q, want %q", vcs.repoPath, root)
	}
	if vcs.itemsMetadataStoreDir != "items_metadata" {
		t.Fatalf("itemsMetadataStoreDir = %q, want %q", vcs.itemsMetadataStoreDir, "items_metadata")
	}
	wantTitlePageImgDir := filepath.Join("data", "titlepages", "imgs")
	if vcs.titlePageImgDir != wantTitlePageImgDir {
		t.Fatalf("titlePageImgDir = %q, want %q", vcs.titlePageImgDir, wantTitlePageImgDir)
	}
}

func TestVCSMgtWatchedPathspecsExcludeTitlePageVariants(t *testing.T) {
	vcs := &VCSMgt{
		itemsMetadataStoreDir: filepath.Join("ocrflow", "store", "items_metadata"),
		titlePageImgDir:       filepath.Join("ocrflow", "store", "data", "tps", "imgs"),
	}

	got := vcs.watchedPathspecs()
	want := []string{
		"ocrflow/store/data/tps/imgs",
		"ocrflow/store/items_metadata",
		":(exclude)ocrflow/store/data/tps/imgs/_variants",
	}
	if len(got) != len(want) {
		t.Fatalf("watchedPathspecs() length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("watchedPathspecs()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
