package tei

import (
	"fmt"
	"sort"
	"strings"
)

// mentionInLine is the internal shape for one inline mention; MentionID is auto-assigned in document order.
type mentionInLine struct {
	Start     int
	End       int
	Ref       string
	Ana       string
	MentionID string // e.g. "m_1", "m_2"
}

func hasPosition(it *EntityItem) bool {
	return strings.TrimSpace(it.PageID) != "" && strings.TrimSpace(it.LineID) != "" && it.End > it.Start
}

// getLineText returns the line text for a given occKey; used to resolve positions for entities without position.
// If nil, entities without position are not placed in the text.
type getLineText func(key string) string

// buildEntitiesOccurrences returns mentions per (page, block, line) with auto-assigned mention IDs in document order.
// orderedKeys must be the list of occKey values in document order (so mention IDs are stable).
// If getLineText is non-nil, entities without position are placed by searching for their Value in each line (first match per entity).
func buildEntitiesOccurrences(entities []EntityItem, orderedKeys []string, lineText getLineText) map[string][]mentionInLine {
	// Split: items with position vs unplaced (no position but have Value to search)
	var withPosition []*EntityItem
	var unplaced []*EntityItem
	for i := range entities {
		it := &entities[i]
		if hasPosition(it) {
			withPosition = append(withPosition, it)
		} else if strings.TrimSpace(it.Value) != "" {
			unplaced = append(unplaced, it)
		}
	}

	// Group items with position by key
	byKey := make(map[string][]*EntityItem)
	for _, it := range withPosition {
		it.Element = "name"
		k := occKey(it.PageID, it.BlockID, it.LineID)
		byKey[k] = append(byKey[k], it)
	}

	// Resolve unplaced: search each line for Value, first match wins per entity
	if lineText != nil && len(unplaced) > 0 {
		placed := make([]bool, len(unplaced))
		for _, k := range orderedKeys {
			text := lineText(k)
			if text == "" {
				continue
			}
			for i, it := range unplaced {
				if placed[i] {
					continue
				}
				val := strings.TrimSpace(it.Value)
				idx := strings.Index(text, val)
				if idx < 0 {
					continue
				}
				// Clone item with resolved position and append to this key's list
				resolved := *it
				resolved.Start = idx
				resolved.End = idx + len(val)
				resolved.Element = "name"
				byKey[k] = append(byKey[k], &resolved)
				placed[i] = true
			}
		}
	}

	// Sort each key's list by Start
	for k := range byKey {
		sort.Slice(byKey[k], func(i, j int) bool {
			if byKey[k][i].Start != byKey[k][j].Start {
				return byKey[k][i].Start < byKey[k][j].Start
			}
			return byKey[k][i].End > byKey[k][j].End
		})
	}
	// Assign mention IDs in document order and build result
	occByKey := make(map[string][]mentionInLine)
	mid := 0
	for _, k := range orderedKeys {
		items := byKey[k]
		if len(items) == 0 {
			continue
		}
		list := make([]mentionInLine, 0, len(items))
		for _, it := range items {
			mid++
			ref := "#" + normalizedEntityID(it.Ref)
			list = append(list, mentionInLine{
				Start:     it.Start,
				End:       it.End,
				Ref:       ref,
				Ana:       it.Ana,
				MentionID: fmt.Sprintf("m_%d", mid),
			})
		}
		occByKey[k] = list
	}
	return occByKey
}
