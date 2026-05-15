package service

import (
	"path/filepath"
	"testing"
)

func TestNewVCSMgtKeepsRepoPathAndRelativeWatchedPaths(t *testing.T) {
	root := t.TempDir()
	itemsMetadataDir := filepath.Join(root, "items_metadata")
	titlePageImgDir := filepath.Join(root, "data", "titlepages", "imgs")

	vcs := NewVCSMgt(itemsMetadataDir, titlePageImgDir)

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
