package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/samber/lo"
)

const maxBackupsToStore = 5

type Backup struct {
	baseDataDir      string
	modelsDir        string
	itemsMetadataDir string
	dbPath           string

	backupdir             string
	restoreFromBackupPath string

	shutdownFunc func() error
}

func NewBackup(baseData, models, items, db, backupDir, restoreDir string, shutdownFunc func() error) *Backup {
	return &Backup{
		baseDataDir:           baseData,
		modelsDir:             models,
		itemsMetadataDir:      items,
		dbPath:                db,
		backupdir:             backupDir,
		restoreFromBackupPath: restoreDir,
		shutdownFunc:          shutdownFunc,
	}
}

// ListBackups lists available backups.
func (s *Backup) ListBackups() ([]string, error) {
	entries, err := os.ReadDir(s.backupdir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("list backups: read dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()

		// case-insensitive check for .zip, but preserve original name
		if strings.EqualFold(filepath.Ext(n), ".zip") {
			names = append(names, strings.TrimSuffix(n, filepath.Ext(n)))
		}
	}

	sort.Slice(names, func(i, j int) bool {
		return names[i] > names[j] // timestamped names sort naturally
	})

	return names, nil
}

// CreateBackup creates a new backup of the current system state.
func (s *Backup) CreateBackup() (string, error) {
	if err := s.ensureBackupDir(); err != nil {
		return "", err
	}

	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("backup_%s.zip", ts)
	dst := filepath.Join(s.backupdir, name)

	// Overwrite protection: if exists, add a monotonic suffix.
	for fileExists(dst) {
		return "", fmt.Errorf("backup: file already exists: %s", dst)
	}

	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create backup: create zip: %w", err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	defer func() { _ = zw.Close() }()

	// Keep paths stable inside the ZIP, independent of absolute paths.
	// Layout:
	//   base_data/...    (contents of baseDataDir)
	//   models/...       (contents of modelsDir)
	//   items_metadata/... (contents of itemsMetadataDir)
	//   db/<basename>    (dbPath file)
	if err := addDirToZip(zw, s.baseDataDir, "base_data"); err != nil {
		return "", err
	}
	if err := addDirToZip(zw, s.modelsDir, "models"); err != nil {
		return "", err
	}
	if err := addDirToZip(zw, s.itemsMetadataDir, "items_metadata"); err != nil {
		return "", err
	}
	if err := addFileToZip(zw, s.dbPath, filepath.Join("db", filepath.Base(s.dbPath))); err != nil {
		return "", err
	}

	log.Printf("backup: created %s", dst)

	if err := s.ensureMaxBackups(); err != nil {
		log.Printf("warning: failed to ensure max backups after creating backup from zip: %v", err)
	}

	return strings.TrimSuffix(name, ".zip"), nil
}

// SetupRestoreBackup restores the system state from the backup with the given ID (or latest if backupId is empty or "latest").
func (s *Backup) SetupRestoreBackup(backupId string) error {
	backups, err := s.ListBackups()
	if err != nil {
		return err
	}
	if len(backups) == 0 {
		return errors.New("setup restore: no backups found")
	}

	backupToRestore := backupId
	if backupToRestore == "" || backupToRestore == "latest" {
		backupToRestore = backups[0]
	} else {
		if !lo.Contains(backups, backupToRestore) {
			return fmt.Errorf("setup restore: backup not found: %s", backupToRestore)
		}
	}

	log.Printf("backup: selected backup for restore: %s", backupToRestore)

	latest := filepath.Join(s.backupdir, backupToRestore+".zip")
	if !fileExists(latest) {
		return fmt.Errorf("setup restore: backup file not found: %s", latest)
	}

	// Extract into restoreFromBackupPath (clean first).
	if s.restoreFromBackupPath == "" {
		return errors.New("setup restore: restoreFromBackupPath is empty")
	}
	if err := os.RemoveAll(s.restoreFromBackupPath); err != nil {
		return fmt.Errorf("setup restore: clean restore dir: %w", err)
	}
	if err := os.MkdirAll(s.restoreFromBackupPath, 0o755); err != nil {
		return fmt.Errorf("setup restore: create restore dir: %w", err)
	}

	if err := unzip(latest, s.restoreFromBackupPath); err != nil {
		return fmt.Errorf("setup restore: unzip latest backup: %w", err)
	}

	log.Printf("backup: extracted latest backup %s to %s", latest, s.restoreFromBackupPath)

	// Shutdown so restore can happen safely on next start.
	if s.shutdownFunc != nil {
		log.Printf("backup: shutdown function")
		time.Sleep(5 * time.Second) // Give some time for logs to flush etc.
		if err := s.shutdownFunc(); err != nil {
			return fmt.Errorf("setup restore: shutdown: %w", err)
		}
	}
	return nil
}

func (s *Backup) RestoreLatestBackupIfRelevant() error {
	if s.restoreFromBackupPath == "" {
		return nil
	}
	if !dirExists(s.restoreFromBackupPath) {
		log.Printf("backup: restore from backup not found, skipping restore: %s", s.restoreFromBackupPath)
		return nil
	}

	// A valid extracted backup should contain expected top-level dirs.
	baseSrc := filepath.Join(s.restoreFromBackupPath, "base_data")
	modelsSrc := filepath.Join(s.restoreFromBackupPath, "models")
	itemsSrc := filepath.Join(s.restoreFromBackupPath, "items_metadata")
	dbSrcDir := filepath.Join(s.restoreFromBackupPath, "db")

	if !dirExists(baseSrc) || !dirExists(modelsSrc) || !dirExists(itemsSrc) || !dirExists(dbSrcDir) {
		return fmt.Errorf("restore: invalid restore directory structure at %s", s.restoreFromBackupPath)
	}

	// DB: take the first file found in db/ (should be exactly one).
	dbFile, err := firstRegularFile(dbSrcDir)
	if err != nil {
		return fmt.Errorf("restore: db file: %w", err)
	}

	log.Printf("restore: restoring base data from %s to %s", baseSrc, s.baseDataDir)
	if err := replaceDir(s.baseDataDir, baseSrc); err != nil {
		return fmt.Errorf("restore: base data: %w", err)
	}

	log.Printf("restore: restoring models from %s to %s", modelsSrc, s.modelsDir)
	if err := replaceDir(s.modelsDir, modelsSrc); err != nil {
		return fmt.Errorf("restore: models: %w", err)
	}

	log.Printf("restore: restoring items metadata from %s to %s", itemsSrc, s.itemsMetadataDir)
	if err := replaceDir(s.itemsMetadataDir, itemsSrc); err != nil {
		return fmt.Errorf("restore: items metadata: %w", err)
	}

	log.Printf("restore: restoring db from %s to %s", dbFile, s.dbPath)
	if err := replaceFile(s.dbPath, dbFile); err != nil {
		return fmt.Errorf("restore: db: %w", err)
	}

	// Cleanup restore payload.
	if err := os.RemoveAll(s.restoreFromBackupPath); err != nil {
		log.Printf("restore: warning: failed cleaning restore dir %s: %v", s.restoreFromBackupPath, err)
	} else {
		log.Printf("restore: cleaned restore dir %s", s.restoreFromBackupPath)
	}

	log.Printf("restore: completed")
	return nil
}

func (s *Backup) ensureBackupDir() error {
	if s.backupdir == "" {
		return errors.New("backup: backupdir is not set (empty string, check configuration)")
	}
	if err := os.MkdirAll(s.backupdir, 0o755); err != nil {
		return fmt.Errorf("backup: create backupdir: %w", err)
	}
	return nil
}

func (s *Backup) CreateBackupFromZip(file multipart.File, f func(dstPath string) error) (string, error) {
	dst, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		return "", fmt.Errorf("create backup from zip: create temp file: %w", err)
	}
	defer func() { _ = dst.Close() }()
	defer os.Remove(dst.Name())

	_, err = io.Copy(dst, file)
	if err != nil {
		return "", fmt.Errorf("create backup from zip: save file: %w", err)
	}

	if err := f(dst.Name()); err != nil {
		return "", fmt.Errorf("create backup from zip: process zip: %w", err)
	}

	// validations can be added here if needed, e.g. check for expected files/dirs in the extracted content.

	if err := futils.CopyFile(dst.Name(), filepath.Join(s.backupdir, fmt.Sprintf("backup_%s.zip", time.Now().UTC().Format("20060102T150405Z")))); err != nil {
		return "", fmt.Errorf("create backup from zip: save backup: %w", err)
	}

	if err := s.ensureMaxBackups(); err != nil {
		log.Printf("warning: failed to ensure max backups after creating backup from zip: %v", err)
	}

	return "backup created from zip", nil
}

