package service

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotationrule"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store/filesys"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/alto"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/llm"
	"github.com/stretchr/testify/require"
)

type transcriptionAnnotationStore struct {
	annotations map[string]*annotation.Annotation
}

func (s *transcriptionAnnotationStore) GetAnnotation(datasetID, id string) (*annotation.Annotation, error) {
	ann := s.annotations[id]
	if ann == nil || ann.DatasetID != datasetID {
		return nil, nil
	}
	return ann, nil
}

type transcriptionDatasetStore struct {
	dataset *model.Dataset
}

func (s *transcriptionDatasetStore) GetDataset(id string) (*model.Dataset, error) {
	if s.dataset == nil || s.dataset.ID != id {
		return nil, nil
	}
	return s.dataset, nil
}

type transcriptionLLMCall struct {
	prompt       llm.Prompt
	image, label string
}

type transcriptionLLM struct {
	calls []transcriptionLLMCall
}

func (l *transcriptionLLM) ExecPromptResultWithLogLabel(_, _ string, prompt llm.Prompt, attachmentPath, logLabel string) (llm.Result, error) {
	l.calls = append(l.calls, transcriptionLLMCall{prompt: prompt, image: attachmentPath, label: logLabel})
	cost := 0.01
	return llm.Result{
		Text: fmt.Sprintf("corrected round %d", len(l.calls)),
		Usage: llm.Usage{
			InputTokens: 10, OutputTokens: 2, TotalTokens: 12, CostUSD: &cost,
		},
	}, nil
}

func TestApplyLLMTranscriptionCorrectorUsesAnnotationAndEditionInputs(t *testing.T) {
	root := t.TempDir()
	manager := filesys.NewFileSystemManager(root, filepath.Join(root, "models"), filepath.Join(root, "diagrams"), filepath.Join(root, "defaults"))
	target := &annotation.Annotation{DatasetID: "dataset", Ocred: true, Pages: "1"}
	target.ID = "target"
	additional := &annotation.Annotation{DatasetID: "dataset", Ocred: true, Pages: "1"}
	additional.ID = "additional"
	for _, ann := range []*annotation.Annotation{target, additional} {
		dir := manager.DatasetAnnotationAltoDir(ann)
		require.NoError(t, os.MkdirAll(dir, 0o755))
		require.NoError(t, alto.SaveToFile(testCorrectionALTO(ann.ID), filepath.Join(dir, "page-0001.xml")))
	}
	imagesDir := manager.DatasetImagesDirByID("dataset")
	require.NoError(t, os.MkdirAll(imagesDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(imagesDir, "page-0001.png"), []byte("image"), 0o644))
	editionPageDir := manager.EditionTxtPageTranscriptionDir("edition", "1")
	require.NoError(t, os.MkdirAll(editionPageDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(editionPageDir, "original.md"), []byte("edition transcription\n"), 0o644))

	llmClient := &transcriptionLLM{}
	applier := NewAnnotationRuleApplier(
		nil,
		manager,
		"",
		nil,
		&transcriptionAnnotationStore{annotations: map[string]*annotation.Annotation{"additional": additional}},
		&transcriptionDatasetStore{dataset: &model.Dataset{EditionID: "edition"}},
		llmClient,
	)
	applier.datasetStore.(*transcriptionDatasetStore).dataset.ID = "dataset"
	rule := annotationrule.NewLLMTranscriptionCorrector("ollama", "vision", []string{"additional"}, true)
	rule.Rounds = 3

	got, err := applier.applyLLMTranscriptionCorrector(imagesDir, target, rule)
	require.NoError(t, err)
	require.Same(t, target, got)
	require.Len(t, llmClient.calls, 3)
	require.Contains(t, llmClient.calls[0].prompt.Dynamic, "target")
	require.Contains(t, llmClient.calls[0].prompt.Dynamic, "additional")
	require.Contains(t, llmClient.calls[0].prompt.Dynamic, "edition transcription")
	require.Equal(t, filepath.Join(imagesDir, "page-0001.png"), llmClient.calls[0].image)
	require.FileExists(t, filepath.Join(manager.AnnotationTxtTranscriptionDir(target), "page-0001", "round-01.md"))
	final, err := os.ReadFile(filepath.Join(manager.AnnotationTxtTranscriptionDir(target), "page-0001", "original.md"))
	require.NoError(t, err)
	require.Equal(t, "corrected round 3\n", string(final))
	require.NotNil(t, rule.Usage)
	require.EqualValues(t, 30, rule.Usage.InputTokens)
	require.EqualValues(t, 6, rule.Usage.OutputTokens)
	require.EqualValues(t, 36, rule.Usage.TotalTokens)
	require.NotNil(t, rule.Usage.CostUSD)
	require.InDelta(t, 0.03, *rule.Usage.CostUSD, 0.0000001)
}

