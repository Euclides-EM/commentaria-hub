package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferPagesFromAltoDirIgnoresNonXMLArtifacts(t *testing.T) {
	dir := t.TempDir()
	files := []string{
		"page-0198.xml",
		"page-0199.xml",
		"page-0198-MainZone-baselines.json",
		"page-0198-MainZone-mask.png",
		"page-0199-DropCapitalZone-Plain-mask.png",
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	pages, err := inferPagesFromAltoDir(dir)
	if err != nil {
		t.Fatalf("infer ALTO pages: %v", err)
	}
	if len(pages) != 2 || pages[0] != 198 || pages[1] != 199 {
		t.Fatalf("pages = %v, want [198 199]", pages)
	}
}
