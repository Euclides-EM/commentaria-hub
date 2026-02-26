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

	// Helper: feature lookup
	findFeat := func(id string) *feature.Feature {
		return lo.FindOrElse(feats, nil, func(f *feature.Feature) bool { return f.ID == id })
	}

	// Build EntityItems (mentions + profile rows)
	buildItems := func() []tei2.EntityItem {
		var items []tei2.EntityItem
		for _, res := range results {
			feat := findFeat(res.Feature)
			if feat == nil {
				log.Printf("warning: feature %s not found for result %s, skipping", res.Feature, res.ID)
				continue
			}

			for _, val := range res.Values {
				normalized := lo.Ternary(val.Normalized != "", val.Normalized, strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(val.Surface, "¬", ""), "\n", " ")))
				entRef := fmt.Sprintf("#ent_%s", normalized)

				// --- Mention row (only if location looks usable) ---
				// For ALTO builder: Location.TextBlockID/TextLineID should be ALTO ids.
				// For Lines builder: Location might be empty or not match synthetic l%04d. If empty, we skip mention.
				if val.Location.TextLineID != "" && val.Location.CharactersSpan.End > val.Location.CharactersSpan.Start {
					// @ana must point at the feature taxonomy category (#feat_origin_language etc.) so spans and facts are self-describing.
					ana := "#" + tei2.FeatureCategoryID(feat.Name)
					if ana == "#" {
						ana = ""
					}
					items = append(items, tei2.EntityItem{
						Ref: entRef,
						Start: tei2.EntityLocationIndex{
							PageID:     pageNumOrKey,
							BlockID:    val.Location.TextBlockID,
							LineID:     val.Location.TextLineID,
							ByteOffset: val.Location.CharactersSpan.Start,
						},
						End: tei2.EntityLocationIndex{
							PageID:     pageNumOrKey,
							BlockID:    val.Location.TextBlockID,
							LineID:     val.Location.TextLineID,
							ByteOffset: val.Location.CharactersSpan.End,
						},
						Ana: ana,
					})
				}

				// --- Profile rows (dedup happens in TEI builder) ---
				items = append(items, tei2.EntityItem{
					Ref:   entRef,
					Type:  "feature_name",
					Value: feat.Name,
				})
				items = append(items, tei2.EntityItem{
					Ref:   entRef,
					Type:  "value",
					Value: normalized,
				})

				// If you want surface too (often useful):
				if strings.TrimSpace(val.Surface) != "" {
					items = append(items, tei2.EntityItem{
						Ref:   entRef,
						Type:  "surface",
						Value: val.Surface,
					})
				}

				// If you want to keep ProfileID as a profile field
				if strings.TrimSpace(val.ProfileID) != "" {
					items = append(items, tei2.EntityItem{
						Ref:   entRef,
						Type:  "profile_id",
						Value: val.ProfileID,
					})
				}
			}
		}
		return items
	}

	items := buildItems()

	// 1) Try ALTO page first
	if a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey); err == nil {
		// Best-effort image url
		imageURL := ""
		// If you store images by dataset+key:
		imageURL = path.Join(t.fileSysMgt.DatasetImagesDirByID(ann.DatasetID), pagesparser.PageOrKeyToPNGFilename(pageNumOrKey))

		// Build from ALTO (mentions must have Location.TextBlockID/TextLineID aligned to ALTO ids)
		return tei2.BuildTEIFromALTO(pageNumOrKey, a, items, imageURL)
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

	// IMPORTANT: For the Lines builder, mentions must use BlockID="b1" and LineID="l%04d".
	// Your results may not have these IDs in Location. If they do not, the mention rows above get skipped.
	// Profiles still get emitted, which is safe.
	return tei2.BuildTEIFromLines(pageNumOrKey, pageLines, items, imageURL)
}
