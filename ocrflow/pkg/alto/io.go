package alto

import (
	"encoding/xml"
	"fmt"
	"math"
	"os"
)

func LoadFromFile(path string) (*Alto, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ALTO: %w", err)
	}
	var a Alto
	if err := xml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("unmarshal ALTO: %w", err)
	}
	if a.Xmlns == "" {
		a.Xmlns = "http://www.loc.gov/standards/alto/ns-v4#"
	}
	if a.XmlnsXsi == "" {
		a.XmlnsXsi = "http://www.w3.org/2001/XMLSchema-instance"
	}
	if a.SchemaLocation == "" {
		a.SchemaLocation = "http://www.loc.gov/standards/alto/ns-v4# http://www.loc.gov/standards/alto/v4/alto-4-3.xsd"
	}
	EnsureLineBoundaries(&a)
	return &a, nil
}

// EnsureLineBoundaries sets Shape.Polygon.Points from the line bbox for any TextLine
// that has no boundary. eScriptorium requires a boundary per line for extraction.
func EnsureLineBoundaries(a *Alto) {
	for pi := range a.Layout.Page {
		ps := &a.Layout.Page[pi].PrintSpace
		for tbi := range ps.TextBlocks {
			tb := &ps.TextBlocks[tbi]
			for li := range tb.Lines {
				tl := &tb.Lines[li]
				if tl.Shape.Polygon.Points != "" {
					continue
				}
				if tl.Width < 0 || tl.Height < 0 {
					continue
				}
				minX, minY := tl.HPOS, tl.VPOS
				maxX := minX + tl.Width
				maxY := minY + tl.Height
				tl.Shape.Polygon.Points = rectToPointsString(minX, minY, maxX, maxY)
			}
		}
	}
}

func rectToPointsString(minX, minY, maxX, maxY float64) string {
	x1, y1 := int(math.Round(minX)), int(math.Round(minY))
	x2, y2 := int(math.Round(maxX)), int(math.Round(maxY))
	return fmt.Sprintf("%d %d %d %d %d %d %d %d %d %d",
		x1, y1, x2, y1, x2, y2, x1, y2, x1, y1)
}

func SaveToFile(af *Alto, path string) error {
	data, err := xml.MarshalIndent(af, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal ALTO: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write ALTO: %w", err)
	}
	return nil
}