func (s *Backup) GetBackupFullPath(backupId string) (string, error) {
	l, err := s.ListBackups()
	if err != nil {
		return "", err
	}
	if len(l) == 0 {
		return "", errors.New("no backups found")
	}
	if backupId == "latest" {
		backupId = l[0]
	}
	if !lo.Contains(l, backupId) {
		return "", fmt.Errorf("backup not found: %s", backupId)
	}
	return filepath.Join(s.backupdir, backupId+".zip"), nil
}

func (s *Backup) ensureMaxBackups() error {
	l, err := s.ListBackups()
	if err != nil {
		return fmt.Errorf("list backups: %w", err)
	}
	if len(l) > maxBackupsToStore {
		toDelete := l[maxBackupsToStore:]
		for _, b := range toDelete {
			p := filepath.Join(s.backupdir, b+".zip")
			if err := os.Remove(p); err != nil {
				log.Printf("warning: failed to remove old backup %s: %v", p, err)
			} else {
				log.Printf("removed old backup %s", p)
			}
		}
	}
	return nil
}

// todo: the following should be in futils, some duplicate code...

func addDirToZip(zw *zip.Writer, srcDir, zipPrefix string) error {
	if srcDir == "" {
		return fmt.Errorf("zip: srcDir empty for %s", zipPrefix)
	}
	info, err := os.Stat(srcDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Treat missing dirs as empty.
			return nil
		}
		return fmt.Errorf("zip: stat dir %s: %w", srcDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("zip: expected dir, got file: %s", srcDir)
	}

	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("zip: walk %s: %w", srcDir, walkErr)
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("zip: rel path: %w", err)
		}
		rel = filepath.ToSlash(rel)

		if d.IsDir() {
			// Explicit directory entries are optional, but harmless.
			if rel == "." {
				return nil
			}
			h := &zip.FileHeader{
				Name:   filepath.ToSlash(filepath.Join(zipPrefix, rel)) + "/",
				Method: zip.Deflate,
			}
			h.SetMode(0o755)
			_, err := zw.CreateHeader(h)
			return err
		}

		return addFileToZip(zw, path, filepath.ToSlash(filepath.Join(zipPrefix, rel)))
	})
}

