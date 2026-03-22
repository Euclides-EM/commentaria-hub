package store

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func InitMountProjToStore(rootPath, storePath string) error {
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	repoStorePath := filepath.Join(absRoot, "store")

	absStore, err := filepath.Abs(storePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absStore, 0o755); err != nil {
		return err
	}

	if filepath.Clean(repoStorePath) == filepath.Clean(absStore) {
		log.Printf("store path is the same as repo store, nothing to mount: %s", absStore)
		return nil
	}

	log.Printf("linking runtime store to repo store: %s -> %s", absStore, repoStorePath)

	for _, rel := range []string{
		"data/tps/imgs",
		"data/transcriptions",
		"items_metadata",
	} {
		target := filepath.Join(repoStorePath, rel) // real data (repo)
		link := filepath.Join(absStore, rel)        // runtime path

		// ensure parent dirs exist
		if err := os.MkdirAll(filepath.Dir(link), 0o755); err != nil {
			return err
		}

		existing, err := os.Lstat(link)
		if err == nil {
			if existing.Mode()&os.ModeSymlink != 0 {
				currentTarget, err := os.Readlink(link)
				if err == nil && filepath.Clean(currentTarget) == filepath.Clean(target) {
					log.Printf("symlink already correct, skipping: %s -> %s", link, target)
					continue
				}
				// wrong symlink → safe to remove
				if err := os.Remove(link); err != nil {
					return err
				}
			} else {
				// IMPORTANT: do NOT delete real data
				return fmt.Errorf("runtime path exists and is not a symlink: %s", link)
			}
		} else if !os.IsNotExist(err) {
			return err
		}

		log.Printf("creating symlink: %s -> %s", link, target)
		if err := os.Symlink(target, link); err != nil {
			return err
		}
	}

	log.Printf("completed linking runtime store to repo store")
	return nil
}
