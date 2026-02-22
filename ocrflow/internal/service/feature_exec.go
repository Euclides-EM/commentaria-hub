package service

import (
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/annotation"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/feature"
	fpstore "github.com/MiaMish/elements-dh/ocrflow/internal/store"
	"github.com/MiaMish/elements-dh/ocrflow/internal/store/filesys"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/llm"
)

type Execution struct {
	featureRevisionsSvc *Revision
	featuresSvc         *Feature
	featureResultsSvc   *Result
	annotationSvc       *Annotation
	store               *fpstore.FeatureExecutionSQL
	filesysManager      *filesys.Manager
	datasetImg          *DatasetImg
	llmClient           *llm.Client
	mu                  sync.Mutex // Protects status updates in goroutines
}

// NewExecution returns a new Execution service using the given store (e.g. *storefeatureplat.FeatureExecutionSQL).
func NewExecution(featureRevisionsSvc *Revision, featuresSvc *Feature, featureResultsSvc *Result, annotationSvc *Annotation, store *fpstore.FeatureExecutionSQL, filesysManager *filesys.Manager, datasetImg *DatasetImg, llmClient *llm.Client) *Execution {
	return &Execution{
		featureRevisionsSvc: featureRevisionsSvc,
		featuresSvc:         featuresSvc,
		featureResultsSvc:   featureResultsSvc,
		annotationSvc:       annotationSvc,
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

	var applyFuncs []func() ([]*feature.Result, error)
	for _, key := range exec.Keys {
		for _, item := range exec.Apply {
			fr, err := fe.featureRevisionsSvc.GetFeatureRevision(exec.DatasetID, item.Feature, item.Revision)
			if err != nil {
				return nil, fmt.Errorf("failed to get feature revision for feature %s and revision %s: %w", item.Feature, item.Revision, err)
			}
			feat, err := fe.featuresSvc.GetFeature(exec.DatasetID, item.Feature, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to get feature for feature %s: %w", item.Feature, err)
			}
			switch fr.ExecutionStrategy {
			case feature.ExecutionStrategyPrompt:
				// (ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string)
				applyFuncs = append(applyFuncs, fe.execPrompt(ann, key, []*feature.Revision{fr}, []*feature.Feature{feat}, exec.ID))
			case feature.ExecutionStrategyRegex:
				applyFuncs = append(applyFuncs, fe.execRegex())
			default:
				return nil, fmt.Errorf("unsupported execution strategy %s for feature %s", fr.ExecutionStrategy, item.Feature)
			}
		}
	}

	if err := fe.store.Create(exec); err != nil {
		return nil, err
	}

	// Run all apply functions in parallel and wait for them to finish
	go func(executionID string, funcs []func() ([]*feature.Result, error)) {
		var wg sync.WaitGroup
		var results []*feature.Result
		errors := make([]error, len(funcs))

		for i, fn := range funcs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				result, err := fn()
				results = append(results, result...)
				errors[i] = err
			}(i)
		}

		wg.Wait()

		// Check for errors and update execution status accordingly
		hasError := false
		for _, err := range errors {
			if err != nil {
				log.Printf("error in execution %s: %v\n", executionID, err)
				hasError = true
				break
			}
		}

		if hasError {
			if err := fe.store.UpdateStatus(executionID, feature.ExecutionStatusFailed, "error in execution, check logs"); err != nil {
				log.Printf("failed to update execution status for execution %s: %v\n", executionID, err)
			}
			return
		}

		newStatus := feature.ExecutionStatusSuccess
		err = fe.featureResultsSvc.CreateResults(results)
		if err != nil {
			log.Printf("failed to create result for execution %s: %v\n", executionID, err)
			newStatus = feature.ExecutionStatusFailed
		}
		if err := fe.store.UpdateStatus(executionID, newStatus, "failed to store execution results"); err != nil {
			log.Printf("failed to update execution status for execution %s: %v\n", executionID, err)
		}

	}(exec.ID, applyFuncs)

	return exec, nil
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

func (fe *Execution) execPrompt(ann *annotation.Annotation, key string, frs []*feature.Revision, fes []*feature.Feature, execID string) func() ([]*feature.Result, error) {
	return func() ([]*feature.Result, error) {
		img, err := fe.datasetImg.GetImageMetadata(ann.DatasetID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get image metadata for dataset %s and key %s: %w", ann.DatasetID, key, err)
		}
		imgDir := fe.filesysManager.DatasetImagesDirByID(ann.DatasetID)
		imgPath := filepath.Join(imgDir, img.Filename)
		dsPromptDesc := "historical title pages of translations of Euclid's Elements"
		dsPromptShortDesc := "title page"
		var definitions []string
		var outputFormat string
		featureNameToIndex := make(map[string]int)
		for i, _ := range frs {
			featureName := fmt.Sprintf("%s-rev-%s", fes[i].Name, frs[i].ID)
			featureNameToIndex[featureName] = i
			definitions = append(definitions, fmt.Sprintf("- %s: %s", featureName, frs[i].Prompt))
			if fes[i].IsList {
				outputFormat += fmt.Sprintf(`  "%s": [...], // zero or more quotes`+"\n", featureName)
			} else {
				outputFormat += fmt.Sprintf(`  "%s": "...", // a single quote or empty if not applicable`+"\n", featureName)
			}
		}
		prompt := fmt.Sprintf(`You are an AI agent designed to extract structured metadata from %s.

You will be given:
- The transcribed text of a %s.
- The language of the transcription.

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
`, dsPromptDesc, dsPromptShortDesc, outputFormat, definitions)

		var response map[string][]string
		response, err = fe.llmClient.Exec(prompt, imgPath)
		if err != nil {
			return nil, fmt.Errorf("failed to execute LLM prompt for dataset %s and key %s: %w", ann.DatasetID, key, err)
		}

		var results []*feature.Result
		for fn, quotes := range response {
			source := feature.ResultSource{
				Resp:     "auto",
				Id:       execID,
				Revision: frs[featureNameToIndex[fn]].ID,
				Name:     "llm",
			}
			res := &feature.Result{
				DatasetID:    ann.DatasetID,
				AnnotationID: ann.ID,
				Feature:      fes[featureNameToIndex[fn]].Name,
				Key:          key,
				Source:       source,
			}
			var resultValues []feature.ResultValue
			for _, quote := range quotes {
				resultValues = append(resultValues, feature.ResultValue{
					Root:   quote,
					Source: &source,
				})
			}
			res.Values = resultValues
		}
		return results, nil
	}
}

func (fe *Execution) execRegex() func() ([]*feature.Result, error) {
	return func() ([]*feature.Result, error) {
		return nil, fmt.Errorf("regex execution strategy not implemented yet")
	}
}
