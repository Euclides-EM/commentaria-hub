package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Facsimile struct {
	facsimileStore           *store.FacsimileSQL
	facsimilesPDFDir         string
	remoteAPIURL             string
	remoteAuthToken          string
	driveInbox               *RcloneDrive
	facsimilesGDriveFolderID string
}

func NewFacsimileService(facsimileStore *store.FacsimileSQL, facsimilesPDFDir, remoteAPIURL, remoteAuthToken, rcloneRemoteName, facsimilesGDriveFolderID string) *Facsimile {
	return &Facsimile{
		facsimileStore:           facsimileStore,
		facsimilesPDFDir:         strings.TrimSpace(facsimilesPDFDir),
		remoteAPIURL:             strings.TrimRight(strings.TrimSpace(remoteAPIURL), "/"),
		remoteAuthToken:          strings.TrimSpace(remoteAuthToken),
		driveInbox:               NewRcloneDrive(rcloneRemoteName, facsimilesGDriveFolderID, "facsimile inbox rclone"),
		facsimilesGDriveFolderID: strings.TrimSpace(facsimilesGDriveFolderID),
	}
}

func (e *Facsimile) ListFacsimiles(editionIDs []string) ([]*model.Facsimile, error) {
	return e.facsimileStore.ListFacsimiles(editionIDs)
}

func (e *Facsimile) GetFacsimile(facsimileID string) (*model.Facsimile, error) {
	fac, err := e.facsimileStore.GetFacsimileByID(facsimileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile from store: %w", err)
	}
	if fac == nil {
		return nil, fmt.Errorf("facsimile with id %s not found", facsimileID)
	}
	return fac, nil
}

func (e *Facsimile) CreateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	f.ID = idgen.GenerateID(store.FacsimileIDPrefix)
	f.CreatedAt = time.Now()
	f.UpdatedAt = f.CreatedAt
	return e.facsimileStore.InsertFacsimile(f)
}

func (e *Facsimile) UpdateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	existing, err := e.facsimileStore.GetFacsimileByID(f.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("facsimile with id %s not found", f.ID)
	}
	return e.facsimileStore.UpdateFacsimile(f)
}

func (e *Facsimile) UpdateFromConfiguredSource() error {
	if e.facsimilesPDFDir != "" {
		return e.UpdateFromLocalDir(e.facsimilesPDFDir)
	}
	if e.remoteAPIURL != "" {
		return e.UpdateFromRemoteAPI()
	}
	log.Printf("facsimiles: no FACSIMILES_PDF_DIR or FACSIMILES_REMOTE_API_URL configured; skipping facsimile sync")
	return nil
}

func (e *Facsimile) UpdateFromLocalDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to list local facsimiles in %s: %w", dir, err)
	}

	existingFacsimiles, err := e.ListFacsimiles(nil)
	if err != nil {
		return fmt.Errorf("failed to list existing facsimiles: %w", err)
	}
	existingByEditionID := lo.SliceToMap(existingFacsimiles, func(f *model.Facsimile) (string, *model.Facsimile) {
		return f.EditionID, f
	})

	added := 0
	updated := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".pdf" {
			continue
		}
		editionID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		pdfURL, err := futils.LocalFilePathToURL(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Printf("failed to build local file URL for %s: %v", entry.Name(), err)
			continue
		}

		if existing := existingByEditionID[editionID]; existing != nil {
			if existing.ScanURL == pdfURL {
				continue
			}
			existing.ScanURL = pdfURL
			if _, err := e.UpdateFacsimile(existing); err != nil {
				log.Printf("failed to update local facsimile for %s: %v", editionID, err)
				continue
			}
			updated++
			continue
		}

		if _, err := e.CreateFacsimile(&model.Facsimile{
			EditionID: editionID,
			ScanURL:   pdfURL,
		}); err != nil {
			log.Printf("failed to create local facsimile for %s: %v", editionID, err)
			continue
		}
		added++
	}

	log.Printf("local facsimile sync from %s added %d and updated %d facsimiles", dir, added, updated)
	return nil
}

