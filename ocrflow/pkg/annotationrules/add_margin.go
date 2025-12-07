package annotationrules

import "errors"

type AddMargin struct {
	Side     Side
	Margin   float64
	Category string
}

type addMarginBuilder struct {
	addMargin *AddMargin
}

func AddMarginBuilder() *addMarginBuilder {
	return &addMarginBuilder{
		addMargin: &AddMargin{},
	}
}

func (b *addMarginBuilder) Side(side Side) *addMarginBuilder {
	b.addMargin.Side = side
	return b
}

func (b *addMarginBuilder) Margin(margin float64) *addMarginBuilder {
	b.addMargin.Margin = margin
	return b
}

func (b *addMarginBuilder) Category(category string) *addMarginBuilder {
	b.addMargin.Category = category
	return b
}

func (b *addMarginBuilder) Build() (*AddMargin, error) {
	if b.addMargin.Side == "" {
		return nil, errors.New("side is required")
	}
	if b.addMargin.Margin <= 0 {
		return nil, errors.New("margin must be greater than 0")
	}
	if b.addMargin.Category == "" {
		return nil, errors.New("category is required")
	}
	return b.addMargin, nil
}
