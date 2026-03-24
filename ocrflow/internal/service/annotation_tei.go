package service

import (
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/titlepage"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	tei2 "github.com/MiaMish/elements-dh/ocrflow/pkg/tei"
	teimodel "github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/textmatch"
	"github.com/samber/lo"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	datasetSvc    *Dataset
	fileSysMgt    *filesys.Manager
	resultSvc     *Result
	featureSvc    *Feature
	editionSvc    *Edition
}

func NewAnnotationTEI(annotationSvc *Annotation, datasetSvc *Dataset, fileSysMgt *filesys.Manager, resultSvc *Result, featureSvc *Feature, editionSvc *Edition) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		datasetSvc:    datasetSvc,
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

	if !t.annotationCanBeRepresentedAsTEI(ann) {
		return nil, fmt.Errorf("annotation %s cannot be represented as TEI", ann.ID)
	}

	features, err = t.normalizeFeatureIDs(datasetID, features)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize feature IDs for dataset %s: %v", datasetID, err)
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

func (t *AnnotationTEI) normalizeFeatureIDs(datasetID string, features []string) ([]string, error) {
	if len(features) != 0 {
		return features, nil
	}

	// if no features specified, default to all features for the dataset.
	allFeatures, err := t.featureSvc.ListFeatures(datasetID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list features for dataset %s: %v", datasetID, err)
	}
	return lo.Map(allFeatures, func(f *feature.Feature, _ int) string {
		return f.ID
	}), nil
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
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
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

	return tei2.BuildTEIFromLines(pageNumOrKey, pageLines, buildItems(results, feats, lines), imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
}

func (t *AnnotationTEI) getTitlePageTexts(editionKey string) (transcription []string, translations map[string][]string, err error) {
	edition, err := t.editionSvc.GetEditionByID(editionKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get edition by ID %s: %v", editionKey, err)
	}
	if edition == nil {
		return nil, nil, fmt.Errorf("%w: edition with key %s does not exist", ErrEditionNotFound, editionKey)
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

func (t *AnnotationTEI) GetTEIs(datasetID string, annotationID string, keys []string, features []string, fallbackToOrigin bool) ([]byte, error) {
	features, err := t.normalizeFeatureIDs(datasetID, features)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize feature IDs for dataset %s: %v", datasetID, err)
	}

	teis := map[string]*teimodel.TEI{}
	for _, key := range keys {
		ann, err := t.annotationSvc.Get(datasetID, annotationID)
		if err != nil {
			return nil, err
		}

		if !t.annotationCanBeRepresentedAsTEI(ann) {
			continue
		}

		tei, err := t.getTEI(ann, key, features, fallbackToOrigin)
		if err != nil {
			return nil, fmt.Errorf("failed to get TEI for annotation %s: %v", ann.ID, err)
		}

		teis[key] = tei
	}

	combinedTEI, err := tei2.CombineTEIsByKey(teis)
	if err != nil {
		return nil, fmt.Errorf("failed to combine TEIs: %v", err)
	}
	xml, err := combinedTEI.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize combined TEI to XML: %v", err)
	}
	return xml, nil
}

func (t *AnnotationTEI) annotationCanBeRepresentedAsTEI(ann *annotation.Annotation) bool {
	return ann.Ocred || ann.LinesDetected || ann.Segmented
}

func (t *AnnotationTEI) getBibleMetadata(datasetID string, pageOrKey string) *teimodel.BiblFull {
	edID := pageOrKey
	ds, err := t.datasetSvc.Get(datasetID)
	if err == nil && ds.EditionID != "" {
		edID = ds.EditionID
	}
	ed, err := t.editionSvc.GetEditionByID(edID)
	if err != nil {
		return nil
	}
	return model.EditionToBiblFull(ed)
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

			matches := textmatch.FindLoosePhraseMatches(fullContent, surface)
			if len(matches) == 0 {
				log.Printf("warning: no matches found for result, skipping: featurd=%s key=%s dataset=%s", res.FeatureID, res.PageKey, res.DatasetID)
			}
			for _, match := range matches {
				startIndex := match[0]
				endIndex := match[1]
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

			}
		}
	}

	return items
}