func (e *Facsimile) UpdateFromRemoteAPI() error {
	remoteFacsimiles, err := e.listRemoteFacsimiles()
	if err != nil {
		return err
	}
	existingFacsimiles, err := e.ListFacsimiles(nil)
	if err != nil {
		return fmt.Errorf("failed to list existing facsimiles: %w", err)
	}
	existingByEditionID := lo.SliceToMap(existingFacsimiles, func(f *model.Facsimile) (string, *model.Facsimile) {
		return f.EditionID, f
	})

	added := 0
	updated := 0
	for _, remote := range remoteFacsimiles {
		if remote.EditionID == "" {
			continue
		}
		downloadURL := e.remoteEditionPDFURL(remote.EditionID)
		if existing := existingByEditionID[remote.EditionID]; existing != nil {
			existing.ScanURL = downloadURL
			existing.MainTextPages = remote.MainTextPages
			if existing.Name == "" {
				existing.Name = remote.Name
			}
			if existing.Description == "" {
				existing.Description = remote.Description
			}
			if _, err := e.UpdateFacsimile(existing); err != nil {
				log.Printf("failed to update remote facsimile for %s: %v", remote.EditionID, err)
				continue
			}
			updated++
			continue
		}
		remote.ID = idgen.GenerateID(store.FacsimileIDPrefix)
		remote.ScanURL = downloadURL
		if _, err := e.CreateFacsimile(remote); err != nil {
			log.Printf("failed to create remote facsimile for %s: %v", remote.EditionID, err)
			continue
		}
		added++
	}
	log.Printf("remote facsimile sync from %s added %d and updated %d facsimiles", e.remoteAPIURL, added, updated)
	return nil
}

