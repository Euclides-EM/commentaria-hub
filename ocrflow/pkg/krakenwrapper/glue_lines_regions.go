package krakenwrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

// ---------- JSON structures ----------

type BaselineDoc struct {
	Type          string          `json:"type"`
	ImageName     string          `json:"imagename"`
	TextDirection string          `json:"text_direction"`
	ScriptDetect  bool            `json:"script_detection"`
	Lines         []Line          `json:"lines"`
	Regions       json.RawMessage `json:"regions"` // not used here
	LineOrders    []interface{}   `json:"line_orders"`
}

type Line struct {
	ID       string            `json:"id"`
	Baseline [][]float64       `json:"baseline"`
	Boundary [][]float64       `json:"boundary"`
	Text     *string           `json:"text"`
	Tags     map[string]string `json:"tags"`
	Regions  []string          `json:"regions"`
	Type     string            `json:"type"`
	Image    *string           `json:"imagename"`
	Split    interface{}       `json:"split"`
	BaseDir  interface{}       `json:"base_dir"`
}

// ---------- ALTO helpers ----------

type textBlockInfo struct {
	elem *etree.Element
	x    float64
	y    float64
	w    float64
	h    float64
}

func (tb textBlockInfo) containsPoint(px, py float64) bool {
	return px >= tb.x && px <= tb.x+tb.w && py >= tb.y && py <= tb.y+tb.h
}

// compute bbox from a list of [x,y] points
func bbox(points [][]float64) (minX, minY, maxX, maxY float64) {
	minX, minY = math.MaxFloat64, math.MaxFloat64
	maxX, maxY = -math.MaxFloat64, -math.MaxFloat64

	for _, p := range points {
		if len(p) != 2 {
			continue
		}
		x, y := p[0], p[1]
		if x < minX {
			minX = x
		}
		if y < minY {
			minY = y
		}
		if x > maxX {
			maxX = x
		}
		if y > maxY {
			maxY = y
		}
	}

	if minX == math.MaxFloat64 {
		minX, minY, maxX, maxY = 0, 0, 0, 0
	}
	return
}

func pointsToString(points [][]float64) string {
	var sb strings.Builder
	first := true
	for _, p := range points {
		if len(p) != 2 {
			continue
		}
		if !first {
			sb.WriteByte(' ')
		}
		first = false
		sb.WriteString(strconv.Itoa(int(math.Round(p[0]))))
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(int(math.Round(p[1]))))
	}
	return sb.String()
}

func baselineToString(points [][]float64) string {
	return pointsToString(points)
}

// bboxToPointsString returns POINTS for a closed rectangle (eScriptorium needs a boundary per line).
func bboxToPointsString(minX, minY, maxX, maxY float64) string {
	return fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d",
		int(math.Round(minX)), int(math.Round(minY)),
		int(math.Round(maxX)), int(math.Round(minY)),
		int(math.Round(maxX)), int(math.Round(maxY)),
		int(math.Round(minX)), int(math.Round(maxY)),
		int(math.Round(minX)), int(math.Round(minY)))
}

func floatAttr(elem *etree.Element, name string) (float64, error) {
	attr := elem.SelectAttr(name)
	if attr == nil {
		return 0, fmt.Errorf("missing attribute %s", name)
	}
	v, err := strconv.ParseFloat(attr.Value, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s=%q: %w", name, attr.Value, err)
	}
	return v, nil
}

// ---------- main glue logic ----------

