package formatcov

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func DeskewPNGs(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read src dir %q: %w", src, err)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("create dst dir %q: %w", dst, err)
	}

	for i, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".png" {
			continue
		}

		inPath := filepath.Join(src, name)
		outPath := filepath.Join(dst, name)

		log.Printf("[%d/%d] Deskewing %q -> %q", i+1, len(entries), inPath, outPath)
		// todo: use embedded python script + probably batch is possible...
		return envexec.Cmd("deskew", "--output", outPath, inPath)
	}

	return nil
}
