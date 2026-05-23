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

type featureGroup struct {
	revisions []*feature.Revision
	features  []*feature.Feature
}

func partitionFeatures(features []*feature.Feature, revisions []*feature.Revision, pred func(*feature.Feature) bool) (match, rest featureGroup) {
	for i, f := range features {
		if pred(f) {
			match.revisions = append(match.revisions, revisions[i])
			match.features = append(match.features, f)
		} else {
			rest.revisions = append(rest.revisions, revisions[i])
			rest.features = append(rest.features, f)
		}
	}
	return
}

func (fe *Execution) runAnnotationGroup(ann *annotation.Annotation, key, execID, textLanguage, fullText string, catGroup, promptGroup featureGroup) ([]*feature.Result, error) {
	var results []*feature.Result
	var execErrs []error

	if len(catGroup.revisions) > 0 && fullText != "" {
		r, err := fe.annotationCategorizeApplyFunc(ann, key, catGroup.revisions, catGroup.features, execID, fullText)()
		if err != nil {
			execErrs = append(execErrs, err)
		}
		results = append(results, r...)
	}

	if len(promptGroup.revisions) > 0 && fullText != "" {
		for _, group := range groupPromptRevisionsByAIConfig(promptGroup.revisions, promptGroup.features) {
			r, err := fe.annotationPromptApplyFunc(ann, key, group.revisions, group.features, execID, textLanguage, fullText)()
			if err != nil {
				execErrs = append(execErrs, err)
			}
			results = append(results, r...)
		}
	}

	return results, errors.Join(execErrs...)
}

func (fe *Execution) annotationApplyFunc(exec *feature.Execution, key string, actions *executionActions) applyFunc {
	return func() ([]*feature.Result, error) {
		textLanguages, err := fe.languageResolver.Resolve(exec.Scope.DatasetID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get text language for key %s: %w", key, err)
		}
		textLanguage := strings.Join(textLanguages, " and ")

		ann, err := fe.annotationSvc.Get(exec.Scope.DatasetID, exec.Scope.AnnotationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get annotation for dataset %s and annotation %s: %w", exec.Scope.DatasetID, exec.Scope.AnnotationID, err)
		}

		isImprintFeat := func(f *feature.Feature) bool {
			return strings.Contains(strings.ToLower(f.Name), "in imprint")
		}

		catImprint, catNonImprint := partitionFeatures(actions.categorizerFeatures, actions.categorizerRevisions, isImprintFeat)
		promptImprint, promptNonImprint := partitionFeatures(actions.promptFeatures, actions.promptRevisions, isImprintFeat)

		needImprint := len(catImprint.features)+len(promptImprint.features) > 0
		needNonImprint := len(catNonImprint.features)+len(promptNonImprint.features) > 0

		var imprintFullText string
		if needImprint {
			text, err := fe.annotationTEISvc.GetTxt(exec.Scope.DatasetID, exec.Scope.AnnotationID, key, lo.ToPtr(true))
			if err != nil {
				return nil, fmt.Errorf("failed to get full text for annotation %s and key %s: %w", exec.Scope.AnnotationID, key, err)
			}
			imprintFullText = strings.TrimSpace(text)
		}

		var nonImprintFullText string
		if needNonImprint {
			text, err := fe.annotationTEISvc.GetTxt(exec.Scope.DatasetID, exec.Scope.AnnotationID, key, lo.ToPtr(false))
			if err != nil {
				return nil, fmt.Errorf("failed to get full text for annotation %s and key %s: %w", exec.Scope.AnnotationID, key, err)
			}
			nonImprintFullText = strings.TrimSpace(text)
		}

		var results []*feature.Result
		var execErrs []error

		r, err := fe.runAnnotationGroup(ann, key, exec.ID, textLanguage, imprintFullText, catImprint, promptImprint)
		if err != nil {
			execErrs = append(execErrs, err)
		}
		results = append(results, r...)

		r, err = fe.runAnnotationGroup(ann, key, exec.ID, textLanguage, nonImprintFullText, catNonImprint, promptNonImprint)
		if err != nil {
			execErrs = append(execErrs, err)
		}
		results = append(results, r...)

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
