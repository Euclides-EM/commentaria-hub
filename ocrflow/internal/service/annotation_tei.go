package service

import (
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/titlepage"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	tei2 "github.com/MiaMish/elements-dh/ocrflow/pkg/tei"
	teimodel "github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	fileSysMgt    *filesys.Manager
	resultSvc     *Result
	featureSvc    *Feature
	editionSvc    *Edition
}

func NewAnnotationTEI(annotationSvc *Annotation, fileSysMgt *filesys.Manager, resultSvc *Result, featureSvc *Feature, editionSvc *Edition) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		fileSysMgt:    fileSysMgt,
		resultSvc:     resultSvc,
		featureSvc:    featureSvc,
		editionSvc:    editionSvc,
	}
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNumOrKey string, features []string, fallbackToOrigin bool) ([]byte, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return nil, err
	}

	if !ann.Ocred {
		return nil, fmt.Errorf("annotation %s is not OCRed", ann.ID)
	}

	// if no features specified, default to all default features for the dataset.
	if len(features) == 0 {
		allFeatures, err := t.featureSvc.ListFeatures(datasetID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list features for dataset %s: %v", datasetID, err)
		}
		features = lo.Map(lo.Filter(allFeatures, func(f *feature.Feature, _ int) bool {
			return f.IsDefault
		}), func(f *feature.Feature, _ int) string {
			return f.ID
		})
	}

	tei, err := t.getTEI(ann, pageNumOrKey, features, fallbackToOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to get TEI for annotation %s: %v", ann.ID, err)
	}

	xml, err := tei.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize TEI to XML for annotation %s: %v", ann.ID, err)
	}

	return xml, nil
}

func (t *AnnotationTEI) GetTxt(datasetID string, annotationID string, pageNumOrKey string) (string, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return "", err
	}

	if !ann.Ocred {
		return "", fmt.Errorf("annotation %s is not OCRed", ann.ID)
	}

	// 1) Try ALTO page first
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		alto.ExtractTextContentsFromAlto(a)
	}

	// 2) TXT fallback: transcription
	var lines []string
	if ann.DatasetID == titlepage.DatasetID {
		if lines, _, err = t.getTitlePageTexts(pageNumOrKey); err != nil {
			return "", fmt.Errorf("failed to get title page texts for TPS annotation: %v", err)
		}
	} else {
		if lines, _, err = t.fileSysMgt.RetrieveAnnotationTXTPage(ann, pageNumOrKey); err != nil {
			return "", fmt.Errorf("failed to retrieve TXT page for annotation %s: %v", ann.ID, err)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func (t *AnnotationTEI) getTEI(ann *annotation.Annotation, pageNumOrKey string, features []string, fallbackToOrigin bool) (*teimodel.TEI, error) {
	// Load results + feature definitions first (shared by both ALTO and TXT paths)
	results, err := t.resultSvc.ListResults(ann.DatasetID, ann.ID, []string{pageNumOrKey}, features, fallbackToOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to list results for annotation %s: %v", ann.ID, err)
	}
	feats, err := t.featureSvc.ListFeatures(ann.DatasetID, []feature.ExpandOptions{feature.ExpandRevisions})
	if err != nil {
		return nil, fmt.Errorf("failed to list features for annotation %s: %v", ann.ID, err)
	}

	// 1) Try ALTO page first
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		// todo: support items in alto
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL)
	}

	// 2) TXT fallback: transcription + translations + image url
	var (
		lines        []string
		translations map[string][]string
		imageURL     string
	)

	if ann.DatasetID == titlepage.DatasetID {
		if lines, translations, err = t.getTitlePageTexts(pageNumOrKey); err != nil {
			return nil, fmt.Errorf("failed to get title page texts for TPS annotation: %v", err)
		}
		imageURL = futils.LocateFileInDir(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID),
			func(filename string) bool {
				return strings.HasPrefix(filename, pageNumOrKey)
			},
		)
		if imageURL == "" {
			log.Printf("warning: failed to locate image for TPS page %s in dataset %s", pageNumOrKey, ann.DatasetID)
		}
	} else {
		if lines, translations, err = t.fileSysMgt.RetrieveAnnotationTXTPage(ann, pageNumOrKey); err != nil {
			return nil, fmt.Errorf("failed to retrieve TXT page for annotation %s: %v", ann.ID, err)
		}
		imageURL = path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
	}

	pageLines := tei2.Lines{
		TranscriptionLines: lines,
		Translations:       translations,
	}

	return tei2.BuildTEIFromLines(pageNumOrKey, pageLines, buildItems(results, feats, lines), imageURL)
}

