package futils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var TmpDir = os.TempDir()

func SetTmpDir(dir string) {
	if dir == "" {
		dir = os.TempDir()
	}
	log.Printf("Setting tmp dir to: %s", dir)
	TmpDir = dir
}

func MkdirTemp(pattern string) (string, error) {
	return os.MkdirTemp(TmpDir, fmt.Sprintf("ocrflow-%s-*", pattern))
}

func CreateTemp(pattern string) (*os.File, error) {
	return CreateTempInDir(TmpDir, pattern)
}

func CreateTempInDir(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, fmt.Sprintf("ocrflow-%s", pattern))
}

func CleanTemp() {
	log.Printf("Deleting temp dir files and directories in %q", TmpDir)
	globPattern := filepath.Join(TmpDir, "ocrflow-*")
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		log.Fatalf("Failed to match %q: %v", globPattern, err)
	}

	for _, match := range matches {
		log.Printf("Deleting temp file or directory: %q", match)
		if err := os.RemoveAll(match); err != nil {
			log.Printf("Failed to remove %q: %v", match, err)
		}
	}

	log.Printf("Finished deleting temp files and directories in %q", TmpDir)
}

func InitTemp(dir string) {
	SetTmpDir(dir)
	CleanTemp()
}
