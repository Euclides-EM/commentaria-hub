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

type annotationNotRepresentableAsTEIError struct {
	annotationID string
}

func (e annotationNotRepresentableAsTEIError) Error() string {
	return fmt.Sprintf("annotation %s cannot be represented as TEI", e.annotationID)
}

func (annotationNotRepresentableAsTEIError) HTTPStatusCode() int {
	return 400
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
	ann, features, feats, err := t.prepareTEIRequest(datasetID, annotationID, features)
	if err != nil {
		return nil, err
	}

	tei, err := t.getScaledTEIWithFeatures(ann, pageNumOrKey, features, fallbackToOrigin, feats, imageVariant)
	if err != nil {
		return nil, err
	}

	xml, err := tei.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize TEI to XML for annotation %s: %v", ann.ID, err)
	}

	return xml, nil
}

func (t *AnnotationTEI) prepareTEIRequest(datasetID string, annotationID string, features []string) (*annotation.Annotation, []string, []*feature.Feature, error) {
	ann, err := t.annotationSvc.getWithoutFallback(datasetID, annotationID)
	if err != nil {
		return nil, nil, nil, err
	}
	if !t.annotationCanBeRepresentedAsTEI(ann) {
		return nil, nil, nil, annotationNotRepresentableAsTEIError{annotationID: ann.ID}
	}

	feats, err := t.featureSvc.ListFeatures(feature.NewDatasetDefScope(datasetID), []feature.ExpandOptions{feature.ExpandRevisions})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to list features for annotation %s: %v", ann.ID, err)
	}
	if len(features) == 0 {
		features = lo.Map(feats, func(f *feature.Feature, _ int) string { return f.ID })
	}

	return ann, features, feats, nil
}

func (t *AnnotationTEI) getScaledTEIWithFeatures(ann *annotation.Annotation, pageNumOrKey string, features []string, fallbackToOrigin bool, feats []*feature.Feature, imageVariant model.ImageVariant) (*teimodel.TEI, error) {
	tei, err := t.getTEIWithFeatures(ann, pageNumOrKey, features, fallbackToOrigin, feats)
	if err != nil {
		return nil, fmt.Errorf("failed to get TEI for annotation %s: %v", ann.ID, err)
	}
	if err := t.scaleTEIToVariant(ann.DatasetID, pageNumOrKey, tei, imageVariant); err != nil {
		return nil, fmt.Errorf("failed to scale TEI for annotation %s page %s: %v", ann.ID, pageNumOrKey, err)
	}
	return tei, nil
}

func (t *AnnotationTEI) GetTxt(datasetID string, annotationID string, pageNumOrKey string, imprintOnly *bool) (string, error) {
	ann, err := t.annotationSvc.getWithoutFallback(datasetID, annotationID)
	if err != nil {
		return "", err
	}

	if !ann.Ocred {
		return "", fmt.Errorf("annotation %s is not OCRed", ann.ID)
	}

	// Annotation-level transcriptions override the annotation's original ALTO.
	if a, _, err := t.fileSysMgt.RetrieveAnnotationTranscriptionAltoPage(ann, pageNumOrKey); err == nil {
		return alto.ExtractTextContentsFromAlto(a), nil
	}
	if md, err := t.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, pageNumOrKey); err == nil {
		return md.Content, nil
	}
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		return alto.ExtractTextContentsFromAlto(a), nil
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
	feats, err := t.featureSvc.ListFeatures(feature.NewDatasetDefScope(ann.DatasetID), []feature.ExpandOptions{feature.ExpandRevisions})
	if err != nil {
		return nil, fmt.Errorf("failed to list features for annotation %s: %v", ann.ID, err)
	}
	return t.getTEIWithFeatures(ann, pageNumOrKey, features, fallbackToOrigin, feats)
}

func (t *AnnotationTEI) getTEIWithFeatures(ann *annotation.Annotation, pageNumOrKey string, features []string, fallbackToOrigin bool, feats []*feature.Feature) (*teimodel.TEI, error) {
	if files, listErr := t.fileSysMgt.ListAnnotationTranscriptionFiles(ann, pageNumOrKey); listErr == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		return buildTEIFromTranscriptionFiles(pageNumOrKey, files, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Annotation-level transcriptions override the annotation's original ALTO.
	// Within the transcription directory, ALTO is preferred over Markdown.
	if a, _, err := t.fileSysMgt.RetrieveAnnotationTranscriptionAltoPage(ann, pageNumOrKey); err == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}
	if md, err := t.fileSysMgt.RetrieveAnnotationMarkdownPage(ann, pageNumOrKey); err == nil {
		return tei2.BuildTEIFromMarkdown(pageNumOrKey, md, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Fall back to the ALTO page created in the platform.
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		imageURL := path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))
		// todo: support items in alto
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, nil, imageURL, t.getBibleMetadata(ann.DatasetID, pageNumOrKey))
	}

	// Text
	var (
		lines        []string
		translations map[string][]string
		imageURL     string
		err          error
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
	results, err := t.resultSvc.ListResults(feature.NewDatasetExecScope(ann.DatasetID, ann.ID), []string{pageNumOrKey}, features, fallbackToOrigin)
	if err != nil {
		return nil, fmt.Errorf("failed to list results for annotation %s: %v", ann.ID, err)
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
	ann, features, feats, err := t.prepareTEIRequest(datasetID, annotationID, features)
	if err != nil {
		return nil, err
	}

	teis := map[string]*teimodel.TEI{}
	for _, key := range keys {
		tei, err := t.getScaledTEIWithFeatures(ann, key, features, fallbackToOrigin, feats, imageVariant)
		if err != nil {
			return nil, err
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
