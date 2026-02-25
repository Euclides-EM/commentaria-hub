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

// locationLess returns true if a is before b in document order (page -> block -> line -> byte).
func locationLess(a, b EntityLocationIndex) bool {
	if a.PageID != b.PageID {
		return a.PageID < b.PageID
	}
	if a.BlockID != b.BlockID {
		return a.BlockID < b.BlockID
	}
	if a.LineID != b.LineID {
		return a.LineID < b.LineID
	}
	return a.ByteOffset < b.ByteOffset
}

func sameLocationLine(a, b EntityLocationIndex) bool {
	return a.PageID == b.PageID && a.BlockID == b.BlockID && a.LineID == b.LineID
}

func hasPosition(it *EntityItem) bool {
	return strings.TrimSpace(it.Start.PageID) != "" && strings.TrimSpace(it.Start.LineID) != "" && locationLess(it.Start, it.End)
}

// getLineText returns the line text for a given occKey; used to resolve positions for entities without position.
// If nil, entities without position are not placed in the text.
type getLineText func(key string) string

// segment is a single-line slice of an entity mention (byte range within one line).
type segment struct {
	key       string
	startByte int
	endByte   int
	ref       string
	ana       string
}

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

	keyIndex := make(map[string]int)
	for i, k := range orderedKeys {
		keyIndex[k] = i
	}

	var segments []segment

	// Expand items with position into per-line segments (one or more per entity)
	for _, it := range withPosition {
		startKey := occKey(it.Start.PageID, it.Start.BlockID, it.Start.LineID)
		endKey := occKey(it.End.PageID, it.End.BlockID, it.End.LineID)
		ref := "#" + normalizedEntityID(it.Ref)
		if sameLocationLine(it.Start, it.End) {
			segments = append(segments, segment{
				key:       startKey,
				startByte: it.Start.ByteOffset,
				endByte:   it.End.ByteOffset,
				ref:       ref,
				ana:       it.Ana,
			})
			continue
		}
		// Multi-line: emit one segment per line in [startKey .. endKey]
		startIdx, okStart := keyIndex[startKey]
		endIdx, okEnd := keyIndex[endKey]
		if !okStart || !okEnd || startIdx > endIdx {
			continue
		}
		if lineText == nil {
			continue
		}
		for i := startIdx; i <= endIdx; i++ {
			k := orderedKeys[i]
			lineLen := len(lineText(k))
			var segStart, segEnd int
			if i == startIdx {
				segStart = it.Start.ByteOffset
			} else {
				segStart = 0
			}
			if i == endIdx {
				segEnd = it.End.ByteOffset
			} else {
				segEnd = lineLen
			}
			if segEnd > segStart {
				segments = append(segments, segment{key: k, startByte: segStart, endByte: segEnd, ref: ref, ana: it.Ana})
			}
		}
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
				segments = append(segments, segment{
					key:       k,
					startByte: idx,
					endByte:   idx + len(val),
					ref:       "#" + normalizedEntityID(it.Ref),
					ana:       it.Ana,
				})
				placed[i] = true
			}
		}
	}

	// Group segments by key
	byKey := make(map[string][]segment)
	for _, s := range segments {
		byKey[s.key] = append(byKey[s.key], s)
	}

	// Sort each key's list by start then end desc
	for k := range byKey {
		sort.Slice(byKey[k], func(i, j int) bool {
			if byKey[k][i].startByte != byKey[k][j].startByte {
				return byKey[k][i].startByte < byKey[k][j].startByte
			}
			return byKey[k][i].endByte > byKey[k][j].endByte
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
		for _, s := range items {
			mid++
			list = append(list, mentionInLine{
				Start:     s.startByte,
				End:       s.endByte,
				Ref:       s.ref,
				Ana:       s.ana,
				MentionID: fmt.Sprintf("m_%d", mid),
			})
		}
		occByKey[k] = list
	}
	return occByKey
}
