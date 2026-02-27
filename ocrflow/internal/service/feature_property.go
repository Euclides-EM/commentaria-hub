package service

import (
	"fmt"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/normalize"
	"github.com/samber/lo"
)

type FeatureProperty struct {
}

func NewFeatureProperty() *FeatureProperty {
	return &FeatureProperty{}
}

func (fp *FeatureProperty) ListFeaturePropertyKeys() ([]string, error) {
	return lo.Keys(featurePropToFunc), nil
}

func (fp *FeatureProperty) CalcFeaturePropertyByPropertyKey(propKey string) (string, error) {
	propFunc, ok := featurePropToFunc[propKey]
	if !ok {
		return "", fmt.Errorf("unknown feature property key: %s", propKey)
	}
	return propFunc(propKey), nil
}

var featurePropToFunc = map[string]func(v string) string{
	"normalized":      normalize.String,
	"language":        normalize.Language,
	"institution":     normalize.Institution,
	"ancient_persona": normalize.AncientPersona,
}
