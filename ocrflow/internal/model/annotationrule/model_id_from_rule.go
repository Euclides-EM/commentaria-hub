package annotationrule

func ExtractModelIDsFromRules(rule []AnnotationRule) []string {
	modelIDSet := make(map[string]struct{})

	var extract func(r AnnotationRule)
	extract = func(r AnnotationRule) {
		switch v := r.(type) {
		case *ModelDetect:
			if v.Model != "" {
				modelIDSet[v.Model] = struct{}{}
			}
		}
	}
	for _, r := range rule {
		extract(r)
	}

	var modelIDs []string
	for id := range modelIDSet {
		modelIDs = append(modelIDs, id)
	}
	return modelIDs
}
