package store

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/cache"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/samber/lo"
)

type EditionCSV struct {
	itemsMetadataDir string
	cacheStore       *cache.Cache
}

const (
	relItemsManuscript  = "items_manuscript.csv"
	relItemsPrint       = "items_print.csv"
	relMDManuscript     = "metadata_elements_manuscripts.csv"
	relMDPrint          = "metadata_elements_print.csv"
	relTranscriptions   = "paratext_transcriptions.csv"
	relShelfmarks       = "shelfmarks.csv"
	relTranslations     = "translations.csv"
	relCorpuses         = "corpuses.csv"
	relBibliography     = "bibliography.csv"
	relReviews          = "reviews.csv"
	relClusters         = "clusters.csv"
	relClusterItems     = "cluster_items.csv"
	relVisualElements   = "visual_elements.csv"
	relVisualElementsEx = "visual_elements_examples.csv"
	relLocators         = "locators.csv"
	relDiagramDirs      = "diagram-directories.json"
)

var (
	cacheWarmupError = fmt.Errorf("try again in a few moments when cache warmup is complete")
)

func NewEditionCSV(itemsMetadataDir string) *EditionCSV {
	return &EditionCSV{
		itemsMetadataDir: itemsMetadataDir,
		cacheStore:       cache.NewCache(),
	}
}

func (s *EditionCSV) WarmCache() error {
	return s.cacheStore.Warmup(func() (map[string]interface{}, error) {
		m, err := s.LoadAllEditions()
		if err != nil {
			return nil, err
		}
		return lo.MapValues(m, func(ed *model.Edition, _ string) any {
			return ed
		}), nil
	})
}

func (s *EditionCSV) csvPath(rel string) string {
	return filepath.Join(s.itemsMetadataDir, rel)
}

