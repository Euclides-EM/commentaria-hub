package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode"

	"github.com/MiaMish/elements-dh/ocrflow/internal/features"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/textmatch"
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

type llmParseResult struct {
	results                []*feature.Result
	hallucinatedFeatureIDs []string
}

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

	// Run all apply functions in parallel — results are written to the DB as each job completes
	go func(executionID string, funcs []applyFunc, policy *feature.ExecutionPolicy) {
		var wg sync.WaitGroup
		var hasError atomic.Bool
		var completedKeys atomic.Int64
		totalKeys := int64(len(funcs))
		type executionJob struct {
			fn applyFunc
		}
		jobs := make(chan executionJob)

		workerCount := min(featureExecutionWorkerCount, len(funcs))
		for range workerCount {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobs {
					results, err := job.fn()
					if err != nil {
						log.Printf("error in execution %s: %v", executionID, err)
						hasError.Store(true)
					}
					if len(results) > 0 {
						if err := fe.featureResultsSvc.CreateResults(results, lo.IfF(policy != nil, func() bool { return policy.PushToOrigin }).Else(false)); err != nil {
							log.Printf("failed to create result for execution %s: %v", executionID, err)
							hasError.Store(true)
						}
					}
					done := completedKeys.Add(1)
					log.Printf("execution %s progress: key #%d/%d (%d%%)", executionID, done, totalKeys, done*100/totalKeys)
				}
			}()
		}
		for _, fn := range funcs {
			jobs <- executionJob{fn: fn}
		}
		close(jobs)

		wg.Wait()

		newStatus := feature.ExecutionStatusSuccess
		statusReason := ""
		if hasError.Load() {
			newStatus = feature.ExecutionStatusFailed
			statusReason = "one or more actions failed, check logs"
		}
		if err := fe.store.UpdateStatus(executionID, newStatus, statusReason); err != nil {
			log.Printf("failed to update execution status for execution %s: %v\n", executionID, err)
		}

	}(exec.ID, applyFuncs, exec.Policy)

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
		state.Add(result)
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

func buildPromptComponents(frs []*feature.Revision, fes []*feature.Feature) (featureNameToIndex map[string]int, definitions []string, outputFormat string) {
	featureNameToIndex = make(map[string]int)
	for i := range frs {
		featureName := fmt.Sprintf("%s-rev-%s", fes[i].ID, frs[i].ID[0:3])
		featureNameToIndex[featureName] = i
		definitions = append(definitions, fmt.Sprintf("- %s: %s", featureName, frs[i].Prompt))
		if fes[i].IsList {
			outputFormat += fmt.Sprintf(`  "%s": [...], // zero or more values`+"\n", featureName)
		} else {
			outputFormat += fmt.Sprintf(`  "%s": "...", // a single value or empty if not applicable`+"\n", featureName)
		}
	}
	outputFormat = strings.TrimSpace(outputFormat)
	return
}

func parseLLMResults(rawFields map[string]json.RawMessage, frs []*feature.Revision, fes []*feature.Feature, featureNameToIndex map[string]int, execID string, scope feature.ExecScope, key string, contextDesc string, sourceText string, checkHallucinations bool) (*llmParseResult, error) {
	for fn, rawValue := range rawFields {
		_, ok := featureNameToIndex[fn]
		if !ok {
			return nil, fmt.Errorf("llm response contained unknown field %q for %s\n%s", fn, contextDesc, rawValue)
		}
	}

	results := make([]*feature.Result, 0, len(frs))
	hallucinatedFeatureIDs := make([]string, 0)
	for i := range frs {
		fn := fmt.Sprintf("%s-rev-%s", fes[i].ID, frs[i].ID[0:3])
		rawValue, ok := rawFields[fn]
		var values []string
		if ok {
			if fes[i].IsList {
				if err := json.Unmarshal(rawValue, &values); err != nil {
					var val string
					if retryErr := json.Unmarshal(rawValue, &val); retryErr != nil {
						return nil, fmt.Errorf("failed to parse list response for field %q in %s: %w:\n%s", fn, contextDesc, err, rawValue)
					}
					if val != "" {
						values = []string{val}
					}
				}
			} else {
				var val string
				if err := json.Unmarshal(rawValue, &val); err != nil {
					return nil, fmt.Errorf("failed to parse scalar response for field %q in %s: %w\n%s", fn, contextDesc, err, rawValue)
				}
				if val != "" {
					values = []string{val}
				}
			}
		}

		source := feature.ResultSource{
			Resp:     "auto",
			Id:       execID,
			Revision: frs[i].ID,
			Name:     "llm",
		}
		res := &feature.Result{
			Scope:     scope,
			FeatureID: fes[i].ID,
			Key:       key,
			Source:    source,
		}
		var resultValues []feature.ResultValue
		for _, v := range values {
			v = trimFeatureValue(v)
			if v == "" {
				continue
			}
			if checkHallucinations && len(textmatch.FindLoosePhraseMatches(sourceText, v)) == 0 {
				log.Printf("!!! llm hallucination omitted: feature=%s revision=%s key=%s context=%s value=%q", fes[i].ID, frs[i].ID, key, contextDesc, v)
				if !slices.Contains(hallucinatedFeatureIDs, fes[i].ID) {
					hallucinatedFeatureIDs = append(hallucinatedFeatureIDs, fes[i].ID)
				}
				continue
			}
			resultValues = append(resultValues, feature.ResultValue{
				Surface: v,
			})
		}
		res.Values = resultValues
		results = append(results, res)
	}
	return &llmParseResult{
		results:                results,
		hallucinatedFeatureIDs: hallucinatedFeatureIDs,
	}, nil
}

func trimFeatureValue(s string) string {
	return strings.TrimFunc(strings.TrimSpace(s), func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	})
}
