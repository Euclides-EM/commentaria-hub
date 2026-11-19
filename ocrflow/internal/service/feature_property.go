package service

import (
	"fmt"
	"strings"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/normalize"
	"github.com/samber/lo"
)

type FeatureProperty struct {
}

func NewFeatureProperty() *FeatureProperty {
	return &FeatureProperty{}
}

func (fp *FeatureProperty) ListFeaturePropertyKeys() []string {
	return lo.Keys(featurePropToFunc)
}

func (fp *FeatureProperty) CalcValByPropertyKey(s, propKey string) (string, error) {
	vals, err := fp.CalcValsByPropertyKey(s, propKey)
	if err != nil {
		return "", err
	}
	if len(vals) == 0 {
		return "", nil
		//return "", fmt.Errorf("no values calculated for property key: %s", propKey)
	}
	strVals := lo.Map(vals, func(v normalize.MappedOriginal, _ int) string {
		return v.Mapped
	})
	strVals = lo.Uniq(strVals)
	return strings.Join(strVals, "::"), nil
}

func (fp *FeatureProperty) CalcValsByPropertyKey(s, propKey string) ([]normalize.MappedOriginal, error) {
	propFunc, ok := featurePropToFunc[propKey]
	if !ok {
		return nil, fmt.Errorf("unknown feature property key: %s", propKey)
	}
	return propFunc(s), nil
}

func (fp *FeatureProperty) ListDefaultFeaturePropertyKeys() []string {
	return []string{"normalized"}
}

var featurePropToFunc = map[string]func(v string) []normalize.MappedOriginal{
	"normalized":      normalize.String,
	"language":        normalize.Language,
	"institution":     normalize.Institution,
	"ancient_persona": normalize.AncientPersona,
}
