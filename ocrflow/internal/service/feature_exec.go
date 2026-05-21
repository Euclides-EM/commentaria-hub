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
	editionSvc          *Edition
	languageResolver    *LanguagesResolver
	featurePropertySvc  *FeatureProperty
	store               *fpstore.FeatureExecutionStore
	filesysManager      *filesys.Manager
	datasetImg          *DatasetImg
	llmClient           *llm.Client
	mu                  sync.Mutex // Protects status updates in goroutines
}

type executionActions struct {
	promptRevisions      []*feature.Revision
	promptFeatures       []*feature.Feature
	categorizerRevisions []*feature.Revision
	categorizerFeatures  []*feature.Feature
}

type aiConfigKey struct {
	provider feature.AIProvider
	model    string
}

func (a *executionActions) empty() bool {
	return len(a.categorizerRevisions) == 0 && len(a.promptRevisions) == 0
}

type applyFunc func() ([]*feature.Result, error)

// NewExecution returns a new Execution service using the given store (e.g. *storefeatureplat.FeatureExecutionStore).
func NewExecution(featureRevisionsSvc *Revision, featuresSvc *Feature, featureResultsSvc *Result, annotationSvc *Annotation, annotationTEISvc *AnnotationTEI, editionSvc *Edition, languageResolver *LanguagesResolver, featurePropertySvc *FeatureProperty, store *fpstore.FeatureExecutionStore, filesysManager *filesys.Manager, datasetImg *DatasetImg, llmClient *llm.Client) *Execution {
	return &Execution{
		featureRevisionsSvc: featureRevisionsSvc,
		featuresSvc:         featuresSvc,
		featureResultsSvc:   featureResultsSvc,
		annotationSvc:       annotationSvc,
		annotationTEISvc:    annotationTEISvc,
		editionSvc:          editionSvc,
		languageResolver:    languageResolver,
		featurePropertySvc:  featurePropertySvc,
		store:               store,
		filesysManager:      filesysManager,
		datasetImg:          datasetImg,
		llmClient:           llmClient,
	}
}

func (fe *Execution) ListFeatureExecutions(scope feature.DefScope, featureIds []string, statuses []feature.ExecutionStatus) ([]*feature.Execution, error) {
	res, err := fe.store.List(scope, featureIds, statuses)
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
	exec.ID = idgen.GenerateID("exec")
	exec.Status = feature.ExecutionStatusInProgress
	exec.StatusReason = ""

	skipState, err := fe.loadExecutionSkipState(exec)
	if err != nil {
		return nil, err
	}

	var applyFuncs []applyFunc
	for _, key := range exec.Keys {
		actions, err := fe.loadExecutionActions(exec, key, skipState)
		if err != nil {
			return nil, err
		}
		if actions.empty() {
			log.Printf("skipping execution %s edition %s because all actions were skipped by policy", exec.ID, key)
			continue
		}
		switch exec.Scope.Type {
		case feature.ScopeTypeDataset:
			applyFuncs = append(applyFuncs, fe.annotationApplyFunc(exec, key, actions))
		case feature.ScopeTypeEditions:
			applyFuncs = append(applyFuncs, fe.editionApplyFunc(key, actions, exec.ID))
		default:
			return nil, fmt.Errorf("invalid execution scope: %s", exec.Scope.Type)
		}

	}

	if err := fe.store.Create(exec); err != nil {
		return nil, err
	}

	// Run all apply functions in parallel and wait for them to finish
	go func(executionID string, funcs []applyFunc) {
		var wg sync.WaitGroup
		resultBatches := make([][]*feature.Result, len(funcs))
		errs := make([]error, len(funcs))
		type executionJob struct {
			index int
			fn    applyFunc
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
					errs[job.index] = err
				}
			}()
		}
		for i, fn := range funcs {
			jobs <- executionJob{index: i, fn: fn}
		}
		close(jobs)

		wg.Wait()

		hasError := false
		for _, err := range errs {
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
			if err := fe.featureResultsSvc.CreateResults(results, lo.IfF(exec.Policy != nil, func() bool { return exec.Policy.PushToOrigin }).Else(false)); err != nil {
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
	results, err := fe.featureResultsSvc.ListResultsForExecutionPolicy(exec, featureIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to preload feature results for execution policy: %w", err)
	}

	for _, result := range results {
		if result == nil {
			continue
		}
		state.Add(result.FeatureID, result.Key, result.Source.Revision, result.Source.Resp == "human")
	}

	return state, nil
}

func (fe *Execution) loadExecutionActions(exec *feature.Execution, key string, skipState *features.ExecutionSkipState) (*executionActions, error) {
	actions := &executionActions{}
	for _, item := range exec.Apply {
		feat, err := fe.featuresSvc.GetFeatureInScope(exec.Scope.DefScope, item.Feature, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get feature for feature %s: %w", item.Feature, err)
		}
		fr, err := fe.featureRevisionsSvc.GetFeatureRevisionInScope(exec.Scope.DefScope, item.Feature, item.Revision)
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
			actions.categorizerRevisions = append(actions.categorizerRevisions, fr)
			actions.categorizerFeatures = append(actions.categorizerFeatures, feat)
		} else if fr.Prompt != "" {
			actions.promptRevisions = append(actions.promptRevisions, fr)
			actions.promptFeatures = append(actions.promptFeatures, feat)
		} else {
			return nil, fmt.Errorf("feature revision for feature %s and revision %s does not have a valid execution strategy", item.Feature, item.Revision)
		}
	}
	return actions, nil
}

