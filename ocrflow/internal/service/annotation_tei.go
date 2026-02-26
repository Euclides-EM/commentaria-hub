package service

import (
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/futils"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/pagesparser"
	tei2 "github.com/MiaMish/elements-dh/ocrflow/pkg/tei"
	teimodel "github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
	"github.com/samber/lo"
)

type AnnotationTEI struct {
	annotationSvc     *Annotation
	fileSysMgt        *filesys.Manager
	resultSvc         *Result
	featureSvc        *Feature
	tpsTranscriptions *store.TPSTranscriptions
}

func NewAnnotationTEI(annotationSvc *Annotation, fileSysMgt *filesys.Manager, resultSvc *Result, featureSvc *Feature, tpsTranscriptions *store.TPSTranscriptions) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc:     annotationSvc,
		fileSysMgt:        fileSysMgt,
		resultSvc:         resultSvc,
		featureSvc:        featureSvc,
		tpsTranscriptions: tpsTranscriptions,
	}
}

func (t *AnnotationTEI) GetTEI(datasetID string, annotationID string, pageNumOrKey string, features []string) ([]byte, error) {
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

	tei, err := t.getTEI(ann, pageNumOrKey, features)
	if err != nil {
		return nil, fmt.Errorf("failed to get TEI for annotation %s: %v", ann.ID, err)
	}

	xml, err := tei.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize TEI to XML for annotation %s: %v", ann.ID, err)
	}

	return xml, nil
}

func (t *AnnotationTEI) getTEI(ann *annotation.Annotation, pageNumOrKey string, features []string) (*teimodel.TEI, error) {
	// Load results + feature definitions first (shared by both ALTO and TXT paths)
	results, err := t.resultSvc.ListResults(ann.DatasetID, ann.ID, []string{pageNumOrKey}, features)
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

	if ann.DatasetID == "tps" && ann.ID == "ann_1" {
		transcription, enTranslation, err := t.tpsTranscriptions.Get(pageNumOrKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get TPS transcription for key %s: %v", pageNumOrKey, err)
		}
		lines = strings.Split(transcription, "\n")
		translations = map[string][]string{
			"en": strings.Split(enTranslation, "\n"),
		}
		imageURL = futils.LocateFileInDir(
			t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID),
			func(filename string) bool {
				return strings.HasPrefix(filename, pageNumOrKey)
			},
		)
		if imageURL == "" {
			log.Printf("warning: failed to locate image for TPS page %s in dataset %s", pageNumOrKey, ann.DatasetID)
		}
	} else {
		var err error
		lines, translations, err = t.fileSysMgt.RetrieveAnnotationTXTPage(ann, pageNumOrKey)
		if err != nil {
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

func buildItems(results []*feature.Result, feats []*feature.Feature, transcription []string) []tei2.EntityItem {
	findFeat := func(id string) *feature.Feature {
		return lo.FindOrElse(feats, nil, func(f *feature.Feature) bool { return f.ID == id })
	}

	var items []tei2.EntityItem
	for _, res := range results {
		feat := findFeat(res.Feature)
		if feat == nil {
			log.Printf("warning: feature %s not found for result %s, skipping", res.Feature, res.ID)
			continue
		}

		for _, val := range res.Values {
			// todo: use https://gist.github.com/ReallyLiri/661e03381ed4f3aede6aa2d9f20fefef to normalize surface forms for better matching, instead of just lowercasing and removing newlines.
			normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(val.Surface, "¬", ""), "\n", " "))

			props := map[string]string{
				"value": normalized,
			}
			if strings.TrimSpace(val.Surface) != "" {
				props["surface"] = val.Surface
			}

			for k, v := range val.Properties {
				props[k] = v
			}

			fullContent := strings.Join(transcription, "\n")
			startIndex := strings.Index(fullContent, val.Surface)
			if startIndex == -1 {
				log.Printf("warning: surface form '%s' not found in transcription for result %s, skipping", val.Surface, res.ID)
				continue
			}
			startAtLineNum := 0
			startAtCharCount := 0
			for i, line := range transcription {
				if startAtCharCount+len(line)+1 > startIndex { // +1 for the newline that was removed in fullContent
					startAtLineNum = i
					break
				}
				startAtCharCount += len(line) + 1 // +1 for the newline
			}
			endAtLineNum := startAtLineNum
			endAtCharCount := startAtCharCount
			for i := startAtLineNum; i < len(transcription); i++ {
				line := transcription[i]
				if endAtCharCount+len(line)+1 > startIndex+len(val.Surface) {
					endAtLineNum = i
					break
				}
				endAtCharCount += len(line) + 1 // +1 for the newline
			}

			items = append(items, tei2.EntityItem{
				Start: tei2.EntityLocationIndex{
					BlockID:    "1",
					LineID:     fmt.Sprintf("%d", startAtLineNum),
					ByteOffset: startAtCharCount,
				},
				End: tei2.EntityLocationIndex{
					BlockID:    "1",
					LineID:     fmt.Sprintf("%d", endAtLineNum),
					ByteOffset: endAtCharCount,
				},
				Category:   feat.Name,
				Properties: props,
			})
		}
	}
	return items
}
