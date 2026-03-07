package tei

import (
	"encoding/xml"
	"fmt"
	"sort"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/tei/model"
)

type rect struct {
	x float64
	y float64
	w float64
	h float64
}

func (r rect) contains(o rect) bool {
	return o.x >= r.x &&
		o.y >= r.y &&
		o.x+o.w <= r.x+r.w &&
		o.y+o.h <= r.y+r.h
}

type lineItem struct {
	parentBlockIdx int
	srcBlockIdx    int

	y float64
	x float64

	line model.L
}

// BuildTEIFromALTO builds TEI for a single page ALTO.
func BuildTEIFromALTO(
	pageKey string,
	a *alto.Alto,
	entities []EntityItem,
	imageUrl string,
) (*model.TEI, error) {

	if len(a.Layout.Page) != 1 {
		return nil, fmt.Errorf("expected exactly one page in ALTO, got %d", len(a.Layout.Page))
	}

	page := a.Layout.Page[0]
	blocks := page.PrintSpace.TextBlocks

	doc := &model.TEI{
		XMLName: xml.Name{Local: "TEI"},
		Xmlns:   "http://www.tei-c.org/ns/1.0",
		Header: model.Header{
			FileDesc: buildFileDesc(),
			StandOff: buildStandOff(pageKey, entities, blocks),
		},
		Facsimile: buildFacsimileForAlto(pageKey, imageUrl, a),
		Text:      model.Text{Body: model.Body{}},
	}

	doc.Text.Body.PBs = append(doc.Text.Body.PBs, model.PB{
		Facs: "#" + doc.Facsimile.Surfaces[0].XmlID,
		N:    pageKey,
	})

	// 1) Compute block rectangles
	blockRects := make([]rect, len(blocks))
	for i, b := range blocks {
		blockRects[i] = rect{
			x: b.HPOS,
			y: b.VPOS,
			w: b.Width,
			h: b.Height,
		}
	}

	// 2) For each block, find a containing parent block (smallest container wins)
	parentOf := make([]int, len(blocks))
	for i := range blocks {
		parentOf[i] = i // default: self
		src := blockRects[i]

		best := -1
		bestArea := float64(int(^uint(0) >> 1)) // max int
		for j := range blocks {
			if i == j {
				continue
			}
			cand := blockRects[j]
			if !cand.contains(src) {
				continue
			}
			area := cand.w * cand.h
			if area < bestArea {
				bestArea = area
				best = j
			}
		}
		if best != -1 {
			parentOf[i] = best
		}
	}

	// Optional: if containment is not enough, restrict parenting by category.
	// Eg only allow parent if it is "MainZone". This requires a function that
	// extracts your category from the block.
	//
	// parentOf = restrictParentsByCategory(blocks, parentOf)

	// 3) Flatten all lines into one stream
	items := make([]lineItem, 0, 256)
	// Keep per-block line counters to preserve your lineID(pageKey, blockN, lineN) scheme
	perBlockLineN := make([]int, len(blocks))

	for bi, b := range blocks {
		for _, tl := range b.Lines {
			perBlockLineN[bi]++
			j := perBlockLineN[bi]

			nodes := buildInlineNodesWithAnchors(b.ID, tl.ID, alto.ExtractTextFromLine(tl), entities)
			l := model.L{
				XmlID: lineID(pageKey, bi+1, j),
				Facs:  "#" + facZoneLineID(pageKey, bi+1, j),
				Nodes: nodes,
			}

			items = append(items, lineItem{
				parentBlockIdx: parentOf[bi],
				srcBlockIdx:    bi,
				y:              tl.VPOS,
				x:              tl.HPOS,
				line:           l,
			})
		}
	}

	// 4) Sort by reading order (top to bottom, left to right)
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].y != items[j].y {
			return items[i].y < items[j].y
		}
		if items[i].x != items[j].x {
			return items[i].x < items[j].x
		}
		return false
	})

	// 5) Chunk into ABs by (parentBlockIdx, srcBlockIdx) runs
	abs := make([]model.AB, 0, 16)
	var cur *model.AB
	curParent := -1
	curSrc := -1
	abN := 0

	flush := func() {
		if cur != nil && len(cur.Lines) > 0 {
			abs = append(abs, *cur)
		}
		cur = nil
	}

	for _, it := range items {
		if cur == nil || it.parentBlockIdx != curParent || it.srcBlockIdx != curSrc {
			flush()
			abN++
			curParent = it.parentBlockIdx
			curSrc = it.srcBlockIdx

			// Key bit: AB facs points to the *parent* block
			cur = &model.AB{
				XmlID: transcriptionAnonBlockID(pageKey, abN),
				Type:  transcriptionAnonBlockType,
				Facs:  "#" + facZoneBlockID(pageKey, curParent+1),
			}
		}
		cur.Lines = append(cur.Lines, it.line)
	}
	flush()

	transDiv := model.Div{
		Type: "transcription",
		N:    pageKey,
		Abs:  abs,
	}
	doc.Text.Body.Divs = append(doc.Text.Body.Divs, transDiv)

	return doc, nil
}