func TestApplyLLMTranscriptionCorrectorRejectsNonOCRAnnotation(t *testing.T) {
	ann := &annotation.Annotation{Ocred: false}
	ann.ID = "not-ocr"
	applier := &AnnotationRuleApplier{}

	_, err := applier.applyLLMTranscriptionCorrector("images", ann, annotationrule.NewLLMTranscriptionCorrector("ollama", "vision", nil, false))
	require.ErrorContains(t, err, "only applicable to OCR-ed annotations")
}

func TestApplyLLMTranscriptionCorrectorLimitsPages(t *testing.T) {
	root := t.TempDir()
	manager := filesys.NewFileSystemManager(root, filepath.Join(root, "models"), filepath.Join(root, "diagrams"), filepath.Join(root, "defaults"))
	target := &annotation.Annotation{DatasetID: "dataset", Ocred: true, Pages: "1-3"}
	target.ID = "target"
	altoDir := manager.DatasetAnnotationAltoDir(target)
	require.NoError(t, os.MkdirAll(altoDir, 0o755))
	imagesDir := manager.DatasetImagesDirByID("dataset")
	require.NoError(t, os.MkdirAll(imagesDir, 0o755))
	for _, page := range []int{1, 2, 3} {
		pageName := fmt.Sprintf("page-%04d", page)
		require.NoError(t, alto.SaveToFile(testCorrectionALTO(pageName), filepath.Join(altoDir, pageName+".xml")))
		require.NoError(t, os.WriteFile(filepath.Join(imagesDir, pageName+".png"), []byte("image"), 0o644))
	}

	llmClient := &transcriptionLLM{}
	applier := NewAnnotationRuleApplier(nil, manager, "", nil, nil, nil, llmClient)
	rule := annotationrule.NewLLMTranscriptionCorrector("ollama", "vision", nil, false)
	rule.Pages = "2-3"

	_, err := applier.applyLLMTranscriptionCorrector(imagesDir, target, rule)
	require.NoError(t, err)
	require.Len(t, llmClient.calls, 2)
	require.Equal(t, filepath.Join(imagesDir, "page-0002.png"), llmClient.calls[0].image)
	require.Equal(t, filepath.Join(imagesDir, "page-0003.png"), llmClient.calls[1].image)
}

func TestApplyLLMTranscriptionCorrectorRejectsPagesOutsideAnnotation(t *testing.T) {
	target := &annotation.Annotation{DatasetID: "dataset", Ocred: true, Pages: "1-3"}
	target.ID = "target"
	applier := &AnnotationRuleApplier{transcriptionLLM: &transcriptionLLM{}}
	rule := annotationrule.NewLLMTranscriptionCorrector("ollama", "vision", nil, false)
	rule.Pages = "4"

	_, err := applier.applyLLMTranscriptionCorrector("images", target, rule)
	require.ErrorContains(t, err, "page 4 is not in annotation target")
}

func TestApplyLLMTranscriptionCorrectorRejectsNonOCRAdditionalAnnotation(t *testing.T) {
	target := &annotation.Annotation{DatasetID: "dataset", Ocred: true}
	target.ID = "target"
	additional := &annotation.Annotation{DatasetID: "dataset", Ocred: false}
	additional.ID = "additional"
	applier := &AnnotationRuleApplier{
		fileSysMgt:       filesys.NewFileSystemManager(t.TempDir(), "", "", ""),
		annotationStore:  &transcriptionAnnotationStore{annotations: map[string]*annotation.Annotation{"additional": additional}},
		transcriptionLLM: &transcriptionLLM{},
	}

	_, err := applier.applyLLMTranscriptionCorrector("images", target, annotationrule.NewLLMTranscriptionCorrector("ollama", "vision", []string{"additional"}, false))
	require.ErrorContains(t, err, "additional annotation additional is not OCR-ed")
}

func testCorrectionALTO(text string) *alto.Alto {
	return &alto.Alto{Layout: alto.Layout{Page: []alto.Page{{PrintSpace: alto.PrintSpace{TextBlocks: []alto.TextBlock{{
		Lines: []alto.TextLine{{Strings: []alto.String{{Content: text}}}},
	}}}}}}}
}
