package service

import (
	"fmt"
	"log"
	"path"
	"sort"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/titlepage"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/futils"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/pagesparser"
	tei2 "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei"
	teimodel "github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/textmatch"
	"github.com/samber/lo"
)

type AnnotationTEI struct {
	annotationSvc *Annotation
	datasetSvc    *Dataset
	fileSysMgt    *filesys.Manager
	datasetImgSvc *DatasetImg
	resultSvc     *Result
	featureSvc    *Feature
	editionSvc    *Edition
}

func NewAnnotationTEI(annotationSvc *Annotation, datasetSvc *Dataset, fileSysMgt *filesys.Manager, datasetImgSvc *DatasetImg, resultSvc *Result, featureSvc *Feature, editionSvc *Edition) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc: annotationSvc,
		datasetSvc:    datasetSvc,
		fileSysMgt:    fileSysMgt,
		datasetImgSvc: datasetImgSvc,
		resultSvc:     resultSvc,
		featureSvc:    featureSvc,
		editionSvc:    editionSvc,
	}
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNumOrKey string, features []string, fallbackToOrigin bool, imageVariant model.ImageVariant) ([]byte, error) {
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
	if err := t.scaleTEIToVariant(datasetID, pageNumOrKey, tei, imageVariant); err != nil {
		return nil, fmt.Errorf("failed to scale TEI for annotation %s: %v", ann.ID, err)
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
	allFeatures, err := t.featureSvc.ListFeatures(feature.NewDatasetDefScope(datasetID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list features for dataset %s: %v", datasetID, err)
	}
	return lo.Map(allFeatures, func(f *feature.Feature, _ int) string {
		return f.ID
	}), nil
}

func (t *AnnotationTEI) GetTxt(datasetID string, annotationID string, pageNumOrKey string, imprintOnly *bool) (string, error) {
	ann, err := t.annotationSvc.Get(datasetID, annotationID)
	if err != nil {
		return "", err
	}

	if !ann.Ocred {
		return "", fmt.Errorf("annotation %s is not OCRed", ann.ID)
	}

	// 1) Try ALTO page first
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		return alto.ExtractTextContentsFromAlto(a), nil
	}
	if a, _, err := t.fileSysMgt.RetrieveAnnotationTranscriptionAltoPage(ann, pageNumOrKey); err == nil {
		return alto.ExtractTextContentsFromAlto(a), nil
	}

	if md, err := t.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, pageNumOrKey); err == nil {
		return md.Content, nil
	}

	// 2) TXT fallback: transcription
	var lines []string
	if ann.DatasetID == titlepage.DatasetID {
		if lines, _, err = t.getTitlePageTexts(pageNumOrKey, imprintOnly); err != nil {
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
	results, err := t.resultSvc.ListResults(feature.NewDatasetExecScope(ann.DatasetID, ann.ID), []string{pageNumOrKey}, features, fallbackToOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to list results for annotation %s: %v", ann.ID, err)
	}
	feats, err := t.featureSvc.ListFeatures(feature.NewDatasetDefScope(ann.DatasetID), []feature.ExpandOptions{feature.ExpandRevisions})
	if err != nil {
		return nil, fmt.Errorf("failed to list features for annotation %s: %v", ann.ID, err)
	}

	// Try ALTO page created in the platform
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		// todo: support items in alto
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Fallbacks: preloaded files from the transcription page
	// Between file formats, ALTO format is preferred over Markdown, which is preferred over TXT.

	// ALTO
	if a, _, err := t.fileSysMgt.RetrieveAnnotationTranscriptionAltoPage(ann, pageNumOrKey); err == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Markdown
	if md, err := t.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, pageNumOrKey); err == nil {
		return tei2.BuildTEIFromMarkdown(pageNumOrKey, md, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Text
	var (
		lines        []string
		translations map[string][]string
		imageURL     string
	)

	if ann.DatasetID == titlepage.DatasetID {
		if lines, translations, err = t.getTitlePageTexts(pageNumOrKey, nil); err != nil {
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

func (t *AnnotationTEI) getTitlePageTexts(editionKey string, imprintOnly *bool) (transcription []string, translations map[string][]string, err error) {
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

	titleLines := strings.Split(*edition.Title, "\n")
	var imprintLines []string
	if edition.Imprint != nil {
		imprintLines = strings.Split(*edition.Imprint, "\n")
	}

	switch {
	case imprintOnly == nil:
		transcription = append(titleLines, imprintLines...)
	case *imprintOnly:
		transcription = imprintLines
	default:
		transcription = titleLines
	}

	if edition.TitleEN == nil {
		return transcription, nil, nil
	}

	titleENLines := strings.Split(*edition.TitleEN, "\n")
	var imprintENLines []string
	if edition.ImprintEN != nil {
		imprintENLines = strings.Split(*edition.ImprintEN, "\n")
	}

	translations = map[string][]string{}
	switch {
	case imprintOnly == nil:
		translations["en"] = append(titleENLines, imprintENLines...)
	case *imprintOnly:
		translations["en"] = imprintENLines
	default:
		translations["en"] = titleENLines
	}

	return transcription, translations, nil
}

func (t *AnnotationTEI) GetTEIs(datasetID string, annotationID string, keys []string, features []string, fallbackToOrigin bool, imageVariant model.ImageVariant) ([]byte, error) {
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
		if err := t.scaleTEIToVariant(datasetID, key, tei, imageVariant); err != nil {
			return nil, fmt.Errorf("failed to scale TEI for annotation %s page %s: %v", ann.ID, key, err)
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

func (t *AnnotationTEI) scaleTEIToVariant(datasetID, pageNumOrKey string, tei *teimodel.TEI, imageVariant model.ImageVariant) error {
	if tei == nil || imageVariant == model.ImageVariantOriginal {
		return nil
	}

	origWidth, origHeight, err := t.datasetImgSvc.ImageDimensions(datasetID, pageNumOrKey, model.ImageVariantOriginal)
	if err != nil {
		return fmt.Errorf("get original image dimensions: %w", err)
	}
	varWidth, varHeight, err := t.datasetImgSvc.ImageDimensions(datasetID, pageNumOrKey, imageVariant)
	if err != nil {
		return fmt.Errorf("get %s image dimensions: %w", imageVariant, err)
	}
	if origWidth <= 0 || origHeight <= 0 || varWidth <= 0 || varHeight <= 0 {
		return fmt.Errorf("invalid dimensions for TEI scaling")
	}

	scaleX := float64(varWidth) / float64(origWidth)
	scaleY := float64(varHeight) / float64(origHeight)
	for si := range tei.Facsimile.Surfaces {
		surface := &tei.Facsimile.Surfaces[si]
		surface.ULX *= scaleX
		surface.ULY *= scaleY
		surface.LRX *= scaleX
		surface.LRY *= scaleY
		for zi := range tei.Facsimile.Surfaces[si].Zones {
			zone := &surface.Zones[zi]
			zone.ULX *= scaleX
			zone.ULY *= scaleY
			zone.LRX *= scaleX
			zone.LRY *= scaleY
		}
	}
	return nil
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
				log.Printf("warning: no matches found for result, skipping: feature=%s key=%s dataset=%s annotatio=%s", res.FeatureID, res.Key, res.Scope.DatasetID, res.Scope.AnnotationID)
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