func (t *AnnotationTEI) getTitlePageTexts(editionKey string) (transcription []string, translations map[string][]string, err error) {
	edition, err := t.editionSvc.GetEditionByID(editionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get edition by ID %s: %v", editionKey, err)
	}
	if edition == nil {
		return nil, nil, fmt.Errorf("edition with key %s does not exist", editionKey)
	}
	if edition.Title == nil {
		return nil, nil, fmt.Errorf("edition %s does not have a title page transcription", editionKey)
	}
	transcription = strings.Split(*edition.Title, "\n")
	if edition.Imprint != nil {
		transcription = append(transcription, strings.Split(*edition.Imprint, "\n")...)
	}
	if edition.TitleEN == nil {
		return transcription, nil, nil
	}
	translations = map[string][]string{
		"en": strings.Split(*edition.TitleEN, "\n"),
	}
	if edition.ImprintEN != nil {
		translations["en"] = append(translations["en"], strings.Split(*edition.ImprintEN, "\n")...)
	}
	return transcription, translations, nil
}

func buildItems(results []*feature.Result, feats []*feature.Feature, transcription []string) []tei2.EntityItem {
	findFeat := func(id string) *feature.Feature {
		return lo.FindOrElse(feats, nil, func(f *feature.Feature) bool { return f.ID == id })
	}

	fullContent := strings.Join(transcription, "\n")

	// Precompute line start indices in fullContent.
	lineStarts := make([]int, 0, len(transcription))
	pos := 0
	for i, line := range transcription {
		lineStarts = append(lineStarts, pos)
		pos += len(line)
		if i < len(transcription)-1 {
			pos++ // '\n'
		}
	}

	// Convert a global index in fullContent to (line index, byte offset within that line).
	globalToLineOffset := func(idx int) (line int, off int) {
		if idx < 0 {
			return 0, 0
		}
		if idx > len(fullContent) {
			idx = len(fullContent)
		}
		if len(lineStarts) == 0 {
			return 0, 0
		}
		// Find first lineStart > idx, then step back one.
		j := sort.Search(len(lineStarts), func(i int) bool { return lineStarts[i] > idx })
		if j == 0 {
			return 0, idx
		}
		line = j - 1
		off = idx - lineStarts[line]
		// Clamp offset to line length (in case idx hits a newline position).
		if line < len(transcription) && off > len(transcription[line]) {
			off = len(transcription[line])
		}
		return line, off
	}

	var items []tei2.EntityItem
	for _, res := range results {
		feat := findFeat(res.FeatureID)
		if feat == nil {
			log.Printf("warning: feature %s not found for result %s, skipping", res.FeatureID, res.ID)
			continue
		}

		for _, val := range res.Values {
			surface := val.Surface
			if strings.TrimSpace(surface) == "" {
				log.Printf("warning: empty surface form for result %s, skipping", res.ID)
				continue
			}

			props := map[string]string{
				"surface": surface,
			}
			for k, v := range val.Properties {
				props[k] = v
			}

			// Find *all* occurrences of surface in fullContent.
			for from := 0; from <= len(fullContent)-len(surface); {
				startIndex := strings.Index(fullContent[from:], surface)
				if startIndex == -1 {
					break
				}
				startIndex += from
				endIndex := startIndex + len(surface)

				startLine, startOff := globalToLineOffset(startIndex)
				endLine, endOff := globalToLineOffset(endIndex)

				items = append(items, tei2.EntityItem{
					Start: tei2.EntityLocationIndex{
						BlockID:    "1",
						LineID:     fmt.Sprintf("%d", startLine),
						ByteOffset: startOff,
					},
					End: tei2.EntityLocationIndex{
						BlockID:    "1",
						LineID:     fmt.Sprintf("%d", endLine),
						ByteOffset: endOff,
					},
					Category:   feat.Name,
					Properties: props,
				})

				// Move forward to find the next occurrence (including overlapping matches).
				from = startIndex + 1
			}
		}
	}

	return items
}
