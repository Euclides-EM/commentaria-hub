package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/features"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/feature"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

func (fe *Execution) ExecuteEphemeral(exec *feature.Execution, revisions []*feature.Revision, featuresList []*feature.Feature) ([]*feature.Result, error) {
	if exec == nil {
		return nil, fmt.Errorf("execution is required")
	}
	if len(exec.Apply) == 0 {
		return nil, fmt.Errorf("execution apply is required")
	}
	if len(revisions) != len(featuresList) {
		return nil, fmt.Errorf("revisions and features length mismatch")
	}

	exec.ID = idgen.GenerateID("exec")
	skipState, err := fe.loadExecutionSkipState(exec)
	if err != nil {
		return nil, err
	}

	revisionsByKey := make(map[string]*feature.Revision, len(revisions))
	featuresByKey := make(map[string]*feature.Feature, len(featuresList))
	for i := range revisions {
		key := ephemeralActionKey(featuresList[i].ID, revisions[i].ID)
		revisionsByKey[key] = revisions[i]
		featuresByKey[key] = featuresList[i]
	}

	var allResults []*feature.Result
	var execErrs []error
	for i, key := range exec.Keys {
		actions, err := fe.loadExecutionActionsEphemeral(exec, key, skipState, revisionsByKey, featuresByKey)
		if err != nil {
			return nil, err
		}
		if actions.empty() {
			continue
		}

		var batch []*feature.Result
		switch exec.Scope.Type {
		case feature.ScopeTypeDataset:
			batch, err = fe.annotationApplyFunc(exec, key, actions)()
		case feature.ScopeTypeEditions:
			batch, err = fe.editionApplyFunc(key, actions, exec.ID, fmt.Sprintf("[%d/%d]", i+1, len(exec.Keys)))()
		default:
			return nil, fmt.Errorf("invalid execution scope: %s", exec.Scope.Type)
		}
		allResults = append(allResults, batch...)
		if err != nil {
			execErrs = append(execErrs, err)
		}
	}

	return allResults, errors.Join(execErrs...)
}

func (fe *Execution) loadExecutionActionsEphemeral(exec *feature.Execution, key string, skipState *features.ExecutionSkipState, revisionsByKey map[string]*feature.Revision, featuresByKey map[string]*feature.Feature) (*executionActions, error) {
	actions := &executionActions{}
	for _, item := range exec.Apply {
		mapKey := ephemeralActionKey(item.Feature, item.Revision)
		feat, ok := featuresByKey[mapKey]
		if !ok {
			return nil, fmt.Errorf("missing ephemeral feature definition for feature %s revision %s", item.Feature, item.Revision)
		}
		fr, ok := revisionsByKey[mapKey]
		if !ok {
			return nil, fmt.Errorf("missing ephemeral revision definition for feature %s revision %s", item.Feature, item.Revision)
		}
		skipReasons := skipState.SkipReasons(exec.Policy, key, item.Feature, item.Revision)
		if len(skipReasons) > 0 {
			log.Printf("skipping ephemeral execution %s key %s feature %s revision %s due to skip policy: %s", exec.ID, key, item.Feature, item.Revision, strings.Join(lo.Map(skipReasons, func(reason feature.ExecutionSkipIf, _ int) string {
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
			return nil, fmt.Errorf("ephemeral feature revision for feature %s and revision %s does not have a valid execution strategy", item.Feature, item.Revision)
		}
	}
	return actions, nil
}

func ephemeralActionKey(featureID, revisionID string) string {
	return featureID + "::" + revisionID
}
