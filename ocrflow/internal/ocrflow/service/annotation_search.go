package service

import (
	"regexp"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/ocrflow/store/filesys"
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

func (s *AnnotationSearch) Search(as *model.AnnotationSearch) (*model.AnnotationSearch, error) {
	ann, err := s.annotationSvc.Get(as.DatasetID, as.AnnotationId)
	if err != nil {
		return nil, err
	}

	pages, err := pagesparser.Parse(ann.Pages)
	if err != nil {
		return nil, err
	}

	rg, err := regexp.Compile(as.Regex)
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		a, _, err := s.fileSysMgt.RetrieveAltoPage(ann, page)
		if err != nil {
			return nil, err
		}

		for _, cat := range as.Categories {
			bls, err := alto.ExtractBlocksByCategory(a, cat)
			if err != nil {
				return nil, err
			}

			for _, bl := range bls {
				contents := alto.ExtractTextContentsFromBlock(bl)
				combined := ""
				for _, content := range contents {
					c := strings.TrimSpace(content)
					if strings.HasSuffix(c, "¬") {
						combined += strings.TrimSuffix(c, "¬")
					} else {
						combined += c + " "
					}
				}
				combined = strings.TrimSpace(combined)
				if rg.MatchString(combined) {
					as.Results = append(as.Results, &model.AnnotationPart{
						Category: cat,
						Content:  combined,
						Location: model.AnnotationLocation{
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
