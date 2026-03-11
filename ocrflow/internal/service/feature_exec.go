package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/features"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/normalize"
	"github.com/samber/lo"
)

const featureExecutionWorkerCount = 8

type Execution struct {
	featureRevisionsSvc *Revision
	featuresSvc         *Feature
	featureResultsSvc   *Result
	annotationSvc       *Annotation
	annotationTEISvc    *AnnotationTEI
	languageResolver    *LanguagesResolver
	featurePropertySvc  *FeatureProperty
	store               *fpstore.FeatureExecutionStore
	filesysManager      *filesys.Manager
	datasetImg          *DatasetImg
	llmClient           *llm.Client
	mu                  sync.Mutex // Protects status updates in goroutines
}

// NewExecution returns a new Execution service using the given store (e.g. *storefeatureplat.FeatureExecutionStore).
func NewExecution(featureRevisionsSvc *Revision, featuresSvc *Feature, featureResultsSvc *Result, annotationSvc *Annotation, annotationTEISvc *AnnotationTEI, languageResolver *LanguagesResolver, featurePropertySvc *FeatureProperty, store *fpstore.FeatureExecutionStore, filesysManager *filesys.Manager, datasetImg *DatasetImg, llmClient *llm.Client) *Execution {
	return &Execution{
		featureRevisionsSvc: featureRevisionsSvc,
		featuresSvc:         featuresSvc,
		featureResultsSvc:   featureResultsSvc,
		annotationSvc:       annotationSvc,
		annotationTEISvc:    annotationTEISvc,
		languageResolver:    languageResolver,
		featurePropertySvc:  featurePropertySvc,
		store:               store,
		filesysManager:      filesysManager,
		datasetImg:          datasetImg,
		llmClient:           llmClient,
	}
}

func (fe *Execution) ListFeatureExecutions(datasetID string, featureIds []string, statuses []feature.ExecutionStatus) ([]*feature.Execution, error) {
	res, err := fe.store.List(datasetID, featureIds, statuses)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(res, func(a, b *feature.Execution) int {
		return b.UpdatedAt.Compare(a.UpdatedAt)
	})
	return res, nil
}

func (fe *Execution) GetFeatureExecution(executionId string) (*feature.Execution, error) {
	return fe.store.GetByID(executionId)
}