func (e *Facsimile) listRemoteFacsimiles() ([]*model.Facsimile, error) {
	req, err := http.NewRequest(http.MethodGet, e.remoteAPIURL+"/facsimilies", nil)
	if err != nil {
		return nil, err
	}
	if e.remoteAuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+e.remoteAuthToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("list remote facsimiles: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("list remote facsimiles: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var facsimiles []*model.Facsimile
	if err := json.NewDecoder(resp.Body).Decode(&facsimiles); err != nil {
		return nil, fmt.Errorf("decode remote facsimiles: %w", err)
	}
	return facsimiles, nil
}

func (e *Facsimile) remoteEditionPDFURL(editionID string) string {
	return e.remoteAPIURL + "/editions/" + url.PathEscape(editionID) + "/facsimile.pdf"
}

type driveFileEntry struct {
	Path    string `json:"Path"`
	Name    string `json:"Name"`
	ModTime string `json:"ModTime"`
	IsDir   bool   `json:"IsDir"`
}

func (e *Facsimile) ImportFromDriveInbox() (*model.FacsimileDriveImportResult, error) {
	if e.facsimilesPDFDir == "" {
		return nil, fmt.Errorf("the facsimiles PDF directory is not configured")
	}
	if e.driveInbox.RemoteName == "" {
		return nil, fmt.Errorf("the rclone remote name for the facsimiles Google Drive inbox is not configured")
	}
	if e.facsimilesGDriveFolderID == "" {
		return nil, fmt.Errorf("the facsimiles GDrive folder ID is not configured")
	}
	if err := os.MkdirAll(e.facsimilesPDFDir, 0o755); err != nil {
		return nil, fmt.Errorf("create facsimiles pdf dir: %w", err)
	}

	entries, err := e.listDriveInboxPDFs()
	if err != nil {
		return nil, err
	}
	log.Printf("found %d new files in Google Drive inbox", len(entries))
	result := &model.FacsimileDriveImportResult{}
	for _, entry := range entries {
		name := filepath.Base(entry.Path)
		if name != entry.Path || strings.HasPrefix(name, ".") || !strings.EqualFold(filepath.Ext(name), ".pdf") {
			result.Skipped = append(result.Skipped, entry.Path)
			continue
		}

		dst := filepath.Join(e.facsimilesPDFDir, name)
		tmp := dst + ".tmp"
		if _, err := e.driveInbox.Run("copyto", e.driveInbox.RemotePath(entry.Path), tmp); err != nil {
			return nil, fmt.Errorf("copy drive facsimile %s: %w", entry.Path, err)
		}
		if err := os.Chmod(tmp, 0o644); err != nil {
			return nil, fmt.Errorf("chmod imported facsimile %s: %w", tmp, err)
		}
		if err := os.Rename(tmp, dst); err != nil {
			return nil, fmt.Errorf("move imported facsimile %s to %s: %w", tmp, dst, err)
		}
		result.Imported = append(result.Imported, name)
	}

	log.Printf("importing %d new files from Google Drive inbox and skipping %d files\nImporting: %v\nskipping: %v", len(result.Imported), len(result.Skipped), result.Imported, result.Skipped)
	if len(result.Imported) > 0 {
		if err := e.UpdateFromLocalDir(e.facsimilesPDFDir); err != nil {
			return nil, fmt.Errorf("update facsimiles from imported PDFs: %w", err)
		}
	}

	log.Printf("finished importing facsimiles from Google Drive inbox, deleting imported files from Drive")
	for _, name := range result.Imported {
		if _, err := e.driveInbox.Run("deletefile", e.driveInbox.RemotePath(name)); err != nil {
			return nil, fmt.Errorf("delete imported drive facsimile %s: %w", name, err)
		}
		result.Deleted = append(result.Deleted, name)
	}

	log.Printf("deleted %d files from Google Drive inbox", len(result.Deleted))
	return result, nil
}

func (e *Facsimile) listDriveInboxPDFs() ([]driveFileEntry, error) {
	out, err := e.driveInbox.Run("lsjson", "--files-only", e.driveInbox.RemotePath(""))
	if err != nil {
		return nil, err
	}
	var entries []driveFileEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parse drive facsimile inbox listing: %w", err)
	}
	var pdfs []driveFileEntry
	for _, entry := range entries {
		if entry.IsDir || !strings.EqualFold(filepath.Ext(entry.Path), ".pdf") {
			continue
		}
		pdfs = append(pdfs, entry)
	}
	slices.SortFunc(pdfs, func(a, b driveFileEntry) int {
		return strings.Compare(a.ModTime, b.ModTime)
	})
	return pdfs, nil
}

func (e *Facsimile) GetFacsimilePDFPath(facsimileID string) (string, error) {
	fac, err := e.GetFacsimile(facsimileID)
	if err != nil {
		return "", err
	}
	return facsimileLocalPDFPath(fac)
}

func (e *Facsimile) GetEditionFacsimilePDFPath(editionID string) (string, error) {
	facs, err := e.ListFacsimiles([]string{editionID})
	if err != nil {
		return "", err
	}
	if len(facs) == 0 {
		return "", fmt.Errorf("no facsimile found for edition %s", editionID)
	}
	return facsimileLocalPDFPath(facs[0])
}

func facsimileLocalPDFPath(fac *model.Facsimile) (string, error) {
	if fac.ScanURL == "" {
		return "", fmt.Errorf("facsimile %s has no scan URL", fac.ID)
	}
	p, err := futils.URLToLocalFilePath(fac.ScanURL)
	if err != nil {
		return "", fmt.Errorf("facsimile %s scan URL is not a local file URL", fac.ID)
	}
	if strings.ToLower(filepath.Ext(p)) != ".pdf" {
		return "", fmt.Errorf("facsimile %s local path is not a PDF", fac.ID)
	}
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("facsimile PDF not available at %s: %w", p, err)
	}
	return p, nil
}
