package futils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Unzip(srcZip, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	return extractZipWithMode(srcZip, destDir, true)
}

func extractZip(src, dst string) error {
	return extractZipWithMode(src, dst, false)
}

func extractZipWithMode(src, dst string, preserveMode bool) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, item := range r.File {
		info := item.FileInfo()
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink in zip archive: %s", item.Name)
		}

		target, err := SafeJoin(dst, item.Name)
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := os.MkdirAll(target, zipEntryMode(info, preserveMode, 0o755)); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		in, err := item.Open()
		if err != nil {
			return err
		}

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipEntryMode(info, preserveMode, 0o644))
		if err != nil {
			in.Close()
			return err
		}

		_, copyErr := io.Copy(out, in)
		closeErr := out.Close()
		in.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func zipEntryMode(info os.FileInfo, preserveMode bool, fallback os.FileMode) os.FileMode {
	if preserveMode {
		return info.Mode()
	}
	return fallback
}

func Zip(srcDir, destZip string) error {
	outFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	fi, err := os.Stat(srcDir)
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not stat source: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("source path is not a directory")
	}
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		fh, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		fh.Name = relPath
		fh.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(fh)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func UnzipFromReader(dstPath string, file io.Reader) error {
	dst, err := CreateTemp("upload-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()
	defer os.Remove(dst.Name())

	_, err = io.Copy(dst, file)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	if err := Unzip(dst.Name(), dstPath); err != nil {
		return fmt.Errorf("failed to unzip file: %w", err)
	}

	return nil
}
