package annotationrule

import "encoding/json"

// AnnotationRules wrapper type to help with custom unmarshaling of AppliedRules.
type AnnotationRules []AnnotationRule

func (a *AnnotationRules) UnmarshalJSON(data []byte) error {
	var rawRules []json.RawMessage
	if err := json.Unmarshal(data, &rawRules); err != nil {
		return err
	}

	*a = make([]AnnotationRule, 0, len(rawRules))
	for _, r := range rawRules {
		rule, err := UnmarshalRuleJSON(r)
		if err != nil {
			return err
		}
		*a = append(*a, rule)
	}

	return nil
}
