package service

import (
	"fmt"
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
)

type AnnotationSearch struct {
	annotationSvc *Annotation
	fileSysMgt    *filesys.Manager
	resultSvc     *Result
}

func NewAnnotationSearch(annotationSvc *Annotation, fileSysMgt *filesys.Manager, resultSvc *Result) *AnnotationSearch {
	return &AnnotationSearch{
		annotationSvc: annotationSvc,
		fileSysMgt:    fileSysMgt,
		resultSvc:     resultSvc,
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

	if strings.TrimSpace(ann.Pages) == "" {
		if err := s.searchAnnotationContents(as, ann, rg); err != nil {
			return nil, err
		}
		return as, nil
	}

	pages, err := pagesparser.IntRange(ann.Pages)
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		a, _, err := s.fileSysMgt.RetrieveAnnotationAltoPage(ann, fmt.Sprintf("%d", page))
		if err != nil {
			return nil, err
		}

		for _, cat := range as.Categories {
			bls, err := alto.ExtractBlocksByCategory(a, cat)
			if err != nil {
				return nil, err
			}

			for _, bl := range bls {
				combined := alto.ExtractTextContentFromBlock(bl)
				if rg.MatchString(combined) {
					highlighted := s.highlightSearchMatch(rg, combined)

					as.Results = append(as.Results, &common.ALTOPart{
						Category: cat,
						Content:  highlighted,
						Location: common.ALTOLocation{
							Page:        page,
							TextBlockID: bl.ID,
						},
					})

					if as.MaxResults > 0 && len(as.Results) >= as.MaxResults {
						return as, nil
					}
				}
			}
		}
	}
	return as, nil
}

func (s *AnnotationSearch) searchAnnotationContents(as *annotation.Search, ann *annotation.Annotation, rg *regexp.Regexp) error {
	keys, err := s.listAnnotationKeys(as, ann)
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		lines, translations, err := s.fileSysMgt.RetrieveAnnotationTXTPage(ann, key)
		if err != nil {
			continue
		}
		page := keyToPage(key)

		originalContent := strings.Join(lines, "\n")
		if rg.MatchString(originalContent) {
			as.Results = append(as.Results, &common.ALTOPart{
				Category: "transcription.original",
				Content:  s.highlightSearchMatch(rg, originalContent),
				Location: common.ALTOLocation{
					Page:        page,
					TextBlockID: key,
				},
			})
			if as.MaxResults > 0 && len(as.Results) >= as.MaxResults {
				return nil
			}
		}

		langs := make([]string, 0, len(translations))
		for lang := range translations {
			langs = append(langs, lang)
		}
		slices.Sort(langs)

		for _, lang := range langs {
			content := strings.Join(translations[lang], "\n")
			if !rg.MatchString(content) {
				continue
			}
			as.Results = append(as.Results, &common.ALTOPart{
				Category: fmt.Sprintf("transcription.%s", lang),
				Content:  s.highlightSearchMatch(rg, content),
				Location: common.ALTOLocation{
					Page:        page,
					TextBlockID: key,
				},
			})
			if as.MaxResults > 0 && len(as.Results) >= as.MaxResults {
				return nil
			}
		}
	}

	results, err := s.resultSvc.ListResults(as.DatasetID, ann.ID, keys, nil)
	if err != nil {
		return err
	}

	for _, result := range results {
		page := keyToPage(result.PageKey)
		for _, val := range result.Values {
			if rg.MatchString(val.Surface) {
				as.Results = append(as.Results, &common.ALTOPart{
					Category: fmt.Sprintf("feature.%s.surface", result.FeatureID),
					Content:  s.highlightSearchMatch(rg, val.Surface),
					Location: common.ALTOLocation{
						Page:        page,
						TextBlockID: result.PageKey,
					},
				})
				if as.MaxResults > 0 && len(as.Results) >= as.MaxResults {
					return nil
				}
			}

			for propertyKey, propertyVal := range val.Properties {
				if !rg.MatchString(propertyVal) {
					continue
				}
				as.Results = append(as.Results, &common.ALTOPart{
					Category: fmt.Sprintf("feature.%s.property.%s", result.FeatureID, propertyKey),
					Content:  s.highlightSearchMatch(rg, propertyVal),
					Location: common.ALTOLocation{
						Page:        page,
						TextBlockID: result.PageKey,
					},
				})
				if as.MaxResults > 0 && len(as.Results) >= as.MaxResults {
					return nil
				}
			}
		}
	}

	return nil
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

	results, err := s.resultSvc.ListResults(as.DatasetID, ann.ID, nil, nil)
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

func keyToPage(key string) int {
	page, err := strconv.Atoi(key)
	if err != nil || page < 1 {
		return 0
	}
	return page
}

func (s *AnnotationSearch) highlightSearchMatch(rg *regexp.Regexp, v string) string {
	return rg.ReplaceAllString(v, "<em>$0</em>")
}