func (fe *Execution) finishExecution(executionID string, policy *feature.ExecutionPolicy, funcs []applyFunc) error {
	newStatus := feature.ExecutionStatusSuccess
	statusReason := "edition metadata prompt execution is stubbed; no results were produced"

	if len(funcs) > 0 {
		var results []*feature.Result
		var execErrs []error
		for _, fn := range funcs {
			batch, err := fn()
			results = append(results, batch...)
			if err != nil {
				execErrs = append(execErrs, err)
			}
		}
		if len(results) > 0 {
			if err := fe.featureResultsSvc.CreateResults(results, lo.IfF(policy != nil, func() bool { return policy.PushToOrigin }).Else(false)); err != nil {
				return fmt.Errorf("failed to store execution results: %w", err)
			}
		}
		if err := errors.Join(execErrs...); err != nil {
			newStatus = feature.ExecutionStatusFailed
			statusReason = "one or more actions failed, check logs"
		}
	}
	return fe.store.UpdateStatus(executionID, newStatus, statusReason)
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

type promptRevisionGroup struct {
	revisions []*feature.Revision
	features  []*feature.Feature
}

func groupPromptRevisionsByAIConfig(revisions []*feature.Revision, features []*feature.Feature) []promptRevisionGroup {
	groupIndexes := make(map[aiConfigKey]int)
	var groups []promptRevisionGroup
	for i, rev := range revisions {
		key := aiConfigKey{provider: rev.AIProvider, model: rev.AIModel}
		groupIndex, ok := groupIndexes[key]
		if !ok {
			groupIndexes[key] = len(groups)
			groups = append(groups, promptRevisionGroup{})
			groupIndex = len(groups) - 1
		}
		groups[groupIndex].revisions = append(groups[groupIndex].revisions, rev)
		groups[groupIndex].features = append(groups[groupIndex].features, features[i])
	}
	return groups
}

func (fe *Execution) annotationPromptApplyFunc(ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string, textLanguage string, fullText string) applyFunc {
	return func() ([]*feature.Result, error) {
		if strings.TrimSpace(fullText) == "" {
			return nil, fmt.Errorf("full text is empty for dataset %s and key %s", ann.DatasetID, key)
		}
		aiProvider := frs[0].AIProvider
		aiModel := frs[0].AIModel
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

		rawResponse, err := fe.llmClient.Exec(aiProvider.ToLLMAIProvider(), aiModel, prompt, "")
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for dataset %s and key %s using %s/%s: %w", ann.DatasetID, key, aiProvider, aiModel, err)
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
				Scope:     feature.NewDatasetExecScope(ann.DatasetID, ann.ID),
				FeatureID: fes[idx].ID,
				Key:       key,
				Source:    source,
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

func (fe *Execution) editionApplyFunc(editionKey string, actions *executionActions, execID string) applyFunc {
	return func() ([]*feature.Result, error) {
		edition, err := fe.editionSvc.GetEditionByID(editionKey)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata for edition %s: %w", editionKey, err)
		}

		log.Printf(
			"stubbed edition metadata execution %s for edition %s (%s) with actions: %v",
			execID,
			editionKey,
			edition.ShortTitle,
			actions,
		)
		return nil, nil
	}
}