func addFileToZip(zw *zip.Writer, srcPath, zipPath string) error {
	if srcPath == "" {
		return errors.New("zip: srcPath empty")
	}
	info, err := os.Stat(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Treat missing file as absent.
			return nil
		}
		return fmt.Errorf("zip: stat file %s: %w", srcPath, err)
	}
	if info.IsDir() {
		return fmt.Errorf("zip: expected file, got dir: %s", srcPath)
	}

	h, err := zip.FileInfoHeader(info)
	if err != nil {
		return fmt.Errorf("zip: header %s: %w", srcPath, err)
	}
	h.Name = filepath.ToSlash(zipPath)
	h.Method = zip.Deflate

	w, err := zw.CreateHeader(h)
	if err != nil {
		return fmt.Errorf("zip: create %s: %w", zipPath, err)
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("zip: open %s: %w", srcPath, err)
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("zip: copy %s: %w", srcPath, err)
	}
	return nil
}

func unzip(zipPath, dstDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("unzip: open %s: %w", zipPath, err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		name := f.Name
		if strings.Contains(name, "..") {
			return fmt.Errorf("unzip: suspicious path %q", name)
		}

		outPath := filepath.Join(dstDir, filepath.FromSlash(name))
		if !strings.HasPrefix(outPath, filepath.Clean(dstDir)+string(os.PathSeparator)) {
			return fmt.Errorf("unzip: path traversal blocked: %q", name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, 0o755); err != nil {
				return fmt.Errorf("unzip: mkdir %s: %w", outPath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("unzip: mkdir parent %s: %w", outPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("unzip: open entry %s: %w", name, err)
		}

		tmp := outPath + ".tmp"
		w, err := os.Create(tmp)
		if err != nil {
			_ = rc.Close()
			return fmt.Errorf("unzip: create %s: %w", tmp, err)
		}

		_, copyErr := io.Copy(w, rc)
		closeErr := w.Close()
		_ = rc.Close()

		if copyErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: write %s: %w", outPath, copyErr)
		}
		if closeErr != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: close %s: %w", outPath, closeErr)
		}

		if err := os.Rename(tmp, outPath); err != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("unzip: rename %s: %w", outPath, err)
		}
	}

	return nil
}

func replaceDir(dstDir, srcDir string) error {
	// Remove dst then copy src -> dst.
	if err := os.RemoveAll(dstDir); err != nil {
		return fmt.Errorf("replace dir: remove %s: %w", dstDir, err)
	}
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("replace dir: mkdir %s: %w", dstDir, err)
	}
	return copyDir(srcDir, dstDir)
}

func copyDir(srcDir, dstDir string) error {
	return filepath.WalkDir(srcDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("copy dir: walk %s: %w", srcDir, walkErr)
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return fmt.Errorf("copy dir: rel: %w", err)
		}
		if rel == "." {
			return nil
		}

		dstPath := filepath.Join(dstDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0o755)
		}

		return copyFile(path, dstPath)
	})
}

func replaceFile(dstFile, srcFile string) error {
	if err := os.MkdirAll(filepath.Dir(dstFile), 0o755); err != nil {
		return fmt.Errorf("replace file: mkdir parent: %w", err)
	}
	tmp := dstFile + ".tmp"
	if err := copyFile(srcFile, tmp); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dstFile); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace file: rename: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("copy file: open src %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()

	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("copy file: stat src %s: %w", src, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("copy file: create dst %s: %w", dst, err)
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()

	if copyErr != nil {
		return fmt.Errorf("copy file: copy to %s: %w", dst, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("copy file: close dst %s: %w", dst, closeErr)
	}

	// Best-effort permissions.
	_ = os.Chmod(dst, info.Mode())
	return nil
}

func firstRegularFile(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		p := filepath.Join(dir, e.Name())
		fi, err := os.Stat(p)
		if err != nil {
			continue
		}
		if fi.Mode().IsRegular() {
			return p, nil
		}
	}
	return "", fmt.Errorf("no regular file found in %s", dir)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}
