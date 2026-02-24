package tei

import "sort"

func buildEntitiesOccurrences(entities *EntitiesInput) map[string][]EntityOccurrence {
	// Index occurrences by (page, block, line)
	occByKey := make(map[string][]EntityOccurrence)
	var entitiesOccurrences []EntityOccurrence
	if entities != nil {
		entitiesOccurrences = entities.Occurrences
	}
	for _, oc := range entitiesOccurrences {
		if oc.Element == "" {
			oc.Element = "name"
		}
		k := occKey(oc.PageID, oc.BlockID, oc.LineID)
		occByKey[k] = append(occByKey[k], oc)
	}

	// Sort occurrences in each line by Start asc, then End desc
	for k := range occByKey {
		sort.Slice(occByKey[k], func(i, j int) bool {
			if occByKey[k][i].Start != occByKey[k][j].Start {
				return occByKey[k][i].Start < occByKey[k][j].Start
			}
			return occByKey[k][i].End > occByKey[k][j].End
		})
	}

	return occByKey
}