func (fe *Execution) CreateFeatureExecution(exec *feature.Execution) (*feature.Execution, error) {
	ann, err := fe.annotationSvc.Get(exec.DatasetID, exec.AnnotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get annotation for dataset %s and annotation %s: %w", exec.DatasetID, exec.AnnotationID, err)
	}
	exec.ID = idgen.GenerateID("exec")
	exec.Status = feature.ExecutionStatusInProgress
	exec.StatusReason = ""

	skipState, err := fe.loadExecutionSkipState(exec)
	if err != nil {
		return nil, err
	}

	var applyFuncs []func() ([]*feature.Result, error)
	for _, key := range exec.Keys {
		var promptRevisions []*feature.Revision
		var promptFeatures []*feature.Feature
		var categorizerRevisions []*feature.Revision
		var categorizerFeatures []*feature.Feature

		for _, item := range exec.Apply {
			feat, err := fe.featuresSvc.GetFeature(exec.DatasetID, item.Feature, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to get feature for feature %s: %w", item.Feature, err)
			}
			fr, err := fe.featureRevisionsSvc.GetFeatureRevision(exec.DatasetID, item.Feature, item.Revision)
			if err != nil {
				return nil, fmt.Errorf("failed to get feature revision for feature %s and revision %s: %w", item.Feature, item.Revision, err)
			}
			skipReasons := skipState.SkipReasons(exec.Policy, key, item.Feature, item.Revision)
			if len(skipReasons) > 0 {
				log.Printf("skipping execution %s key %s feature %s revision %s due to skip policy: %s", exec.ID, key, item.Feature, item.Revision, strings.Join(lo.Map(skipReasons, func(reason feature.ExecutionSkipIf, _ int) string {
					return string(reason)
				}), ", "))
				continue
			}
			if fr.Categorizer != "" {
				categorizerRevisions = append(categorizerRevisions, fr)
				categorizerFeatures = append(categorizerFeatures, feat)
			} else if fr.Prompt != "" {
				promptRevisions = append(promptRevisions, fr)
				promptFeatures = append(promptFeatures, feat)
			} else {
				return nil, fmt.Errorf("feature revision for feature %s and revision %s does not have a valid execution strategy", item.Feature, item.Revision)
			}
		}
		if len(categorizerRevisions) == 0 && len(promptRevisions) == 0 {
			log.Printf("skipping execution %s key %s because all actions were skipped by policy", exec.ID, key)
			continue
		}

		textLanguages, err := fe.languageResolver.Resolve(exec.DatasetID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get text language for key %s: %w", key, err)
		}
		textLanguage := strings.Join(textLanguages, " and ")
		fullText, err := fe.annotationTEISvc.GetTxt(exec.DatasetID, exec.AnnotationID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get full text for annotation %s and key %s: %w", exec.AnnotationID, key, err)
		}
		fullText = strings.TrimSpace(fullText)
		if fullText == "" {
			return nil, fmt.Errorf("full text is empty for annotation %s and key %s", exec.AnnotationID, key)
		}

		applyFuncs = append(applyFuncs, func() ([]*feature.Result, error) {
			results := make([]*feature.Result, 0)
			var execErrs []error
			if len(categorizerRevisions) > 0 {
				categorizerResults, err := fe.execCategorizer(ann, key, categorizerRevisions, categorizerFeatures, exec.ID, fullText)()
				if err != nil {
					execErrs = append(execErrs, err)
				}
				results = append(results, categorizerResults...)
			}
			if len(promptRevisions) > 0 {
				promptResults, err := fe.execPrompt(ann, key, promptRevisions, promptFeatures, exec.ID, textLanguage, fullText)()
				if err != nil {
					execErrs = append(execErrs, err)
				}
				results = append(results, promptResults...)
			}
			return results, errors.Join(execErrs...)
		})
	}

	if err := fe.store.Create(exec); err != nil {
		return nil, err
	}

	// Run all apply functions in parallel and wait for them to finish
	go func(executionID string, funcs []func() ([]*feature.Result, error)) {
		var wg sync.WaitGroup
		resultBatches := make([][]*feature.Result, len(funcs))
		errors := make([]error, len(funcs))
		type executionJob struct {
			index int
			fn    func() ([]*feature.Result, error)
		}
		jobs := make(chan executionJob)

		workerCount := min(featureExecutionWorkerCount, len(funcs))
		for range workerCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					result, err := job.fn()
					resultBatches[job.index] = result
					errors[job.index] = err
				}
			}()
		}
		for i, fn := range funcs {
			jobs <- executionJob{index: i, fn: fn}
		}
		close(jobs)

		wg.Wait()

		hasError := false
		for _, err := range errors {
			if err != nil {
				log.Printf("error in execution %s: %v\n", executionID, err)
				hasError = true
			}
		}

		newStatus := feature.ExecutionStatusSuccess
		statusReason := ""
		var results []*feature.Result
		for _, batch := range resultBatches {
			results = append(results, batch...)
		}
		if len(results) > 0 {
			err = fe.featureResultsSvc.CreateResults(results, lo.IfF(exec.Policy != nil, func() bool { return exec.Policy.PushToOrigin }).Else(false))
			if err != nil {
				log.Printf("failed to create result for execution %s: %v\n", executionID, err)
				newStatus = feature.ExecutionStatusFailed
				statusReason = "failed to store execution results"
			}
		}
		if hasError {
			newStatus = feature.ExecutionStatusFailed
			if statusReason == "" {
				statusReason = "one or more actions failed, check logs"
			}
		}
		if err := fe.store.UpdateStatus(executionID, newStatus, statusReason); err != nil {
			log.Printf("failed to update execution status for execution %s: %v\n", executionID, err)
		}

	}(exec.ID, applyFuncs)

	return exec, nil
}

func (fe *Execution) loadExecutionSkipState(exec *feature.Execution) (*features.ExecutionSkipState, error) {
	state := features.NewExecutionSkipState()
	if exec.Policy == nil || len(exec.Policy.SkipIf) == 0 || len(exec.Keys) == 0 || len(exec.Apply) == 0 {
		return state, nil
	}

	featureIDs := lo.Uniq(lo.Map(exec.Apply, func(item feature.ExecutionApplyItem, _ int) string {
		return item.Feature
	}))
	results, err := fe.featureResultsSvc.ListResultsForExecutionPolicy(
		exec.DatasetID,
		exec.AnnotationID,
		exec.Keys,
		featureIDs,
		exec.Policy.PushToOrigin,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to preload feature results for execution policy: %w", err)
	}

	for _, result := range results {
		if result == nil {
			continue
		}
		state.Add(result.FeatureID, result.PageKey, result.Source.Revision, result.Source.Resp == "human")
	}

	return state, nil
}

