package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/normalize"
	"github.com/samber/lo"
)

func (fe *Execution) annotationApplyFunc(exec *feature.Execution, key string, actions *executionActions) applyFunc {
	return func() ([]*feature.Result, error) {
		textLanguages, err := fe.languageResolver.Resolve(exec.Scope.DatasetID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get text language for key %s: %w", key, err)
		}
		textLanguage := strings.Join(textLanguages, " and ")
		fullText, err := fe.annotationTEISvc.GetTxt(exec.Scope.DatasetID, exec.Scope.AnnotationID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get full text for annotation %s and key %s: %w", exec.Scope.AnnotationID, key, err)
		}
		fullText = strings.TrimSpace(fullText)
		if fullText == "" {
			return nil, fmt.Errorf("full text is empty for annotation %s and key %s", exec.Scope.AnnotationID, key)
		}

		ann, err := fe.annotationSvc.Get(exec.Scope.DatasetID, exec.Scope.AnnotationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get annotation for dataset %s and annotation %s: %w", exec.Scope.DatasetID, exec.Scope.AnnotationID, err)
		}

		results := make([]*feature.Result, 0)
		var execErrs []error
		if len(actions.categorizerRevisions) > 0 {
			categorizerResults, err := fe.annotationCategorizeApplyFunc(ann, key, actions.categorizerRevisions, actions.categorizerFeatures, exec.ID, fullText)()
			if err != nil {
				execErrs = append(execErrs, err)
			}
			results = append(results, categorizerResults...)
		}
		if len(actions.promptRevisions) > 0 {
			for _, group := range groupPromptRevisionsByAIConfig(actions.promptRevisions, actions.promptFeatures) {
				promptResults, err := fe.annotationPromptApplyFunc(ann, key, group.revisions, group.features, exec.ID, textLanguage, fullText)()
				if err != nil {
					execErrs = append(execErrs, err)
				}
				results = append(results, promptResults...)
			}
		}
		return results, errors.Join(execErrs...)
	}
}

func (fe *Execution) annotationPromptApplyFunc(ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string, textLanguage string, fullText string) applyFunc {
	return func() ([]*feature.Result, error) {
		if strings.TrimSpace(fullText) == "" {
			return nil, fmt.Errorf("full text is empty for dataset %s and key %s", ann.DatasetID, key)
		}
		aiProvider := frs[0].AIProvider
		aiModel := frs[0].AIModel
		featureNameToIndex, definitions, outputFormat := buildPromptComponents(frs, fes)
		prompt := fmt.Sprintf(`You are an AI agent designed to extract structured metadata from title pages of early modern European textbooks.

You will be given:
- The transcribed text of a title page in %s.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.
Each field should contain the exact quoted text(s) from the input, with no modifications, rephrasing, or interpretation. Include the original whitespaces, line breaks and punctuation as they appear in the transcription.
Some text may apply to more than one field, so you may return the same text portions in multiple fields if applicable.

Return only a valid JSON. Do not include any other output.

Output format:
{
  %s
}

Definitions:
%s

Transcribed text:
%s
`, textLanguage, outputFormat, strings.Join(definitions, "\n"), fullText)

		contextDesc := fmt.Sprintf("dataset %s and key %s", ann.DatasetID, key)
		rawResponse, err := fe.llmClient.Exec(aiProvider.ToLLMAIProvider(), aiModel, prompt, "")
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for %s using %s/%s: %w", contextDesc, aiProvider, aiModel, err)
		}
		rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM response for %s: %w", contextDesc, err)
		}
		return parseLLMResults(rawFields, frs, fes, featureNameToIndex, execID, feature.NewDatasetExecScope(ann.DatasetID, ann.ID), key, contextDesc)
	}
}

func (fe *Execution) annotationCategorizeApplyFunc(ann *annotation.Annotation, key string, revisions []*feature.Revision, features []*feature.Feature, id string, text string) applyFunc {
	return func() ([]*feature.Result, error) {
		if strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("full text is empty for dataset %s and key %s", ann.DatasetID, key)
		}
		results := make([]*feature.Result, 0)
		var execErrs []error
		for i, rev := range revisions {
			vals, err := fe.featurePropertySvc.CalcValsByPropertyKey(text, rev.Categorizer)
			if err != nil {
				execErrs = append(execErrs, fmt.Errorf("failed to calculate feature property for dataset %s and key %s: %w", ann.DatasetID, key, err))
				continue
			}
			source := feature.ResultSource{
				Resp:     "auto",
				Id:       id,
				Revision: rev.ID,
				Name:     "categorizer",
			}
			res := &feature.Result{
				Scope:     feature.NewDatasetExecScope(ann.DatasetID, ann.ID),
				FeatureID: features[i].ID,
				Key:       key,
				Source:    source,
				Values: lo.Map(vals, func(v normalize.MappedOriginal, _ int) feature.ResultValue {
					return feature.ResultValue{
						Surface: v.Original,
						Properties: map[string]string{
							"normalized": v.Mapped,
						},
					}
				}),
			}
			results = append(results, res)
		}
		return results, errors.Join(execErrs...)
	}
}
