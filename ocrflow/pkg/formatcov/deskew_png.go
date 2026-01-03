package formatcov

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/envexec"
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
		if err := envexec.Cmd("deskew", "--output", outPath, inPath); err != nil {
			return fmt.Errorf("deskew image %q: %w", inPath, err)
		}
	}

	return nil
}
