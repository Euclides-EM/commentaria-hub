package httpwrapper

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
)

// StoreUncompressedDirFromRequest extracts a ZIP file from the HTTP request and stores its contents in the specified destination path.
func StoreUncompressedDirFromRequest(dstPath string, r *http.Request) error {
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		return fmt.Errorf("invalid multipart form: %w", err)
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return fmt.Errorf("missing form file: %w", err)
	}
	defer file.Close()
	if filepath.Ext(header.Filename) != ".zip" {
		return fmt.Errorf("only .zip files are allowed")
	}

	return futils.UnzipFromReader(dstPath, file)
}
