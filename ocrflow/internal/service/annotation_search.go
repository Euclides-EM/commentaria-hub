package service

import (
	"fmt"
	"regexp"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/common"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
)

type AnnotationSearch struct {
	annotationSvc *Annotation
	fileSysMgt    *filesys.Manager
}

func NewAnnotationSearch(annotationSvc *Annotation, fileSysMgt *filesys.Manager) *AnnotationSearch {
	return &AnnotationSearch{
		annotationSvc: annotationSvc,
		fileSysMgt:    fileSysMgt,
	}
}

func (s *AnnotationSearch) Search(as *annotation.Search) (*annotation.Search, error) {
	ann, err := s.annotationSvc.Get(as.DatasetID, as.AnnotationId)
	if err != nil {
		return nil, err
	}

	pages, err := pagesparser.Range(ann.Pages)
	if err != nil {
		return nil, err
	}

	as.Regex = "(?mi)" + as.Regex
	rg, err := regexp.Compile(as.Regex)
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
					highlighted := rg.ReplaceAllString(combined, "<em>$0</em>")

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
