package tei

import (
	"sort"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/tei/model"
)

func buildInlineNodesWithAnchors(blockID, lineID, line string, entities []EntityItem) []model.ABNode {
	type eventKind int
	const (
		evEnd eventKind = iota
		evStart
	)

	type event struct {
		pos  int
		kind eventKind
		i    int // entity index (used for anchor id)
	}

	b := []byte(line)
	n := len(b)

	events := make([]event, 0)

	// Build anchor events with bounds checking.
	for i, ent := range entities {
		if ent.Start.LineID == lineID && ent.Start.BlockID == blockID {
			s := ent.Start.ByteOffset
			// Clamp to line length so we never panic on slicing.
			if s > n {
				s = n
			}
			events = append(events,
				event{pos: s, kind: evStart, i: i},
			)
		}

		if ent.End.LineID == lineID && ent.End.BlockID == blockID {
			e := ent.End.ByteOffset
			if e > n {
				e = n
			}
			events = append(events,
				event{pos: e, kind: evEnd, i: i},
			)
		}
	}

	// Sort by position, then end-before-start at same position, then by entity index for determinism.
	sort.SliceStable(events, func(a, b int) bool {
		if events[a].pos != events[b].pos {
			return events[a].pos < events[b].pos
		}
		if events[a].kind != events[b].kind {
			return events[a].kind < events[b].kind // evEnd (0) before evStart (1)
		}
		return events[a].i < events[b].i
	})

	nodes := make([]model.ABNode, 0, len(events)+1)

	emitText := func(txt []byte) {
		if len(txt) == 0 {
			return
		}
		nodes = append(nodes, model.ABNode{CharData: string(txt)})
	}

	emitAnchor := func(id string) {
		nodes = append(nodes, model.ABNode{
			Anchor: &model.Anchor{XmlID: id},
		})
	}

	cur := 0
	for _, ev := range events {
		// Emit text up to this event position.
		if ev.pos > cur {
			emitText(b[cur:ev.pos])
			cur = ev.pos
		}

		switch ev.kind {
		case evStart:
			emitAnchor(startMentionAnchorID(ev.i))
		case evEnd:
			emitAnchor(endMentionAnchorID(ev.i))
		}
	}

	// Emit remaining text.
	if cur < n {
		emitText(b[cur:])
	}

	return nodes
}
