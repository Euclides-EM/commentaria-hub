package service

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	diagramcropmetadata "github.com/Euclides-EM/commentaria-hub/ocrflow/internal/diagramcrops"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Facsimile struct {
	facsimileStore           *store.FacsimileSQL
	facsimilesPDFDir         string
	remoteAPIURL             string
	remoteAuthToken          string
	driveInbox               *RcloneDrive
	facsimilesGDriveFolderID string
	facsimilesDiagramsPath   string
	diagramsStoreDir         string
	itemsMetadataStoreDir    string
	editionKeys              func() ([]string, error)
	shelfmarks               func() ([]*model.EditionShelfmark, error)
}

type facsimileMappingUpdate struct {
	row       int
	facsimile *model.Facsimile
	shelfmark string
	status    model.FacsimileConnectionConfirmationStatus
}

func NewFacsimileService(facsimileStore *store.FacsimileSQL, editionKeys func() ([]string, error), shelfmarks func() ([]*model.EditionShelfmark, error), facsimilesPDFDir, facsimilesDiagramsPath, diagramsStoreDir, itemsMetadataStoreDir, remoteAPIURL, remoteAuthToken, rcloneRemoteName, facsimilesGDriveFolderID string) *Facsimile {
	return &Facsimile{
		facsimileStore:           facsimileStore,
		facsimilesPDFDir:         strings.TrimSpace(facsimilesPDFDir),
		remoteAPIURL:             strings.TrimRight(strings.TrimSpace(remoteAPIURL), "/"),
		remoteAuthToken:          strings.TrimSpace(remoteAuthToken),
		driveInbox:               NewRcloneDrive(rcloneRemoteName, facsimilesGDriveFolderID, "facsimile inbox rclone"),
		facsimilesGDriveFolderID: strings.TrimSpace(facsimilesGDriveFolderID),
		facsimilesDiagramsPath:   strings.TrimSpace(facsimilesDiagramsPath),
		diagramsStoreDir:         strings.TrimSpace(diagramsStoreDir),
		itemsMetadataStoreDir:    strings.TrimSpace(itemsMetadataStoreDir),
		editionKeys:              editionKeys,
		shelfmarks:               shelfmarks,
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
	if err := e.validateFacsimileShelfmarkMapping(f); err != nil {
		return nil, err
	}
	f.ID = idgen.GenerateID(store.FacsimileIDPrefix)
	f.CreatedAt = time.Now()
	f.UpdatedAt = f.CreatedAt
	created, err := e.facsimileStore.InsertFacsimile(f)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (e *Facsimile) UpdateFacsimile(f *model.Facsimile) (*model.Facsimile, error) {
	existing, err := e.facsimileStore.GetFacsimileByID(f.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get facsimile: %w", err)
	}
	if existing == nil {
		return nil, fmt.Errorf("facsimile with id %s not found", f.ID)
	}
	if err := e.validateFacsimileShelfmarkMapping(f); err != nil {
		return nil, err
	}
	updated, err := e.facsimileStore.UpdateFacsimile(f)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (e *Facsimile) validateFacsimileShelfmarkMapping(f *model.Facsimile) error {
	if f == nil {
		return fmt.Errorf("facsimile is nil")
	}
	f.EditionID = strings.TrimSpace(f.EditionID)
	f.ShelfmarkID = strings.TrimSpace(f.ShelfmarkID)
	f.FacsimileConnectionConfirmationStatus = model.FacsimileConnectionConfirmationStatus(strings.TrimSpace(string(f.FacsimileConnectionConfirmationStatus)))

	if f.EditionID == "" {
		if f.ShelfmarkID != "" || f.FacsimileConnectionConfirmationStatus != "" {
			return fmt.Errorf("standalone facsimiles cannot have shelfmark_id or facsimile_connection_confirmation_status")
		}
		return nil
	}
	if f.ShelfmarkID == "" {
		if f.FacsimileConnectionConfirmationStatus != "" {
			return fmt.Errorf("facsimile_connection_confirmation_status must be blank when shelfmark_id is blank")
		}
		return nil
	}
	if !validFacsimileConnectionStatus(f.FacsimileConnectionConfirmationStatus) {
		return fmt.Errorf("invalid facsimile_connection_confirmation_status %q", f.FacsimileConnectionConfirmationStatus)
	}
	if _, err := e.validateShelfmarkBelongsToEdition(f.ShelfmarkID, f.EditionID); err != nil {
		return err
	}
	facsimiles, err := e.ListFacsimiles([]string{f.EditionID})
	if err != nil {
		return fmt.Errorf("list facsimiles for edition %q: %w", f.EditionID, err)
	}
	for _, existing := range facsimiles {
		if existing == nil || existing.ID == f.ID || strings.TrimSpace(existing.ShelfmarkID) != f.ShelfmarkID {
			continue
		}
		return fmt.Errorf("shelfmark_id %q is already connected to facsimile %q", f.ShelfmarkID, existing.ID)
	}
	return nil
}

func (e *Facsimile) validateShelfmarkBelongsToEdition(shelfmarkID, editionID string) (*model.EditionShelfmark, error) {
	if e.shelfmarks == nil {
		return nil, fmt.Errorf("cannot validate shelfmark_id %q: shelfmark service is not configured", shelfmarkID)
	}
	shelfmarks, err := e.shelfmarks()
	if err != nil {
		return nil, fmt.Errorf("list shelfmarks: %w", err)
	}
	for _, shelfmark := range shelfmarks {
		if shelfmark == nil || shelfmark.ID != shelfmarkID {
			continue
		}
		if shelfmark.EditionID != editionID {
			return nil, fmt.Errorf("shelfmark_id %q belongs to edition %q, not %q", shelfmarkID, shelfmark.EditionID, editionID)
		}
		return shelfmark, nil
	}
	return nil, fmt.Errorf("shelfmark_id %q not found", shelfmarkID)
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
	existingByScanURL := lo.SliceToMap(existingFacsimiles, func(f *model.Facsimile) (string, *model.Facsimile) {
		return f.ScanURL, f
	})
	existingByPDFName := make(map[string]*model.Facsimile, len(existingFacsimiles))
	for _, fac := range existingFacsimiles {
		if path, err := futils.URLToLocalFilePath(fac.ScanURL); err == nil {
			existingByPDFName[filepath.Base(path)] = fac
		}
	}
	editionKeys, err := e.localFacsimileEditionKeys()
	if err != nil {
		return err
	}

	added := 0
	updated := 0
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".pdf" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			log.Printf("failed to stat local facsimile %s: %v", entry.Name(), err)
			continue
		}
		fileKey := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		editionID, ok := facsimileEditionID(fileKey, editionKeys)
		if !ok {
			log.Printf("skipping local facsimile %s: no known edition key matches its filename", entry.Name())
			continue
		}
		pdfURL, err := futils.LocalFilePathToURL(filepath.Join(dir, entry.Name()))
		if err != nil {
			log.Printf("failed to build local file URL for %s: %v", entry.Name(), err)
			continue
		}

		existing := existingByScanURL[pdfURL]
		if existing == nil {
			existing = existingByPDFName[entry.Name()]
		}
		if existing != nil {
			if existing.EditionID == editionID && existing.ScanURL == pdfURL && existing.FileSizeBytes == info.Size() && existing.ImportedAt != nil {
				continue
			}
			existing.EditionID = editionID
			existing.ScanURL = pdfURL
			existing.FileSizeBytes = info.Size()
			if existing.ImportedAt == nil {
				importedAt := time.Now()
				existing.ImportedAt = &importedAt
			}
			if _, err := e.UpdateFacsimile(existing); err != nil {
				log.Printf("failed to update local facsimile for %s: %v", editionID, err)
				continue
			}
			updated++
			continue
		}

		facsimile := &model.Facsimile{
			EditionID: editionID,
			ScanURL:   pdfURL,
		}
		facsimile.Name = fileKey
		facsimile.FileSizeBytes = info.Size()
		importedAt := time.Now()
		facsimile.ImportedAt = &importedAt
		if _, err := e.CreateFacsimile(facsimile); err != nil {
			log.Printf("failed to create local facsimile for %s: %v", editionID, err)
			continue
		}
		added++
	}

	log.Printf("local facsimile sync from %s added %d and updated %d facsimiles", dir, added, updated)
	return nil
}

func (e *Facsimile) localFacsimileEditionKeys() ([]string, error) {
	if e.editionKeys == nil {
		return nil, nil
	}
	keys, err := e.editionKeys()
	if err != nil {
		return nil, fmt.Errorf("load edition keys for local facsimile sync: %w", err)
	}
	slices.SortFunc(keys, func(a, b string) int {
		return len(b) - len(a)
	})
	return keys, nil
}

func facsimileEditionID(fileKey string, editionKeys []string) (string, bool) {
	if len(editionKeys) == 0 {
		return fileKey, true
	}
	keys := append([]string{}, editionKeys...)
	slices.SortFunc(keys, func(a, b string) int {
		return len(b) - len(a)
	})
	for _, key := range keys {
		if fileKey == key || strings.HasPrefix(fileKey, key+"_") {
			return key, true
		}
	}
	return "", false
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
	existingByScanURL := lo.SliceToMap(existingFacsimiles, func(f *model.Facsimile) (string, *model.Facsimile) {
		return f.ScanURL, f
	})

	added := 0
	updated := 0
	for _, remote := range remoteFacsimiles {
		if remote.EditionID == "" {
			continue
		}
		if remote.ID == "" {
			log.Printf("skipping remote facsimile for %s: missing facsimile ID", remote.EditionID)
			continue
		}
		downloadURL := e.remoteFacsimilePDFURL(remote.ID)
		if existing := existingByScanURL[downloadURL]; existing != nil {
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

func (e *Facsimile) remoteFacsimilePDFURL(facsimileID string) string {
	return e.remoteAPIURL + "/facsimilies/" + url.PathEscape(facsimileID) + "/pdf"
}

type driveFileEntry struct {
	Path    string `json:"Path"`
	Name    string `json:"Name"`
	ModTime string `json:"ModTime"`
	IsDir   bool   `json:"IsDir"`
}

func (e *Facsimile) ImportFromDriveInbox() (*model.FacsimileDriveImportResult, error) {
	if e.driveInbox.RemoteName == "" {
		return nil, fmt.Errorf("the rclone remote name for the facsimiles Google Drive inbox is not configured")
	}
	if e.facsimilesGDriveFolderID == "" {
		return nil, fmt.Errorf("the facsimiles GDrive folder ID is not configured")
	}

	entries, err := e.listDriveInboxFiles()
	if err != nil {
		return nil, err
	}
	log.Printf("found %d new files in Google Drive inbox", len(entries))
	result := &model.FacsimileDriveImportResult{}
	for _, entry := range entries {
		name := filepath.Base(entry.Path)
		if name != entry.Path || strings.HasPrefix(name, ".") {
			result.Skipped = append(result.Skipped, entry.Path)
			continue
		}

		switch {
		case strings.EqualFold(filepath.Ext(name), ".pdf"):
			if err := e.importDrivePDF(entry); err != nil {
				return nil, err
			}
			result.ImportedPDFs = append(result.ImportedPDFs, name)
		case futils.IsArchive(name):
			imported, err := e.importDriveDiagramCrops(entry)
			if err != nil {
				return nil, err
			}
			result.ImportedDiagramArchives = append(result.ImportedDiagramArchives, name)
			result.ImportedDiagramCrops = append(result.ImportedDiagramCrops, imported...)
		default:
			result.Skipped = append(result.Skipped, entry.Path)
		}
	}

	log.Printf("importing %d PDFs and %d diagram crop directories from Google Drive inbox and skipping %d files\nPDFs: %v\nDiagram crops: %v\nSkipping: %v", len(result.ImportedPDFs), len(result.ImportedDiagramCrops), len(result.Skipped), result.ImportedPDFs, result.ImportedDiagramCrops, result.Skipped)
	if len(result.ImportedPDFs) > 0 {
		if err := e.UpdateFromLocalDir(e.facsimilesPDFDir); err != nil {
			return nil, fmt.Errorf("update facsimiles from imported PDFs: %w", err)
		}
	}
	if len(result.ImportedDiagramCrops) > 0 {
		if err := e.regenerateDiagramCropMetadata(); err != nil {
			return nil, fmt.Errorf("regenerate diagram crop metadata: %w", err)
		}
	}

	log.Printf("finished importing facsimiles from Google Drive inbox, deleting imported files from Drive")
	importedDriveFiles := append([]string{}, result.ImportedPDFs...)
	importedDriveFiles = append(importedDriveFiles, result.ImportedDiagramArchives...)
	for _, name := range importedDriveFiles {
		if _, err := e.driveInbox.Run("deletefile", e.driveInbox.RemotePath(name)); err != nil {
			return nil, fmt.Errorf("delete imported drive facsimile %s: %w", name, err)
		}
		result.Deleted = append(result.Deleted, name)
	}

	log.Printf("deleted %d files from Google Drive inbox", len(result.Deleted))
	return result, nil
}

func (e *Facsimile) ExportMappingCSVZip() (string, error) {
	facsimiles, err := e.ListFacsimiles(nil)
	if err != nil {
		return "", err
	}
	shelfmarks, err := e.listShelfmarksForMapping()
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp("", "facsimile-mapping-*.zip")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	zw := zip.NewWriter(tmp)
	if err := writeZipCSV(zw, "facsimiles.csv", facsimileMappingRecords(facsimiles, shelfmarks)); err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := writeZipCSV(zw, "shelfmarks.csv", shelfmarkMappingRecords(shelfmarks)); err != nil {
		_ = zw.Close()
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := zw.Close(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func (e *Facsimile) ImportMappingCSV(r io.Reader) error {
	records, err := csv.NewReader(r).ReadAll()
	if err != nil {
		return fmt.Errorf("read facsimile mapping csv: %w", err)
	}
	if len(records) == 0 {
		return fmt.Errorf("facsimile mapping csv is empty")
	}
	header := records[0]
	shelfmarks, err := e.listShelfmarksForMapping()
	if err != nil {
		return err
	}
	shelfmarkByID := lo.SliceToMap(shelfmarks, func(sh *model.EditionShelfmark) (string, *model.EditionShelfmark) {
		return sh.ID, sh
	})
	updates := make([]facsimileMappingUpdate, 0, len(records)-1)
	targetFacsimileByShelfmark := map[string]string{}
	for i, record := range records[1:] {
		row := csvRecordToMap(header, record)
		facsimileID := strings.TrimSpace(row["facsimile_id"])
		if facsimileID == "" {
			return fmt.Errorf("row %d missing facsimile_id", i+2)
		}
		fac, err := e.GetFacsimile(facsimileID)
		if err != nil {
			return fmt.Errorf("row %d: %w", i+2, err)
		}
		editionID := strings.TrimSpace(row["edition_id"])
		if editionID != fac.EditionID {
			return fmt.Errorf("row %d: edition_id %q does not match facsimile edition_id %q", i+2, editionID, fac.EditionID)
		}
		shelfmarkID := strings.TrimSpace(row["shelfmark_id"])
		status := model.FacsimileConnectionConfirmationStatus(strings.TrimSpace(row["facsimile_connection_confirmation_status"]))
		if fac.EditionID == "" {
			if shelfmarkID != "" || status != "" {
				return fmt.Errorf("row %d: standalone facsimiles cannot have shelfmark_id or facsimile_connection_confirmation_status", i+2)
			}
		}
		if shelfmarkID == "" {
			if status != "" {
				return fmt.Errorf("row %d: facsimile_connection_confirmation_status must be blank when shelfmark_id is blank", i+2)
			}
		} else {
			sh := shelfmarkByID[shelfmarkID]
			if sh == nil {
				return fmt.Errorf("row %d: shelfmark_id %q not found", i+2, shelfmarkID)
			}
			if sh.EditionID != fac.EditionID {
				return fmt.Errorf("row %d: shelfmark_id %q belongs to edition %q, not %q", i+2, shelfmarkID, sh.EditionID, fac.EditionID)
			}
			if !validFacsimileConnectionStatus(status) {
				return fmt.Errorf("row %d: invalid facsimile_connection_confirmation_status %q", i+2, status)
			}
			if existingFacsimileID := targetFacsimileByShelfmark[shelfmarkID]; existingFacsimileID != "" && existingFacsimileID != facsimileID {
				return fmt.Errorf("row %d: shelfmark_id %q is also assigned to facsimile %q in this CSV", i+2, shelfmarkID, existingFacsimileID)
			}
			targetFacsimileByShelfmark[shelfmarkID] = facsimileID
		}
		updates = append(updates, facsimileMappingUpdate{
			row:       i + 2,
			facsimile: fac,
			shelfmark: shelfmarkID,
			status:    status,
		})
	}
	if err := e.validateMappingUpdatesDoNotReuseShelfmarks(updates); err != nil {
		return err
	}
	for _, update := range updates {
		if update.shelfmark != "" {
			continue
		}
		update.facsimile.ShelfmarkID = ""
		update.facsimile.FacsimileConnectionConfirmationStatus = ""
		if _, err := e.UpdateFacsimile(update.facsimile); err != nil {
			return fmt.Errorf("row %d: update facsimile %s: %w", update.row, update.facsimile.ID, err)
		}
	}
	for _, update := range updates {
		if update.shelfmark == "" {
			continue
		}
		update.facsimile.ShelfmarkID = update.shelfmark
		update.facsimile.FacsimileConnectionConfirmationStatus = update.status
		if _, err := e.UpdateFacsimile(update.facsimile); err != nil {
			return fmt.Errorf("row %d: update facsimile %s: %w", update.row, update.facsimile.ID, err)
		}
	}
	return nil
}

func (e *Facsimile) validateMappingUpdatesDoNotReuseShelfmarks(updates []facsimileMappingUpdate) error {
	targetByFacsimile := map[string]string{}
	updatedFacsimiles := map[string]struct{}{}
	for _, update := range updates {
		targetByFacsimile[update.facsimile.ID] = update.shelfmark
		updatedFacsimiles[update.facsimile.ID] = struct{}{}
	}
	current, err := e.ListFacsimiles(nil)
	if err != nil {
		return fmt.Errorf("list current facsimiles: %w", err)
	}
	assigned := map[string]string{}
	for _, facsimile := range current {
		if facsimile == nil {
			continue
		}
		shelfmarkID := strings.TrimSpace(facsimile.ShelfmarkID)
		if _, ok := updatedFacsimiles[facsimile.ID]; ok {
			shelfmarkID = targetByFacsimile[facsimile.ID]
		}
		if shelfmarkID == "" {
			continue
		}
		if existingFacsimileID := assigned[shelfmarkID]; existingFacsimileID != "" && existingFacsimileID != facsimile.ID {
			return fmt.Errorf("shelfmark_id %q would be connected to both facsimile %q and facsimile %q", shelfmarkID, existingFacsimileID, facsimile.ID)
		}
		assigned[shelfmarkID] = facsimile.ID
	}
	return nil
}

func (e *Facsimile) listShelfmarksForMapping() ([]*model.EditionShelfmark, error) {
	if e.shelfmarks == nil {
		return nil, nil
	}
	return e.shelfmarks()
}

func validFacsimileConnectionStatus(status model.FacsimileConnectionConfirmationStatus) bool {
	switch status {
	case model.FacsimileConnectionStatusGuessedByMachine, model.FacsimileConnectionStatusGuessedByHuman, model.FacsimileConnectionStatusHumanConfirmed:
		return true
	default:
		return false
	}
}

func writeZipCSV(zw *zip.Writer, name string, records [][]string) error {
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	if err := cw.WriteAll(records); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func facsimileMappingRecords(facsimiles []*model.Facsimile, shelfmarks []*model.EditionShelfmark) [][]string {
	header := []string{"edition_id", "facsimile_id", "facsimile_name", "scan_url", "scan_filename", "download_available", "diagram_crops_available", "file_size_bytes", "imported_at", "shelfmark_id", "facsimile_connection_confirmation_status", "note"}
	records := [][]string{header}
	shelfmarksByEdition := make(map[string][]*model.EditionShelfmark)
	for _, sh := range shelfmarks {
		if sh != nil {
			shelfmarksByEdition[sh.EditionID] = append(shelfmarksByEdition[sh.EditionID], sh)
		}
	}
	facsimilesByEdition := make(map[string][]*model.Facsimile)
	for _, fac := range facsimiles {
		if fac != nil && fac.EditionID != "" {
			facsimilesByEdition[fac.EditionID] = append(facsimilesByEdition[fac.EditionID], fac)
		}
	}
	for _, fac := range facsimiles {
		if fac == nil {
			continue
		}
		shelfmarkID := fac.ShelfmarkID
		status := string(fac.FacsimileConnectionConfirmationStatus)
		if shelfmarkID == "" && fac.EditionID != "" && len(shelfmarksByEdition[fac.EditionID]) == 1 && len(facsimilesByEdition[fac.EditionID]) == 1 {
			shelfmarkID = shelfmarksByEdition[fac.EditionID][0].ID
			status = string(model.FacsimileConnectionStatusGuessedByMachine)
		}
		records = append(records, []string{
			fac.EditionID,
			fac.ID,
			fac.Name,
			fac.ScanURL,
			facsimileMappingScanFilename(fac.ScanURL),
			strconv.FormatBool(fac.DownloadAvailable),
			strconv.FormatBool(fac.DiagramCropsAvailable),
			strconv.FormatInt(fac.FileSizeBytes, 10),
			formatMappingTime(fac.ImportedAt),
			shelfmarkID,
			status,
			"",
		})
	}
	return records
}

func facsimileMappingScanFilename(scanURL string) string {
	name := facsimileScanBasename(scanURL)
	if name == "." {
		return ""
	}
	return name
}

func shelfmarkMappingRecords(shelfmarks []*model.EditionShelfmark) [][]string {
	records := [][]string{{"edition_id", "shelfmark_id", "volume", "shelfmark", "scan", "copyright", "note"}}
	for _, sh := range shelfmarks {
		if sh == nil {
			continue
		}
		volume := ""
		if sh.Volume != nil {
			volume = strconv.Itoa(*sh.Volume)
		}
		records = append(records, []string{sh.EditionID, sh.ID, volume, sh.Shelfmark, sh.Scan, sh.Copyright, sh.Note})
	}
	return records
}

func formatMappingTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func csvRecordToMap(header, record []string) map[string]string {
	out := make(map[string]string, len(header))
	for i, name := range header {
		if i < len(record) {
			out[name] = record[i]
		}
	}
	return out
}

func (e *Facsimile) listDriveInboxFiles() ([]driveFileEntry, error) {
	out, err := e.driveInbox.Run("lsjson", "--files-only", e.driveInbox.RemotePath(""))
	if err != nil {
		return nil, err
	}
	var entries []driveFileEntry
	if err := json.Unmarshal(out, &entries); err != nil {
		return nil, fmt.Errorf("parse drive facsimile inbox listing: %w", err)
	}
	files := make([]driveFileEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir {
			continue
		}
		files = append(files, entry)
	}
	slices.SortFunc(files, func(a, b driveFileEntry) int {
		return strings.Compare(a.ModTime, b.ModTime)
	})
	return files, nil
}

func (e *Facsimile) importDrivePDF(entry driveFileEntry) error {
	if e.facsimilesPDFDir == "" {
		return fmt.Errorf("the facsimiles PDF directory is not configured")
	}
	if err := os.MkdirAll(e.facsimilesPDFDir, 0o755); err != nil {
		return fmt.Errorf("create facsimiles pdf dir: %w", err)
	}

	name := filepath.Base(entry.Path)
	dst := filepath.Join(e.facsimilesPDFDir, name)
	tmp := dst + ".tmp"
	if _, err := e.driveInbox.Run("copyto", e.driveInbox.RemotePath(entry.Path), tmp); err != nil {
		return fmt.Errorf("copy drive facsimile %s: %w", entry.Path, err)
	}
	if err := os.Chmod(tmp, 0o644); err != nil {
		return fmt.Errorf("chmod imported facsimile %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		return fmt.Errorf("move imported facsimile %s to %s: %w", tmp, dst, err)
	}
	return nil
}

func (e *Facsimile) importDriveDiagramCrops(entry driveFileEntry) ([]string, error) {
	diagramsDir, err := futils.LocalDirFromPathOrURL(e.facsimilesDiagramsPath)
	if err != nil {
		return nil, err
	}
	if diagramsDir == "" {
		return nil, fmt.Errorf("the facsimiles diagrams path must be an absolute local path or file:// URL")
	}
	if err := os.MkdirAll(diagramsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create facsimiles diagrams dir: %w", err)
	}

	tmpArchive, err := futils.CreateTemp("drive-diagram-crops-*" + futils.ArchiveExt(entry.Path))
	if err != nil {
		return nil, fmt.Errorf("create temp diagram crop archive: %w", err)
	}
	tmpArchivePath := tmpArchive.Name()
	if err := tmpArchive.Close(); err != nil {
		return nil, fmt.Errorf("close temp diagram crop archive: %w", err)
	}
	defer os.Remove(tmpArchivePath)

	if _, err := e.driveInbox.Run("copyto", e.driveInbox.RemotePath(entry.Path), tmpArchivePath); err != nil {
		return nil, fmt.Errorf("copy drive diagram crops %s: %w", entry.Path, err)
	}

	tmpDir, err := futils.MkdirTemp("drive-diagram-crops")
	if err != nil {
		return nil, fmt.Errorf("create temp diagram crop dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := futils.ExtractArchive(tmpArchivePath, tmpDir); err != nil {
		return nil, fmt.Errorf("extract diagram crop archive %s: %w", entry.Path, err)
	}

	cropDirs, err := findCropDirs(tmpDir)
	if err != nil {
		return nil, err
	}
	if len(cropDirs) == 0 {
		return nil, fmt.Errorf("diagram crop archive %s contains no <edition_key>/crops/*.jpg directories", entry.Path)
	}

	imported := make([]string, 0, len(cropDirs))
	for key, src := range cropDirs {
		dst := filepath.Join(diagramsDir, key)
		if err := os.RemoveAll(dst); err != nil {
			return nil, fmt.Errorf("remove existing diagram crops for %s: %w", key, err)
		}
		if err := futils.CopyDir(src, dst); err != nil {
			return nil, fmt.Errorf("install diagram crops for %s: %w", key, err)
		}
		if err := e.deleteDiagramCropMetadata(key); err != nil {
			return nil, err
		}
		imported = append(imported, key)
	}
	slices.Sort(imported)
	return imported, nil
}

func (e *Facsimile) regenerateDiagramCropMetadata() error {
	if e.diagramsStoreDir == "" || e.itemsMetadataStoreDir == "" {
		return fmt.Errorf("diagram metadata store directories are not configured")
	}
	return diagramcropmetadata.GenerateFromPaths(e.facsimilesDiagramsPath, e.diagramsStoreDir, e.itemsMetadataStoreDir, diagramcropmetadata.Options{Force: true})
}

func (e *Facsimile) deleteDiagramCropMetadata(key string) error {
	if e.diagramsStoreDir == "" {
		return nil
	}
	baseKey := diagramcropmetadata.BaseKey(key)
	if baseKey == "" {
		return nil
	}
	path := filepath.Join(e.diagramsStoreDir, baseKey+".json")
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete diagram crop metadata %s: %w", path, err)
	}
	return nil
}

func findCropDirs(root string) (map[string]string, error) {
	cropDirs := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() || d.Name() != "crops" {
			return nil
		}
		parent := filepath.Dir(path)
		key := filepath.Base(parent)
		if !diagramcropmetadata.ValidKey(key) {
			return fmt.Errorf("invalid diagram crop directory key: %s", key)
		}
		hasJPG, err := futils.DirIncludesByFileExt(path, "jpg")
		if err != nil {
			return err
		}
		if hasJPG {
			cropDirs[key] = parent
		}
		return filepath.SkipDir
	})
	return cropDirs, err
}

func (e *Facsimile) GetFacsimilePDFPath(facsimileID string) (string, error) {
	fac, err := e.GetFacsimile(facsimileID)
	if err != nil {
		return "", err
	}
	return facsimileLocalPDFPath(fac)
}

func downloadableFacsimilePDFPath(facs []*model.Facsimile) (string, error) {
	for _, fac := range facs {
		if path, err := facsimileLocalPDFPath(fac); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("no local facsimile PDF available")
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
