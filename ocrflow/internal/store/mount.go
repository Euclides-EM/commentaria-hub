package store

import (
	"log"
	"os"
	"path/filepath"
)

func InitMountProjToStore(rootPath, storePath string) error {
	// Normalize paths to avoid false mismatches
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	rootStorePath := filepath.Join(absRoot, "store")

	absStore, err := filepath.Abs(storePath)
	if err != nil {
		return err
	}

	// Create store dir if it doesn't exist
	if err := os.MkdirAll(absStore, 0o755); err != nil {
		return err
	}

	// If rootPath/store == storePath → nothing to do
	if filepath.Clean(rootStorePath) == filepath.Clean(absStore) {
		log.Printf("store path is the same as root store, no need to initial symlinks between them: %s", absStore)
		return nil
	}

	log.Printf("initializing symlinks from store to root/store: %s -> %s", rootStorePath, absStore)

	if err := createSymlink(absStore, rootStorePath, "data/tps/imgs"); err != nil {
		return err
	}

	if err := createSymlink(absStore, rootStorePath, "data/transcriptions"); err != nil {
		return err
	}

	if err := createSymlink(absStore, rootStorePath, "items_metadata"); err != nil {
		return err
	}

	log.Printf("completed symlinks from store to root/store: %s -> %s", rootStorePath, absStore)

	return nil
}

func createSymlink(storePath, rootPath, rel string) error {
	src := filepath.Join(storePath, rel)
	dst := filepath.Join(rootPath, rel)

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// If already exists, check if correct
	if existing, err := os.Lstat(dst); err == nil {
		if existing.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(dst)
			if err == nil && filepath.Clean(target) == filepath.Clean(src) {
				log.Printf("symlink already, skipping creation: %s -> %s", src, dst)
				return nil
			}
		}
		// Remove incorrect file/symlink
		log.Printf("symlink already exists but points to %s instead of %s, removing it and creating correct symlink", dst, src)
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	log.Printf("creating symlink: %s -> %s", src, dst)
	return os.Symlink(src, dst)
}