func GlueLinesToAlto(altoPath, baselinesJsonPath, outPath string) error {
	// Load JSON
	jf, err := os.ReadFile(baselinesJsonPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", baselinesJsonPath, err)
	}
	// Empty file: Kraken wrote nothing (no lines in mask region, or failed silently). Treat as no lines to glue.
	if len(jf) == 0 {
		log.Printf("skipping glue: %s is empty (Kraken found no lines or failed; mask may not be bitonal)", baselinesJsonPath)
		return nil
	}
	var doc BaselineDoc
	if err := json.Unmarshal(jf, &doc); err != nil {
		return fmt.Errorf("parsing %s: %w (Kraken may have failed; check mask is bitonal)", baselinesJsonPath, err)
	}

	// Load ALTO with etree
	altodoc := etree.NewDocument()
	if err := altodoc.ReadFromFile(altoPath); err != nil {
		return fmt.Errorf("reading %s: %w", altoPath, err)
	}

	// Collect TextBlocks with bounding boxes
	var blocks []textBlockInfo
	for _, tb := range altodoc.FindElements("//TextBlock") {
		x, err1 := floatAttr(tb, "HPOS")
		y, err2 := floatAttr(tb, "VPOS")
		w, err3 := floatAttr(tb, "WIDTH")
		h, err4 := floatAttr(tb, "HEIGHT")
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			// skip malformed blocks
			continue
		}
		blocks = append(blocks, textBlockInfo{
			elem: tb,
			x:    x,
			y:    y,
			w:    w,
			h:    h,
		})
	}

	if len(blocks) == 0 {
		return fmt.Errorf("no TextBlock elements found in ALTO")
	}

	// Optionally ensure a line-type OtherTag exists (for default)
	ensureDefaultLineTag(altodoc)

	// For each line, compute bbox and assign to a TextBlock
	for _, ln := range doc.Lines {
		if len(ln.Boundary) == 0 && len(ln.Baseline) == 0 {
			continue
		}

		points := ln.Boundary
		if len(points) == 0 {
			points = ln.Baseline
		}
		minX, minY, maxX, maxY := bbox(points)
		cx := (minX + maxX) / 2
		cy := (minY + maxY) / 2

		// find containing block
		var target *textBlockInfo
		for i := range blocks {
			if blocks[i].containsPoint(cx, cy) {
				target = &blocks[i]
				break
			}
		}
		if target == nil {
			// if no containing block, attach to the first one as fallback
			target = &blocks[0]
		}

		// Build TextLine element
		tl := etree.NewElement("TextLine")
		tl.CreateAttr("ID", fmt.Sprintf("eSc_line_%s", shortID(ln.ID)))
		tl.CreateAttr("TAGREFS", "LT_default") // or whatever you want
		tl.CreateAttr("BASELINE", baselineToString(ln.Baseline))

		hpos := int(math.Round(minX))
		vpos := int(math.Round(minY))
		width := int(math.Round(maxX - minX))
		height := int(math.Round(maxY - minY))

		tl.CreateAttr("HPOS", strconv.Itoa(hpos))
		tl.CreateAttr("VPOS", strconv.Itoa(vpos))
		tl.CreateAttr("WIDTH", strconv.Itoa(width))
		tl.CreateAttr("HEIGHT", strconv.Itoa(height))

		// Shape / Polygon: use boundary if present, otherwise bbox so eScriptorium has a line region
		{
			shape := etree.NewElement("Shape")
			poly := etree.NewElement("Polygon")
			if len(ln.Boundary) > 0 {
				poly.CreateAttr("POINTS", pointsToString(ln.Boundary))
			} else {
				// eScriptorium requires a boundary for line extraction; use bbox as polygon
				poly.CreateAttr("POINTS", bboxToPointsString(minX, minY, maxX, maxY))
			}
			shape.AddChild(poly)
			tl.AddChild(shape)
		}

		// String with same bbox, empty CONTENT for now
		str := etree.NewElement("String")
		if ln.Text != nil {
			str.CreateAttr("CONTENT", *ln.Text)
		} else {
			str.CreateAttr("CONTENT", "")
		}
		str.CreateAttr("HPOS", strconv.Itoa(hpos))
		str.CreateAttr("VPOS", strconv.Itoa(vpos))
		str.CreateAttr("WIDTH", strconv.Itoa(width))
		str.CreateAttr("HEIGHT", strconv.Itoa(height))
		tl.AddChild(str)

		// append to block
		target.elem.AddChild(tl)
	}

	// Write result

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", outPath, err)
	}
	altodoc.Indent(2)
	if err := altodoc.WriteToFile(outPath); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	log.Printf("Wrote glued ALTO to %s", outPath)
	return nil
}

// ensureDefaultLineTag inserts an OtherTag for line type "default" if not present
func ensureDefaultLineTag(doc *etree.Document) {
	tags := doc.FindElement("//Tags")
	if tags == nil {
		return
	}
	// see if we already have something for "default" lines
	for _, ot := range tags.SelectElements("OtherTag") {
		label := ot.SelectAttrValue("LABEL", "")
		desc := ot.SelectAttrValue("DESCRIPTION", "")
		if strings.EqualFold(label, "default") && strings.Contains(desc, "line type") {
			return
		}
	}
	// if not found, create a simple one
	ot := etree.NewElement("OtherTag")
	ot.CreateAttr("ID", "LT_default")
	ot.CreateAttr("LABEL", "default")
	ot.CreateAttr("DESCRIPTION", "line type default")
	tags.AddChild(ot)
}

// shortID trims a UUID for nicer IDs, fallback to full id
func shortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