func (fe *Execution) CancelFeatureExecution(executionId string) (*feature.Execution, error) {
	// todo: currently, this is just a mock...

	exec, err := fe.store.GetByID(executionId)
	if err != nil {
		return nil, err
	}
	if exec.Status != feature.ExecutionStatusInProgress && exec.Status != feature.ExecutionStatusCanceling {
		return nil, fmt.Errorf("feature execution cannot be canceled as it is not in progress or already canceling")
	}
	if err := fe.store.UpdateStatus(executionId, feature.ExecutionStatusCanceling, "cancel request by user"); err != nil {
		return nil, err
	}
	exec.Status = feature.ExecutionStatusCanceling

	// Start a goroutine that will update the status from canceling to canceled after 30 seconds
	go func(executionID string) {
		time.Sleep(30 * time.Second)
		fe.mu.Lock()
		defer fe.mu.Unlock()
		// Check current status to ensure it's still canceling
		currentExec, err := fe.store.GetByID(executionID)
		if err != nil {
			return
		}
		// Only update if still in canceling state
		if currentExec.Status == feature.ExecutionStatusCanceling {
			_ = fe.store.UpdateStatus(executionID, feature.ExecutionStatusCanceled, "cancel request by user")
		}
	}(executionId)

	return exec, nil
}

func (fe *Execution) execPrompt(ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string, textLanguage string, fullText string) func() ([]*feature.Result, error) {
	return func() ([]*feature.Result, error) {
		if strings.TrimSpace(fullText) == "" {
			return nil, fmt.Errorf("full text is empty for dataset %s and key %s", ann.DatasetID, key)
		}
		dsPromptDesc := "historical title pages of translations of Euclid's Elements"
		dsPromptShortDesc := "title page"
		var definitions []string
		var outputFormat string
		featureNameToIndex := make(map[string]int)
		for i, _ := range frs {
			featureName := fmt.Sprintf("%s-rev-%s", fes[i].ID, frs[i].ID[0:3])
			featureNameToIndex[featureName] = i
			definitions = append(definitions, fmt.Sprintf("- %s: %s", featureName, frs[i].Prompt))
			if fes[i].IsList {
				outputFormat += fmt.Sprintf(`  "%s": [...], // zero or more quotes`+"\n", featureName)
			} else {
				outputFormat += fmt.Sprintf(`  "%s": "...", // a single quote or empty if not applicable`+"\n", featureName)
			}
		}
		outputFormat = strings.TrimSpace(outputFormat)
		prompt := fmt.Sprintf(`You are an AI agent designed to extract structured metadata from %s.

You will be given:
- The transcribed text of a %s in %s.

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
`, dsPromptDesc, dsPromptShortDesc, textLanguage, outputFormat, strings.Join(definitions, "\n"), fullText)

		rawResponse, err := fe.llmClient.Exec(prompt, "")
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for dataset %s and key %s: %w", ann.DatasetID, key, err)
		}
		rawFields, err := llm.ParseJSON[map[string]json.RawMessage](rawResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to parse LLM response for dataset %s and key %s: %w", ann.DatasetID, key, err)
		}

		var results []*feature.Result
		for fn, rawValue := range rawFields {
			idx, ok := featureNameToIndex[fn]
			if !ok {
				return nil, fmt.Errorf("llm response contained unknown field %q for dataset %s and key %s", fn, ann.DatasetID, key)
			}

			var quotes []string
			if fes[idx].IsList {
				if err := json.Unmarshal(rawValue, &quotes); err != nil {
					return nil, fmt.Errorf("failed to parse list response for field %q in dataset %s and key %s: %w", fn, ann.DatasetID, key, err)
				}
			} else {
				var quote string
				if err := json.Unmarshal(rawValue, &quote); err != nil {
					return nil, fmt.Errorf("failed to parse scalar response for field %q in dataset %s and key %s: %w", fn, ann.DatasetID, key, err)
				}
				if quote != "" {
					quotes = []string{quote}
				}
			}

			source := feature.ResultSource{
				Resp:     "auto",
				Id:       execID,
				Revision: frs[idx].ID,
				Name:     "llm",
			}
			res := &feature.Result{
				DatasetID:    ann.DatasetID,
				AnnotationID: ann.ID,
				FeatureID:    fes[idx].ID,
				PageKey:      key,
				Source:       source,
			}
			var resultValues []feature.ResultValue
			for _, quote := range quotes {
				resultValues = append(resultValues, feature.ResultValue{
					Surface: quote,
				})
			}
			res.Values = resultValues
			results = append(results, res)
		}
		return results, nil
	}
}

func (fe *Execution) execCategorizer(ann *annotation.Annotation, key string, revisions []*feature.Revision, features []*feature.Feature, id string, text string) func() ([]*feature.Result, error) {
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
				DatasetID:    ann.DatasetID,
				AnnotationID: ann.ID,
				FeatureID:    features[i].ID,
				PageKey:      key,
				Source:       source,
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
