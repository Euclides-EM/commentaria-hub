package formatcov

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"math"
	"sort"

	"github.com/MiaMish/elements-dh/ocrflow/pkg/alto"
	"github.com/MiaMish/elements-dh/ocrflow/pkg/rbmodel"
)

func Roboflow2ALTO(jsonStr string, imageFileName, pageID string) ([]byte, error) {
	var det rbmodel.DetectionFile
	if err := json.Unmarshal([]byte(jsonStr), &det); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	// Build tag map (class -> BTxxx)
	classSet := map[string]struct{}{}
	for _, p := range det.Predictions {
		classSet[p.Class] = struct{}{}
	}

	classes := make([]string, 0, len(classSet))
	for c := range classSet {
		classes = append(classes, c)
	}
	sort.Strings(classes)

	classToBT := make(map[string]string, len(classes))
	otherTags := make([]alto.OtherTag, 0, len(classes))

	for i, c := range classes {
		id := fmt.Sprintf("BT%03d", i)
		classToBT[c] = id
		otherTags = append(otherTags, alto.OtherTag{
			ID:          id,
			Label:       c,
			Description: "block type " + c,
		})
	}

	// Build TextBlocks
	textBlocks := make([]alto.TextBlock, 0, len(det.Predictions))
	for i, p := range det.Predictions {
		left := int(math.Round(p.X - p.Width/2.0))
		top := int(math.Round(p.Y - p.Height/2.0))
		right := int(math.Round(p.X + p.Width/2.0))
		bottom := int(math.Round(p.Y + p.Height/2.0))

		// It assumes x,y in the JSON are center coordinates of the box. If in your data they are top-left instead, change:
		// left := int(math.Round(p.X))
		//top := int(math.Round(p.Y))
		//right := int(math.Round(p.X + p.Width))
		//bottom := int(math.Round(p.Y + p.Height))

		points := fmt.Sprintf(
			"%d %d %d %d %d %d %d %d %d %d",
			left, top,
			right, top,
			right, bottom,
			left, bottom,
			left, top,
		)

		tb := alto.TextBlock{
			HPOS:    float64(left),
			VPOS:    float64(top),
			Width:   float64(right - left),
			Height:  float64(bottom - top),
			ID:      fmt.Sprintf("blck%03d", i),
			TagRefs: classToBT[p.Class],
			Shape: alto.Shape{
				Polygon: alto.Polygon{
					Points: points,
				},
			},
		}

		textBlocks = append(textBlocks, tb)
	}

	altoRes := alto.Alto{
		Xmlns:          "http://www.loc.gov/standards/alto/ns-v4#",
		XmlnsXsi:       "http://www.w3.org/2001/XMLSchema-instance",
		SchemaLocation: "http://www.loc.gov/standards/alto/ns-v4# http://www.loc.gov/standards/alto/v4/alto-4-2.xsd",
		Description: alto.Description{
			MeasurementUnit: "pixel",
			SourceImageInformation: alto.SourceImageInformation{
				FileName: imageFileName,
			},
		},
		Tags: alto.Tags{
			OtherTags: otherTags,
		},
		Layout: alto.Layout{
			Page: []alto.Page{
				{
					Width:  det.Image.Width,
					Height: det.Image.Height,
					ID:     pageID,
					PrintSpace: alto.PrintSpace{
						HPOS:       0,
						VPOS:       0,
						Width:      float64(det.Image.Width),
						Height:     float64(det.Image.Height),
						TextBlocks: textBlocks,
					},
				},
			},
		},
	}

	out, err := xml.MarshalIndent(altoRes, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("marshal alto: %w", err)
	}

	// Prepend XML header manually
	final := append([]byte(xml.Header), out...)
	return final, nil
}
