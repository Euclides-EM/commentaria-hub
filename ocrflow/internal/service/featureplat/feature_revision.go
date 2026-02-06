package featureplat

import (
	"fmt"
	"time"

	"github.com/MiaMish/elements-dh/ocrflow/internal/model/featureplat"
	"github.com/MiaMish/elements-dh/ocrflow/internal/model/ocrflow"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/idgen"
	"github.com/samber/lo"
)

type Revision struct {
	m map[string]map[string]*featureplat.FeatureRevision
}

func NewRevision() *Revision {
	mockMap := make(map[string]map[string]*featureplat.FeatureRevision)
	mockMap["fea_ts8621"] = map[string]*featureplat.FeatureRevision{
		"fea_rev_001": {
			Meta: ocrflow.Meta{
				ID:        "fea_rev_001",
				Name:      "Initial revision",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Prompt:            "Initial prompt for institutions feature",
			ExecutionStrategy: featureplat.FeatureExecutionStrategyPrompt,
			Note:              "This is the initial revision for the institutions feature.",
			Type:              featureplat.FeatureTypeAnnotation,
		},
	}
	return &Revision{
		m: mockMap,
	}
}

func (fr *Revision) ListFeatureRevisions(collectionId, featureId string) ([]*featureplat.FeatureRevision, error) {
	if featureRevisions, ok := fr.m[featureId]; ok {
		return lo.Values(featureRevisions), nil
	}
	return nil, fmt.Errorf("feature revisions not found for feature ID: %s", featureId)
}

func (fr *Revision) CreateFeatureRevision(collectionId, featureId string, m *featureplat.FeatureRevision) (*featureplat.FeatureRevision, error) {
	m.ID = idgen.GenerateID("rev")
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	if _, ok := fr.m[featureId]; !ok {
		fr.m[featureId] = make(map[string]*featureplat.FeatureRevision)
	}
	fr.m[featureId][m.ID] = m
	return m, nil
}

func (fr *Revision) GetFeatureRevision(collectionId, featureId, revisionId string) (*featureplat.FeatureRevision, error) {
	if featureRevisions, ok := fr.m[featureId]; ok {
		if revision, ok := featureRevisions[revisionId]; ok {
			return revision, nil
		}
	}
	return nil, fmt.Errorf("feature revision not found for feature ID: %s and revision ID: %s", featureId, revisionId)
}

func (fr *Revision) DeleteFeatureRevision(collectionId, featureId, revisionId string) error {
	_, err := fr.GetFeatureRevision(collectionId, featureId, revisionId)
	if err != nil {
		return err
	}
	delete(fr.m[featureId], revisionId)
	return nil
}

func (fr *Revision) UpdateFeatureRevision(collectionId, featureId, revisionId string, m *featureplat.FeatureRevision) (*featureplat.FeatureRevision, error) {
	existing, err := fr.GetFeatureRevision(collectionId, featureId, revisionId)
	if err != nil {
		return nil, err
	}
	existing.Name = m.Name
	existing.Prompt = m.Prompt
	existing.ExecutionStrategy = m.ExecutionStrategy
	existing.Note = m.Note
	existing.Type = m.Type
	existing.UpdatedAt = time.Now()
	return existing, nil
}
