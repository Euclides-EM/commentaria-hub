package annotationrules

import "errors"

type StretchTowardsCategory struct {
	StretchCategory string
	Towards         string
	ContactType     ContactType
	ContactSide     Side
}

type stretchTowardsCategoryBuilder struct {
	stretchTowardsCategory *StretchTowardsCategory
}

func (b *stretchTowardsCategoryBuilder) Stretch(category string) *stretchTowardsCategoryBuilder {
	b.stretchTowardsCategory.StretchCategory = category
	return b
}

func (b *stretchTowardsCategoryBuilder) Towards(towards string, contactType ContactType, contactSide Side) *stretchTowardsCategoryBuilder {
	b.stretchTowardsCategory.Towards = towards
	b.stretchTowardsCategory.ContactType = contactType
	b.stretchTowardsCategory.ContactSide = contactSide
	return b
}

func (b *stretchTowardsCategoryBuilder) Build() (*StretchTowardsCategory, error) {
	if b.stretchTowardsCategory.StretchCategory == "" {
		return nil, errors.New("stretch category is required")
	}
	if b.stretchTowardsCategory.Towards == "" {
		return nil, errors.New("towards is required")
	}
	if b.stretchTowardsCategory.ContactType == "" {
		return nil, errors.New("contact type is required")
	}
	if b.stretchTowardsCategory.ContactSide == "" {
		return nil, errors.New("contact side is required")
	}
	return b.stretchTowardsCategory, nil
}

func StretchTowardsCategoryBuilder() *stretchTowardsCategoryBuilder {
	return &stretchTowardsCategoryBuilder{
		stretchTowardsCategory: &StretchTowardsCategory{},
	}
}
