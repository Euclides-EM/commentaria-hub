package futils

import (
	"io"
	"mime/multipart"
	"os"
	"path"
)

func WriteMultipartFileToPath(src multipart.File, dstPath string) error {
	defer src.Close()

	if err := os.MkdirAll(path.Dir(dstPath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
