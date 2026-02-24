package service

import (
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	tei2 "github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/formatcov/tei/model"
)

type AnnotationTEI struct {
	annotationSvc     *Annotation
	fileSysMgt        *filesys.Manager
	resultSvc         *Result
	tpsTranscriptions *store.TPSTranscriptions
}

func NewAnnotationTEI(annotationSvc *Annotation, fileSysMgt *filesys.Manager, resultSvc *Result, tpsTranscriptions *store.TPSTranscriptions) *AnnotationTEI {
	return &AnnotationTEI{
		annotationSvc:     annotationSvc,
		fileSysMgt:        fileSysMgt,
		resultSvc:         resultSvc,
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

	// a *alto.Alto, entities EntitiesInput, profiles ProfilesInput, imageUrls []string

	// Parse features filter (supports repeated `features` params and comma-separated lists)
	var featureFilter []string
	for _, raw := range features {
		if raw == "" {
			continue
		}
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				featureFilter = append(featureFilter, trimmed)
			}
		}
	}

	// Get feature results for this key
	//results, err := t.resultSvc.ListResults(datasetID, annotationID, []string{pageNumOrKey}, featureFilter)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to get feature results: %w", err)
	//}
	tei, err := t.getTEI(ann, pageNumOrKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get TEI for annotation %s: %v", ann.ID, err)
	}

	xml, err := tei.ToXML()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize TEI to XML for annotation %s: %v", ann.ID, err)
	}

	return xml, nil
}

func (t *AnnotationTEI) getTEI(ann *annotation.Annotation, pageNumOrKey string) (*model.TEI, error) {
	a, _, err := t.fileSysMgt.RetrieveAnnotationAltoPage(ann, pageNumOrKey)
	if err == nil {
		return tei2.BuildTEIFromALTO(a, nil, "")
	}

	var lines []string
	var translations map[string][]string
	if ann.DatasetID == "tps" && ann.ID == "ann_1" {
		transcription, enTranslation, err := t.tpsTranscriptions.Get(pageNumOrKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get TPS transcription for key %s: %v", pageNumOrKey, err)
		}
		lines = strings.Split(transcription, "\n")
		translations = map[string][]string{
			"en": strings.Split(enTranslation, "\n"),
		}
	} else {
		lines, translations, err = t.fileSysMgt.RetrieveAnnotationTXTPage(ann, pageNumOrKey)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve TXT page for annotation %s: %v", ann.ID, err)
		}
	}

	linesInput := tei2.LinesInput{
		LinesByKeys: map[string]tei2.Lines{
			pageNumOrKey: {
				TranscriptionLines: lines,
				Translations:       translations,
			},
		},
	}
	return tei2.BuildTEIFromLines(linesInput, nil, nil)
}
