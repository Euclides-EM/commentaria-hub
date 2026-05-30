package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/normalize"
	"github.com/samber/lo"
)

type featureGroup struct {
	revisions       []*feature.Revision
	features        []*feature.Feature
	textDescription string
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
			r, err := fe.annotationPromptApplyFunc(ann, key, group.revisions, group.features, execID, textLanguage, fullText, promptGroup.textDescription)()
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
		promptImprint.textDescription = "a title page"
		promptNonImprint.textDescription = "the imprint section of a title page"

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

func (fe *Execution) annotationPromptApplyFunc(ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string, textLanguage string, fullText string, textDescription string) applyFunc {
	return func() ([]*feature.Result, error) {
		if strings.TrimSpace(fullText) == "" {
			return nil, fmt.Errorf("full text is empty for dataset %s and key %s", ann.DatasetID, key)
		}
		activeRevisions := frs
		activeFeatures := fes
		resultsByFeatureID := make(map[string]*feature.Result, len(fes))
		for attempt := 0; attempt < 3 && len(activeRevisions) > 0; attempt++ {
			aiProvider := activeRevisions[0].AIProvider
			aiModel := activeRevisions[0].AIModel
			featureNameToIndex, definitions, outputFormat := buildPromptComponents(activeRevisions, activeFeatures)
			prompt := fmt.Sprintf(`You are an AI agent designed to extract structured metadata from title pages of early modern European textbooks.

You will be given:
- The transcribed text of %s in %s.

Your task is to extract specific paratextual features from the transcription and return them as a JSON object.

Extraction rules:
- Extract only the minimal text span that corresponds to the requested feature.
- Omit surrounding adjectives, descriptive phrases, function words, and punctuation unless they are intrinsically part of the feature itself.
- Preserve the original spelling, capitalization, whitespace, line breaks, and punctuation within the extracted span exactly as they appear in the transcription.
- Early modern orthography may differ from modern spelling and letter usage. For example, “v” and “u” or “i” and “j” may be interchangeable (“vpon,” “Iesus”), and other historical spellings may vary. Treat these as normal forms and reproduce the text exactly as written, without modernization or normalization.
- Words or phrases may be split across lines or interrupted by characters such as "-", "=" or similar separators. Interpret these as part of the transcription layout and extract the relevant text accurately.
- Some text may apply to more than one field, so the same text may appear in multiple fields if applicable.
- Do not normalize, modernize, interpret, or correct the text.

Return only a valid JSON object. Do not include explanations or any other output.

Output format:
{
  %s
}

Definitions:
%s

Transcribed text:
%s
`, textDescription, textLanguage, outputFormat, strings.Join(definitions, "\n"), fullText)

			contextDesc := fmt.Sprintf("dataset %s and key %s", ann.DatasetID, key)
			rawResponse, err := fe.llmClient.Exec(aiProvider.ToLLMAIProvider(), aiModel, prompt, "")
			if err != nil {
				return nil, fmt.Errorf("failed to execute LLM prompt for %s using %s/%s: %w", contextDesc, aiProvider, aiModel, err)
			}
			rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
			if err != nil {
				return nil, fmt.Errorf("failed to parse LLM response for %s: %w", contextDesc, err)
			}
			parsed, err := parseLLMResults(rawFields, activeRevisions, activeFeatures, featureNameToIndex, execID, feature.NewDatasetExecScope(ann.DatasetID, ann.ID), key, contextDesc, fullText, true)
			if err != nil {
				return nil, err
			}
			for _, result := range parsed.results {
				resultsByFeatureID[result.FeatureID] = result
			}
			if len(parsed.hallucinatedFeatureIDs) == 0 {
				break
			}
			nextRevisions := make([]*feature.Revision, 0, len(parsed.hallucinatedFeatureIDs))
			nextFeatures := make([]*feature.Feature, 0, len(parsed.hallucinatedFeatureIDs))
			for i, featureItem := range activeFeatures {
				if slices.Contains(parsed.hallucinatedFeatureIDs, featureItem.ID) {
					nextRevisions = append(nextRevisions, activeRevisions[i])
					nextFeatures = append(nextFeatures, activeFeatures[i])
				}
			}
			activeRevisions = nextRevisions
			activeFeatures = nextFeatures
		}

		results := make([]*feature.Result, 0, len(fes))
		for _, featureItem := range fes {
			if result, ok := resultsByFeatureID[featureItem.ID]; ok {
				results = append(results, result)
			}
		}
		return results, nil
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