// UpdateNotes updates the notes field for the given key in items_print.csv.
func (s *EditionCSV) UpdateNotes(key, note string) error {
	if !s.cacheStore.IsWarm() {
		return cacheWarmupError
	}
	header, rows, err := csv.LoadCSVRecords(relItemsPrint)
	if err != nil {
		return err
	}
	found := false
	for i, r := range rows {
		if r["key"] == key {
			rows[i]["notes"] = note
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("item with key %s not found", key)
	}
	if err := csv.SaveCSVRecords(relItemsPrint, header, rows); err != nil {
		return fmt.Errorf("Error saving updated notes: %v\n", err)
	}
	loaded, err := s.loadEditionByKey(key)
	if err != nil {
		return fmt.Errorf("Error reloading edition after notes update: %v\n", err)
	}
	s.cacheStore.Set(key, loaded)
	return nil
}

func (s *EditionCSV) UpsertEdition(ed *model.Edition, user string) error {
	if !s.cacheStore.IsWarm() {
		return cacheWarmupError
	}
	if ed.IsManuscript {
		if err := s.upsertManuscript(ed); err != nil {
			return err
		}
	} else {
		if err := s.upsertPrint(ed); err != nil {
			return err
		}
	}
	if err := s.upsertShelfmarks(ed); err != nil {
		return err
	}
	if err := s.upsertTranslations(ed); err != nil {
		return err
	}
	if len(ed.Corpus) > 0 {
		if err := csv.UpsertRow(s.csvPath(relCorpuses), "key", ed.Key, map[string]string{
			"key":   ed.Key,
			"study": strings.Join(ed.Corpus, ", "),
		}); err != nil {
			return err
		}
	}
	if err := s.upsertBibliography(ed); err != nil {
		return err
	}
	if err := s.upsertClusters(ed); err != nil {
		return err
	}
	if err := s.upsertVisualElements(ed); err != nil {
		return err
	}
	if ed.Verified {
		if err := csv.UpsertRow(relReviews, "key", ed.Key, map[string]string{
			"key":        ed.Key,
			"researcher": user,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}); err != nil {
			return fmt.Errorf("Error upserting review: %v\n", err)
		}
	}
	loaded, err := s.loadEditionByKey(ed.Key)
	if err != nil {
		return fmt.Errorf("Error reloading edition after notes update: %v\n", err)
	}
	s.cacheStore.Set(ed.Key, loaded)
	return nil
}

func (s *EditionCSV) upsertManuscript(ed *model.Edition) error {
	row := map[string]string{
		"key":                ed.Key,
		"short_title":        ed.ShortTitle,
		"short_title_source": ed.ShortTitleSource,
		"year_from":          formatcov.IntPtrToStr(ed.ManuscriptYearFrom),
		"year_to":            formatcov.IntPtrToStr(ed.ManuscriptYearTo),
		"notes":              ed.Notes,
		"has_diagrams":       formatcov.BoolPtrToStr(ed.HasDiagrams),
	}
	if err := csv.UpsertRow(s.csvPath(relItemsManuscript), "key", ed.Key, row); err != nil {
		return fmt.Errorf("Error upserting manuscript item: %v\n", err)
	}
	if ed.IsElements {
		sub := ""
		if ed.ManuscriptSubclass != nil {
			sub = *ed.ManuscriptSubclass
		}
		if err := csv.UpsertRow(relMDManuscript, "key", ed.Key, map[string]string{
			"key":            ed.Key,
			"class":          ed.ManuscriptClass,
			"subclass":       sub,
			"elements_books": formatcov.IntsToCompressedStr(ed.Books),
		}); err != nil {
			return fmt.Errorf("Error upserting manuscript metadata: %v\n", err)
		}
	}
	return nil
}

func (s *EditionCSV) upsertPrint(ed *model.Edition) error {
	langs := make([]string, len(ed.Languages))
	for i, l := range ed.Languages {
		langs[i] = strings.ToUpper(l)
	}
	row := map[string]string{
		"key":                ed.Key,
		"city":               strings.Join(ed.Cities, ", "),
		"short_title":        ed.ShortTitle,
		"short_title_source": ed.ShortTitleSource,
		"year":               formatcov.PtrToStr(ed.Year),
		"language":           strings.Join(langs, ", "),
		"author_or_editor":   strings.Join(ed.Editor, ", "),
		"publisher":          strings.Join(ed.Publisher, ", "),
		"format":             formatcov.IntPtrToStr(ed.Format),
		"volumes":            formatcov.IntPtrToStr(ed.Volumes),
		"ustc_id":            formatcov.PtrToStr(ed.USTCId),
		"notes":              ed.Notes,
		"has_diagrams":       formatcov.BoolPtrToStr(ed.HasDiagrams),
	}
	if err := csv.UpsertRow(relItemsPrint, "key", ed.Key, row); err != nil {
		return fmt.Errorf("Error upserting print item: %v\n", err)
	}
	if ed.IsElements {
		if err := csv.UpsertRow(relMDPrint, "key", ed.Key, map[string]string{
			"key":                      ed.Key,
			"elements_books":           formatcov.IntsToCompressedStr(ed.Books),
			"additional_content":       strings.Join(ed.AdditionalContent, ", "),
			"wardhaugh_classification": "",
		}); err != nil {
			return fmt.Errorf("Error upserting print metadata: %v\n", err)
		}
	}
	if err := csv.UpsertRow(relTranscriptions, "key", ed.Key, map[string]string{
		"key":          ed.Key,
		"colophon":     formatcov.PtrToStr(ed.Colophon),
		"frontispiece": formatcov.PtrToStr(ed.Frontispiece),
		"imprint":      formatcov.PtrToStr(ed.Imprint),
		"title":        formatcov.PtrToStr(ed.Title),
	}); err != nil {
		return fmt.Errorf("Error upserting transcriptions: %v\n", err)
	}
	return nil
}

func (s *EditionCSV) upsertShelfmarks(ed *model.Edition) error {
	var rows []map[string]string
	for _, sh := range ed.Shelfmarks {
		vol := ""
		if sh.Volume != nil {
			vol = strconv.Itoa(*sh.Volume)
		}
		rows = append(rows, map[string]string{
			"key":              ed.Key,
			"volume":           vol,
			"scan":             sh.Scan,
			"title_page_img":   sh.TitlePageImg,
			"frontispiece_img": sh.FrontispieceImg,
			"annotations":      sh.Annotations,
			"shelf_mark":       sh.Shelfmark,
			"copyright":        sh.Copyright,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	if err := csv.ReplaceRowsForKey(relShelfmarks, "key", ed.Key, rows); err != nil {
		return fmt.Errorf("Error upserting shelfmarks: %v\n", err)
	}
	return nil
}

func (s *EditionCSV) upsertTranslations(ed *model.Edition) error {
	if ed.IsManuscript {
		return nil
	}
	var rows []map[string]string
	for _, pair := range []struct{ field, value *string }{
		{formatcov.StrToPtr("title"), ed.TitleEN},
		{formatcov.StrToPtr("imprint"), ed.ImprintEN},
		{formatcov.StrToPtr("colophon"), ed.ColophonEN},
		{formatcov.StrToPtr("frontispiece"), ed.FrontispieceEN},
	} {
		if pair.value != nil && *pair.value != "" {
			rows = append(rows, map[string]string{
				"key":    ed.Key,
				"field":  *pair.field,
				"en":     *pair.value,
				"source": "",
			})
		}
	}
	if len(rows) == 0 {
		return nil
	}
	if err := csv.BatchUpsertRows(relTranslations, "key", "field", rows); err != nil {
		return fmt.Errorf("Error upserting translations: %v\n", err)
	}
	return nil
}

func (s *EditionCSV) upsertBibliography(ed *model.Edition) error {
	var rows []map[string]string
	for _, c := range ed.Bibliography {
		rows = append(rows, map[string]string{"key": ed.Key, "citation": c})
	}
	if len(rows) == 0 {
		return nil
	}
	if err := csv.ReplaceRowsForKey(relBibliography, "key", ed.Key, rows); err != nil {
		fmt.Printf("Error upserting bibliography: %v\n", err)
	}
	return nil
}

func (s *EditionCSV) upsertClusters(ed *model.Edition) error {
	if ed.ReprintOf == nil || *ed.ReprintOf == "" {
		return nil
	}
	parentKey := *ed.ReprintOf
	_, clusterItems, _ := csv.LoadCSVRecords(relClusterItems)
	var parentClusterKey string
	for _, ci := range clusterItems {
		if ci["item_key"] == parentKey {
			parentClusterKey = ci["cluster_key"]
			break
		}
	}
	if parentClusterKey != "" {
		// add current to existing cluster
		found := false
		for _, ci := range clusterItems {
			if ci["item_key"] == ed.Key {
				found = true
				break
			}
		}
		if !found {
			composite := parentClusterKey + "_" + ed.Key
			if err := csv.UpsertRow(relClusterItems, "key", composite, map[string]string{
				"key":         composite,
				"cluster_key": parentClusterKey,
				"item_key":    ed.Key,
			}); err != nil {
				return fmt.Errorf("Error upserting cluster: %v\n", err)
			}
		}
	}
	clusterKey := strings.ToUpper(fmt.Sprintf("%x", rand.Int63())[:6])
	if err := csv.UpsertRow(relClusters, "key", clusterKey, map[string]string{"key": clusterKey, "type": "reprint"}); err != nil {
		return fmt.Errorf("Error upserting cluster: %v\n", err)
	}
	if err := csv.UpsertRow(relClusterItems, "key", clusterKey+"_"+parentKey, map[string]string{
		"key":         clusterKey + "_" + parentKey,
		"cluster_key": clusterKey,
		"item_key":    parentKey,
	}); err != nil {
		return fmt.Errorf("Error upserting cluster item: %v\n", err)
	}
	if err := csv.UpsertRow(relClusterItems, "key", clusterKey+"_"+ed.Key, map[string]string{
		"key":         clusterKey + "_" + ed.Key,
		"cluster_key": clusterKey,
		"item_key":    ed.Key,
	}); err != nil {
		return fmt.Errorf("Error upserting cluster item: %v\n", err)
	}
	return nil

}

func (s *EditionCSV) upsertVisualElements(ed *model.Edition) error {
	if len(ed.VisualElements) == 0 {
		return nil
	}
	// Upsert locators
	for _, ve := range ed.VisualElements {
		if ve.Locator != nil {
			if err := csv.UpsertRow(relLocators, "key", ve.Locator.Key, locatorRow(ve.Locator)); err != nil {
				return fmt.Errorf("Error upserting locator: %v\n", err)
			}
		}
		for _, ex := range ve.Examples {
			if ex.Locator != nil {
				if err := csv.UpsertRow(relLocators, "key", ex.Locator.Key, locatorRow(ex.Locator)); err != nil {
					return fmt.Errorf("Error upserting locator: %v\n", err)
				}
			}
		}
	}
	var veRows []map[string]string
	var exRows []map[string]string
	for _, ve := range ed.VisualElements {
		locKey := ""
		if ve.Locator != nil {
			locKey = ve.Locator.Key
		}
		veRows = append(veRows, map[string]string{
			"key":                 ed.Key,
			"visual_element_type": ve.VisualElementType,
			"locator_type":        ve.LocatorType,
			"locator_key":         locKey,
			"notes":               ve.Notes,
		})
		for _, ex := range ve.Examples {
			exLocKey := ""
			if ex.Locator != nil {
				exLocKey = ex.Locator.Key
			}
			exRows = append(exRows, map[string]string{
				"key":         ed.Key,
				"path":        ex.Img,
				"locator_key": exLocKey,
			})
		}
	}
	if err := csv.ReplaceRowsForKey(relVisualElements, "key", ed.Key, veRows); err != nil {
		return fmt.Errorf("Error upserting visual elements: %v\n", err)
	}
	// Replace examples for this edition's visual elements (by key = edition key in examples)
	if err := csv.ReplaceRowsForKey(relVisualElementsEx, "key", ed.Key, exRows); err != nil {
		return fmt.Errorf("Error upserting visual element examples: %v\n", err)
	}
	return nil
}

func (s *EditionCSV) DeleteEdition(key string) error {
	if !s.cacheStore.IsWarm() {
		return cacheWarmupError
	}
	for _, rel := range []string{
		relItemsManuscript, relItemsPrint, relMDManuscript, relMDPrint,
		relReviews, relShelfmarks, relTranscriptions, relTranslations,
		relCorpuses, relBibliography, relVisualElements,
		relVisualElementsEx, relLocators,
	} {
		if err := csv.DeleteRows(rel, "key", key); err != nil {
			return fmt.Errorf("Error deleting edition from %s: %v\n", rel, err)
		}
	}
	// cluster_items uses item_key for edition membership
	if err := csv.DeleteRows(relClusterItems, "item_key", key); err != nil {
		return fmt.Errorf("Error deleting edition from cluster items: %v\n", err)
	}
	s.cacheStore.Delete(key)
	return nil
}

func (s *EditionCSV) GetEditionByID(key string) (*model.Edition, error) {
	if !s.cacheStore.IsWarm() {
		return nil, cacheWarmupError
	}
	ed, ok := s.cacheStore.Get(key)
	if !ok {
		return nil, nil
	}
	if edTyped, ok := ed.(*model.Edition); ok {
		return edTyped, nil
	}
	return nil, fmt.Errorf("cache contains non-edition value of type %T for key %s", ed, key)
}

// ListEditions returns a page of editions and the total count. Pagination is performed
// in the store: only keys in [offset, offset+limit) are loaded in full; corpus filtering
// uses corpuses.csv only (no full edition load).
func (s *EditionCSV) ListEditions(filter func(e any) bool, orderBy func(e1, e2 any) int, offset, limit int) ([]*model.Edition, int, error) {
	if !s.cacheStore.IsWarm() {
		return nil, 0, cacheWarmupError
	}
	_, editions, total, err := s.cacheStore.GetBulk(filter,
		func(k1 string, k2 string, v1 any, v2 any) int {
			return orderBy(v1, v2)
		}, offset, limit)
	return lo.Map(editions, func(e any, _ int) *model.Edition {
		if ed, ok := e.(*model.Edition); ok {
			return ed
		}
		log.Printf("ListEditions: cache contains non-edition value of type %T", e)
		return nil
	}), total, err
}

// loadCSVRecordsOptional loads CSV rows from path; if file does not exist returns (nil, nil).
func (s *EditionCSV) loadCSVRecordsOptional(rel string) ([]map[string]string, error) {
	_, rows, err := csv.LoadCSVRecords(s.csvPath(rel))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return rows, nil
}

func findRowByKey(rows []map[string]string, keyField, key string) map[string]string {
	for _, r := range rows {
		if r[keyField] == key {
			return r
		}
	}
	return nil
}

// loadEditionByKey loads one edition by key from CSVs. Returns (nil, nil) if key not found.
func (s *EditionCSV) loadEditionByKey(key string) (*model.Edition, error) {
	msRows, _ := s.loadCSVRecordsOptional(relItemsManuscript)
	printRows, _ := s.loadCSVRecordsOptional(relItemsPrint)
	if msRows == nil && printRows == nil {
		return nil, nil
	}
	var itemRow map[string]string
	isManuscript := false
	if itemRow = findRowByKey(msRows, "key", key); itemRow != nil {
		isManuscript = true
	} else if itemRow = findRowByKey(printRows, "key", key); itemRow != nil {
		// print
	} else {
		return nil, nil
	}

	ed := &model.Edition{
		Key:              key,
		ShortTitle:       itemRow["short_title"],
		ShortTitleSource: itemRow["short_title_source"],
		Notes:            itemRow["notes"],
		IsManuscript:     isManuscript,
		HasDiagrams:      formatcov.StrToBoolPtr(itemRow["has_diagrams"]),
	}

	if isManuscript {
		ed.ManuscriptYearFrom = formatcov.IntOpt(itemRow["year_from"])
		ed.ManuscriptYearTo = formatcov.IntOpt(itemRow["year_to"])
		mdRows, _ := s.loadCSVRecordsOptional(relMDManuscript)
		if md := findRowByKey(mdRows, "key", key); md != nil {
			ed.IsElements = true
			ed.ManuscriptClass = md["class"]
			ed.ManuscriptSubclass = formatcov.StrToPtr(md["subclass"])
			ed.Books = formatcov.CompressedStrToInts(md["elements_books"])
		}
	} else {
		ed.Cities = splitNonEmpty(itemRow["city"])
		ed.Year = formatcov.StrToPtr(itemRow["year"])
		ed.Languages = splitNonEmpty(strings.ToLower(itemRow["language"]))
		ed.Editor = splitNonEmpty(itemRow["author_or_editor"])
		ed.Publisher = splitNonEmpty(itemRow["publisher"])
		ed.Format = formatcov.IntOpt(itemRow["format"])
		ed.Volumes = formatcov.IntOpt(itemRow["volumes"])
		ed.USTCId = formatcov.StrToPtr(itemRow["ustc_id"])
		trRows, _ := s.loadCSVRecordsOptional(relTranscriptions)
		if tr := findRowByKey(trRows, "key", key); tr != nil {
			ed.Title = formatcov.StrToPtr(tr["title"])
			ed.Imprint = formatcov.StrToPtr(tr["imprint"])
			ed.Colophon = formatcov.StrToPtr(tr["colophon"])
			ed.Frontispiece = formatcov.StrToPtr(tr["frontispiece"])
		}
		mdRows, _ := s.loadCSVRecordsOptional(relMDPrint)
		if md := findRowByKey(mdRows, "key", key); md != nil {
			ed.IsElements = true
			ed.Books = formatcov.CompressedStrToInts(md["elements_books"])
			ed.AdditionalContent = splitNonEmpty(md["additional_content"])
		}
		tlRows, _ := s.loadCSVRecordsOptional(relTranslations)
		for _, r := range tlRows {
			if r["key"] != key {
				continue
			}
			switch r["field"] {
			case "title":
				ed.TitleEN = formatcov.StrToPtr(r["en"])
			case "imprint":
				ed.ImprintEN = formatcov.StrToPtr(r["en"])
			case "colophon":
				ed.ColophonEN = formatcov.StrToPtr(r["en"])
			case "frontispiece":
				ed.FrontispieceEN = formatcov.StrToPtr(r["en"])
			}
		}
	}

	shRows, _ := s.loadCSVRecordsOptional(relShelfmarks)
	for _, r := range shRows {
		if r["key"] != key {
			continue
		}
		ed.Shelfmarks = append(ed.Shelfmarks, model.EditionShelfmark{
			Volume:          formatcov.IntOpt(r["volume"]),
			Scan:            r["scan"],
			Shelfmark:       r["shelf_mark"],
			TitlePageImg:    r["title_page_img"],
			FrontispieceImg: r["frontispiece_img"],
			Annotations:     r["annotations"],
			Copyright:       r["copyright"],
		})
	}

	corpRows, _ := s.loadCSVRecordsOptional(relCorpuses)
	if cr := findRowByKey(corpRows, "key", key); cr != nil && cr["study"] != "" {
		ed.Corpus = splitNonEmpty(cr["study"])
	}

	bibRows, _ := s.loadCSVRecordsOptional(relBibliography)
	for _, r := range bibRows {
		if r["key"] == key && r["citation"] != "" {
			ed.Bibliography = append(ed.Bibliography, r["citation"])
		}
	}

	revRows, _ := s.loadCSVRecordsOptional(relReviews)
	ed.Verified = findRowByKey(revRows, "key", key) != nil

	ciRows, _ := s.loadCSVRecordsOptional(relClusterItems)
	for _, r := range ciRows {
		if r["item_key"] == key && r["cluster_key"] != "" {
			// find parent in same cluster
			for _, r2 := range ciRows {
				if r2["cluster_key"] == r["cluster_key"] && r2["item_key"] != key {
					ed.ReprintOf = formatcov.StrToPtr(r2["item_key"])
					break
				}
			}
			break
		}
	}

	locRows, _ := s.loadCSVRecordsOptional(relLocators)
	locByKey := make(map[string]*model.EditionLocator)
	for _, r := range locRows {
		loc := rowToLocator(r)
		if loc != nil {
			locByKey[loc.Key] = loc
		}
	}
	veRows, _ := s.loadCSVRecordsOptional(relVisualElements)
	exRows, _ := s.loadCSVRecordsOptional(relVisualElementsEx)
	for _, r := range veRows {
		if r["key"] != key {
			continue
		}
		ve := model.EditionVisualElement{
			VisualElementType: r["visual_element_type"],
			LocatorType:       r["locator_type"],
			Notes:             r["notes"],
		}
		if locKey := r["locator_key"]; locKey != "" {
			ve.Locator = locByKey[locKey]
		}
		for _, ex := range exRows {
			if ex["key"] != key {
				continue
			}
			exLocKey := ex["locator_key"]
			if ve.Locator != nil && exLocKey != ve.Locator.Key {
				continue
			}
			if ve.Locator == nil && exLocKey != "" {
				continue
			}
			exItem := model.EditionVisualExample{Img: ex["path"]}
			if exLocKey != "" {
				exItem.Locator = locByKey[exLocKey]
				exItem.HasLocator = true
			}
			ve.Examples = append(ve.Examples, exItem)
		}
		ed.VisualElements = append(ed.VisualElements, ve)
	}

	diagramKeys, _ := s.loadDiagramDirectoryKeys()
	if diagramKeys != nil {
		_, ed.DiagramCropsAvailable = diagramKeys[key]
	}

	return ed, nil
}

func (s *EditionCSV) collectEditionKeys() ([]string, error) {
	seen := make(map[string]struct{})
	var keys []string
	for _, rel := range []string{relItemsManuscript, relItemsPrint} {
		rows, err := s.loadCSVRecordsOptional(rel)
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			k := r["key"]
			if k != "" {
				if _, ok := seen[k]; !ok {
					seen[k] = struct{}{}
					keys = append(keys, k)
				}
			}
		}
	}
	return keys, nil
}

// preloadedEditionRows holds all CSV data in memory for one bulk load.
type preloadedEditionRows struct {
	msRows         []map[string]string
	printRows      []map[string]string
	mdManuscript   []map[string]string
	mdPrint        []map[string]string
	transcriptions []map[string]string
	translations   []map[string]string
	shelfmarks     []map[string]string
	corpuses       []map[string]string
	bibliography   []map[string]string
	reviews        []map[string]string
	clusterItems   []map[string]string
	locators       []map[string]string
	veRows         []map[string]string
	exRows         []map[string]string
	diagramDirKeys map[string]struct{}
}

// loadAllCSVsOnce reads each edition CSV once. Missing files yield nil slices.
func (s *EditionCSV) loadAllCSVsOnce() (*preloadedEditionRows, error) {
	p := &preloadedEditionRows{}
	p.msRows, _ = s.loadCSVRecordsOptional(relItemsManuscript)
	p.printRows, _ = s.loadCSVRecordsOptional(relItemsPrint)
	if p.msRows == nil && p.printRows == nil {
		return p, nil
	}
	p.mdManuscript, _ = s.loadCSVRecordsOptional(relMDManuscript)
	p.mdPrint, _ = s.loadCSVRecordsOptional(relMDPrint)
	p.transcriptions, _ = s.loadCSVRecordsOptional(relTranscriptions)
	p.translations, _ = s.loadCSVRecordsOptional(relTranslations)
	p.shelfmarks, _ = s.loadCSVRecordsOptional(relShelfmarks)
	p.corpuses, _ = s.loadCSVRecordsOptional(relCorpuses)
	p.bibliography, _ = s.loadCSVRecordsOptional(relBibliography)
	p.reviews, _ = s.loadCSVRecordsOptional(relReviews)
	p.clusterItems, _ = s.loadCSVRecordsOptional(relClusterItems)
	p.locators, _ = s.loadCSVRecordsOptional(relLocators)
	p.veRows, _ = s.loadCSVRecordsOptional(relVisualElements)
	p.exRows, _ = s.loadCSVRecordsOptional(relVisualElementsEx)
	p.diagramDirKeys, _ = s.loadDiagramDirectoryKeys()
	return p, nil
}

// loadDiagramDirectoryKeys reads diagram-directories.json and returns a set of edition keys.
func (s *EditionCSV) loadDiagramDirectoryKeys() (map[string]struct{}, error) {
	path := s.csvPath(relDiagramDirs)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var keys []string
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, err
	}
	out := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		if k != "" {
			out[k] = struct{}{}
		}
	}
	return out, nil
}

// LoadAllEditions reads each CSV once and builds all editions in memory. Use for cache warming.
func (s *EditionCSV) LoadAllEditions() (map[string]*model.Edition, error) {
	var out = make(map[string]*model.Edition)
	keys, err := s.collectEditionKeys()
	if err != nil {
		return nil, err
	}
	if len(keys) == 0 {
		return out, nil
	}
	preloaded, err := s.loadAllCSVsOnce()
	if err != nil {
		return nil, err
	}
	if preloaded.msRows == nil && preloaded.printRows == nil {
		return out, nil
	}
	for _, key := range keys {
		ed := s.buildEditionFromPreloaded(key, preloaded)
		if ed != nil {
			out[key] = ed
		}
	}
	return out, nil
}

// buildEditionFromPreloaded constructs one edition from preloaded CSV rows. Returns nil if key not found.
func (s *EditionCSV) buildEditionFromPreloaded(key string, p *preloadedEditionRows) *model.Edition {
	var itemRow map[string]string
	isManuscript := false
	if itemRow = findRowByKey(p.msRows, "key", key); itemRow != nil {
		isManuscript = true
	} else if itemRow = findRowByKey(p.printRows, "key", key); itemRow != nil {
	} else {
		return nil
	}
	ed := &model.Edition{
		Key:              key,
		ShortTitle:       itemRow["short_title"],
		ShortTitleSource: itemRow["short_title_source"],
		Notes:            itemRow["notes"],
		IsManuscript:     isManuscript,
		HasDiagrams:      formatcov.StrToBoolPtr(itemRow["has_diagrams"]),
	}
	if isManuscript {
		ed.ManuscriptYearFrom = formatcov.IntOpt(itemRow["year_from"])
		ed.ManuscriptYearTo = formatcov.IntOpt(itemRow["year_to"])
		if md := findRowByKey(p.mdManuscript, "key", key); md != nil {
			ed.IsElements = true
			ed.ManuscriptClass = md["class"]
			ed.ManuscriptSubclass = formatcov.StrToPtr(md["subclass"])
			ed.Books = formatcov.CompressedStrToInts(md["elements_books"])
		}
	} else {
		ed.Cities = splitNonEmpty(itemRow["city"])
		ed.Year = formatcov.StrToPtr(itemRow["year"])
		ed.Languages = splitNonEmpty(strings.ToLower(itemRow["language"]))
		ed.Editor = splitNonEmpty(itemRow["author_or_editor"])
		ed.Publisher = splitNonEmpty(itemRow["publisher"])
		ed.Format = formatcov.IntOpt(itemRow["format"])
		ed.Volumes = formatcov.IntOpt(itemRow["volumes"])
		ed.USTCId = formatcov.StrToPtr(itemRow["ustc_id"])
		if tr := findRowByKey(p.transcriptions, "key", key); tr != nil {
			ed.Title = formatcov.StrToPtr(tr["title"])
			ed.Imprint = formatcov.StrToPtr(tr["imprint"])
			ed.Colophon = formatcov.StrToPtr(tr["colophon"])
			ed.Frontispiece = formatcov.StrToPtr(tr["frontispiece"])
		}
		if md := findRowByKey(p.mdPrint, "key", key); md != nil {
			ed.IsElements = true
			ed.Books = formatcov.CompressedStrToInts(md["elements_books"])
			ed.AdditionalContent = splitNonEmpty(md["additional_content"])
		}
		for _, r := range p.translations {
			if r["key"] != key {
				continue
			}
			switch r["field"] {
			case "title":
				ed.TitleEN = formatcov.StrToPtr(r["en"])
			case "imprint":
				ed.ImprintEN = formatcov.StrToPtr(r["en"])
			case "colophon":
				ed.ColophonEN = formatcov.StrToPtr(r["en"])
			case "frontispiece":
				ed.FrontispieceEN = formatcov.StrToPtr(r["en"])
			}
		}
	}
	for _, r := range p.shelfmarks {
		if r["key"] != key {
			continue
		}
		ed.Shelfmarks = append(ed.Shelfmarks, model.EditionShelfmark{
			Volume:          formatcov.IntOpt(r["volume"]),
			Scan:            r["scan"],
			Shelfmark:       r["shelf_mark"],
			TitlePageImg:    r["title_page_img"],
			FrontispieceImg: r["frontispiece_img"],
			Annotations:     r["annotations"],
			Copyright:       r["copyright"],
		})
	}
	if cr := findRowByKey(p.corpuses, "key", key); cr != nil && cr["study"] != "" {
		ed.Corpus = splitNonEmpty(cr["study"])
	}
	for _, r := range p.bibliography {
		if r["key"] == key && r["citation"] != "" {
			ed.Bibliography = append(ed.Bibliography, r["citation"])
		}
	}
	ed.Verified = findRowByKey(p.reviews, "key", key) != nil
	for _, r := range p.clusterItems {
		if r["item_key"] == key && r["cluster_key"] != "" {
			for _, r2 := range p.clusterItems {
				if r2["cluster_key"] == r["cluster_key"] && r2["item_key"] != key {
					ed.ReprintOf = formatcov.StrToPtr(r2["item_key"])
					break
				}
			}
			break
		}
	}
	locByKey := make(map[string]*model.EditionLocator)
	for _, r := range p.locators {
		loc := rowToLocator(r)
		if loc != nil {
			locByKey[loc.Key] = loc
		}
	}
	for _, r := range p.veRows {
		if r["key"] != key {
			continue
		}
		ve := model.EditionVisualElement{
			VisualElementType: r["visual_element_type"],
			LocatorType:       r["locator_type"],
			Notes:             r["notes"],
		}
		if locKey := r["locator_key"]; locKey != "" {
			ve.Locator = locByKey[locKey]
		}
		for _, ex := range p.exRows {
			if ex["key"] != key {
				continue
			}
			exLocKey := ex["locator_key"]
			if ve.Locator != nil && exLocKey != ve.Locator.Key {
				continue
			}
			if ve.Locator == nil && exLocKey != "" {
				continue
			}
			exItem := model.EditionVisualExample{Img: ex["path"]}
			if exLocKey != "" {
				exItem.Locator = locByKey[exLocKey]
				exItem.HasLocator = true
			}
			ve.Examples = append(ve.Examples, exItem)
		}
		ed.VisualElements = append(ed.VisualElements, ve)
	}
	if p.diagramDirKeys != nil {
		_, ed.DiagramCropsAvailable = p.diagramDirKeys[key]
	}
	return ed
}

func splitNonEmpty(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func rowToLocator(r map[string]string) *model.EditionLocator {
	if r["key"] == "" {
		return nil
	}
	return &model.EditionLocator{
		Key:             r["key"],
		Value:           r["value"],
		PageType:        r["page_type"],
		PageValue:       formatcov.StrToPtr(r["page_value"]),
		Type:            formatcov.StrToPtr(r["type"]),
		FirstOrderType:  formatcov.StrToPtr(r["first_order_type"]),
		FirstOrderValue: formatcov.StrToPtr(r["first_order_value"]),
	}
}

func locatorRow(l *model.EditionLocator) map[string]string {
	if l == nil {
		return nil
	}
	m := map[string]string{
		"key":               l.Key,
		"value":             l.Value,
		"page_type":         l.PageType,
		"page_value":        formatcov.PtrToStr(l.PageValue),
		"type":              formatcov.PtrToStr(l.Type),
		"first_order_type":  formatcov.PtrToStr(l.FirstOrderType),
		"first_order_value": formatcov.PtrToStr(l.FirstOrderValue),
	}
	return m
}
