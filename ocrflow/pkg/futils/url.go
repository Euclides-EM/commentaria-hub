package futils

import (
	"net/url"
	"path/filepath"
	"strings"
)

func IsLocalFileURL(s string) bool {

	if _, err := url.Parse(s); err != nil {
		return false
	}
	_, err := URLToLocalFilePath(s)
	return err == nil
}

func URLToLocalFilePath(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(u.Scheme, "file") {
		return "", &url.Error{Op: "parse", URL: s, Err: err}
	}
	if u.Host != "" && !strings.EqualFold(u.Host, "localhost") {
		return "", &url.Error{Op: "parse", URL: s, Err: err}
	}
	return u.Path, nil
}

func LocalFilePathToURL(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return (&url.URL{Scheme: "file", Path: absPath}).String(), nil
}
