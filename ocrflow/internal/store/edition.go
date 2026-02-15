package store

import (
	"fmt"
	"math/rand"
	"mime/multipart"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/csv"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
)

type EditionCSV struct {
	docsDir string
	tpsDir  string
}

const (
	docsSubdir = "csvs"

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
)

func NewEditionCSV(docsPublicDir string) *EditionCSV {
	return &EditionCSV{
		docsDir: filepath.Join(docsPublicDir, docsSubdir),
		tpsDir:  filepath.Join(docsPublicDir, "tps"),
	}
}

func (s *EditionCSV) csvPath(rel string) string {
	return filepath.Join(s.docsDir, rel)
}

// UpdateNotes updates the notes field for the given key in items_print.csv.
func (s *EditionCSV) UpdateNotes(key, note string) error {
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
	return csv.SaveCSVRecords(relItemsPrint, header, rows)
}

func (s *EditionCSV) UpsertEdition(ed *model.Edition, user string) error {
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
		"has_diagrams":       "",
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
		"has_diagrams":       "",
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
	return nil
}

func (s *EditionCSV) UploadImage(key string, typ string, ext string, file multipart.File) (*model.ImageUpload, error) {
	p := path.Join(s.tpsDir, fmt.Sprintf("%s_%s.%s", key, typ, ext))
	if err := futils.WriteMultipartFileToPath(file, p); err != nil {
		return nil, fmt.Errorf("Error saving uploaded image: %v\n", err)
	}
	return &model.ImageUpload{
		Success:  true,
		Filename: filepath.Base(p),
		Path:     filepath.Join(filepath.Dir(p), filepath.Base(p)),
	}, nil
}

func (s *EditionCSV) GetEditionByID(key string) (*model.Edition, error) {
	// todo impl
	return nil, fmt.Errorf("GetEditionByID not implemented yet")
}

func (s *EditionCSV) ListEditions() ([]*model.Edition, error) {
	// todo impl
	return nil, fmt.Errorf("ListEditions not implemented yet")
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
