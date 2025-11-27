package httpwrapper

import (
	"archive/zip"
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// StoreUncompressedDir extracts a ZIP file from the HTTP request and stores its contents in the specified destination path.
func StoreUncompressedDir(dstPath string, r *http.Request) error {
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

	dst, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()
	defer os.Remove(dst.Name())

	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	if err := unzip(dst.Name(), dstPath); err != nil {
		return fmt.Errorf("failed to unzip file: %w", err)
	}

	return nil
}

func unzip(srcZip, destDir string) error {
	r, err := zip.OpenReader(srcZip)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath, err := futils.SafeJoin(destDir, f.Name)
		if err != nil {
			return fmt.Errorf("invalid file path: %w", err)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, f.Mode()); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0o755); err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			rc.Close()
			return err
		}

		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return err
		}
		rc.Close()
		outFile.Close()
	}
	return nil
}
