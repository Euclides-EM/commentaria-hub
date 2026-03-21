package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

type AnnotationSearch struct {
	annotationSvc    *Annotation
	fileSysMgt       *filesys.Manager
	resultSvc        *Result
	annotationTEISvc *AnnotationTEI
}

func NewAnnotationSearch(annotationSvc *Annotation, fileSysMgt *filesys.Manager, resultSvc *Result, annotationTEISvc *AnnotationTEI) *AnnotationSearch {
	return &AnnotationSearch{
		annotationSvc:    annotationSvc,
		fileSysMgt:       fileSysMgt,
		resultSvc:        resultSvc,
		annotationTEISvc: annotationTEISvc,
	}
}

func (s *AnnotationSearch) Search(as *annotation.Search) (*annotation.Search, error) {
	ann, err := s.annotationSvc.Get(as.DatasetID, as.AnnotationId)
	if err != nil {
		return nil, err
	}

	as.Regex = "(?mi)" + as.Regex
	rg, err := regexp.Compile(as.Regex)
	if err != nil {
		return nil, err
	}

	pages, err := pagesparser.Range(ann.Pages)
	if err != nil {
		return nil, err
	}

	if len(as.SearchWithin) == 0 {
		as.SearchWithin = []annotation.SearchWithin{
			annotation.SearchWithinCategories,
		}
	}

	for _, page := range pages {
		for _, sw := range as.SearchWithin {
			maxResultsForPage := as.MaxResults
			if maxResultsForPage > 0 {
				maxResultsForPage = as.MaxResults - len(as.Results)
				if maxResultsForPage <= 0 {
					break
				}
			}
			if sw == annotation.SearchWithinCategories {
				res, err := s.searchWithinCategories(ann, maxResultsForPage, rg, page, as.Categories)
				if err != nil {
					return nil, err
				}
				as.Results = append(as.Results, res...)
				continue
			}
			var extractor func(*model.TEI) []string
			switch sw {
			case annotation.SearchWithinTranscription:
				extractor = tei.ExtractTranscriptionLines
			case annotation.SearchWithinTranslation:
				extractor = tei.ExtractTranslationLines
			case annotation.SearchWithinBiblioMetadata:
				extractor = tei.ExtractBiblMetadataLines
			default:
				return nil, fmt.Errorf("unsupported search within type: %v", sw)
			}
			res, err := s.searchWithinByTEIExtractor(ann, rg, page, extractor)
			if err != nil {
				return nil, err
			}
			if res != nil {
				as.Results = append(as.Results, res)
			}
		}

	}
	return as, nil
}

func (s *AnnotationSearch) listAnnotationKeys(as *annotation.Search, ann *annotation.Annotation) ([]string, error) {
	keySet := map[string]struct{}{}

	transcriptionRoot := s.fileSysMgt.AnnotationTxtTranscriptionDir(ann)
	if entries, err := os.ReadDir(transcriptionRoot); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				keySet[entry.Name()] = struct{}{}
			}
		}
	}

	results, err := s.resultSvc.ListResults(as.DatasetID, ann.ID, nil, nil, true)
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		if strings.TrimSpace(result.PageKey) == "" {
			continue
		}
		keySet[result.PageKey] = struct{}{}
	}

	altoDir := s.fileSysMgt.DatasetAnnotationAltoDir(ann)
	if entries, err := os.ReadDir(altoDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			if filepath.Ext(entry.Name()) != ".xml" {
				continue
			}
			base := strings.TrimSuffix(entry.Name(), ".xml")
			if base == "" || strings.EqualFold(base, "METS") {
				continue
			}
			if strings.HasPrefix(base, "page-") {
				page, convErr := strconv.Atoi(strings.TrimPrefix(base, "page-"))
				if convErr == nil && page > 0 {
					keySet[strconv.Itoa(page)] = struct{}{}
					continue
				}
			}
			keySet[base] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys, nil
}

func (s *AnnotationSearch) highlightSearchMatch(rg *regexp.Regexp, v string) string {
	return rg.ReplaceAllString(v, "<em>$0</em>")
}

func (s *AnnotationSearch) searchWithinByTEIExtractor(ann *annotation.Annotation, rg *regexp.Regexp, page string, linesExtractor func(*model.TEI) []string) (*common.ALTOPart, error) {
	t, err := s.annotationTEISvc.getTEI(ann, page, nil, true)
	if err != nil {
		log.Printf("WARNING: Could not retrieve TEI for page %v: %v", page, err)
		return nil, err
	}
	lines := linesExtractor(t)
	combined := strings.Join(lines, "\n")

	if rg.MatchString(combined) {
		highlighted := s.highlightSearchMatch(rg, combined)

		return &common.ALTOPart{
			Content: highlighted,
			Location: common.ALTOLocation{
				Page: page,
			},
		}, nil
	}

	return nil, nil
}

func (s *AnnotationSearch) searchWithinCategories(ann *annotation.Annotation, maxResults int, rg *regexp.Regexp, page string, categories []string) ([]*common.ALTOPart, error) {
	var results []*common.ALTOPart
	a, _, err := s.fileSysMgt.RetrieveAnnotationAltoPage(ann, page)
	if err != nil {
		log.Printf("WARNING: Could not retrieve ALTO for page: %v", err)
		return results, err
	}
	for _, cat := range categories {
		bls, err := alto.ExtractBlocksByCategory(a, cat)
		if err != nil {
			return nil, err
		}

		for _, bl := range bls {
			combined := alto.ExtractTextContentFromBlock(bl)
			if rg.MatchString(combined) {
				highlighted := s.highlightSearchMatch(rg, combined)

				results = append(results, &common.ALTOPart{
					Category: cat,
					Content:  highlighted,
					Location: common.ALTOLocation{
						Page:        page,
						TextBlockID: bl.ID,
					},
				})

				if maxResults <= 0 || len(results) >= maxResults {
					return results, nil
				}
			}
		}
	}
	return results, nil
}
