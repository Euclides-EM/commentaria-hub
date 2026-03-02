package formatcov

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ErrDeskewUnavailable is returned by DeskewPNGs when the binary was built
// without OpenCV (no -tags gocv). On the server you can build without gocv
// to avoid OpenCV; deskew will be skipped and images copied as-is.
var ErrDeskewUnavailable = errors.New("deskew unavailable: build with -tags gocv and install OpenCV")

// CopyPNGs copies all .png files from srcDir to dstDir (creates dstDir if needed).
// Used when deskew is requested but gocv is not available.
func CopyPNGs(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read src dir %q: %w", srcDir, err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("create dst dir %q: %w", dstDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.ToLower(filepath.Ext(name)) != ".png" {
			continue
		}
		srcPath := filepath.Join(srcDir, name)
		dstPath := filepath.Join(dstDir, name)
		if err := copyFile(srcPath, dstPath); err != nil {
			return fmt.Errorf("copy %q -> %q: %w", srcPath, dstPath, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	return err
}
