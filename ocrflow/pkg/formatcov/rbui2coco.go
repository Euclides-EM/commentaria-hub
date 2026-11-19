package formatcov

import (
	"encoding/json"
	"fmt"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/coco"
	"strconv"
)

type rfBox struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Label      string      `json:"label"`
	X          string      `json:"x"`
	Y          string      `json:"y"`
	Width      string      `json:"width"`
	Height     string      `json:"height"`
	Points     [][]float64 `json:"points"`
	Confidence interface{} `json:"confidence"`
}

type rfInput struct {
	Boxes  []rfBox `json:"boxes"`
	Height int     `json:"height"`
	Width  int     `json:"width"`
	Key    string  `json:"key"`
}

func RoboflowUI2Coco(jsonStrs ...string) (string, error) {
	var result coco.Root

	// Track unique categories
	categoryIndex := map[string]int{}
	nextCategoryID := 1

	annotationID := 1
	imageID := 1

	for _, js := range jsonStrs {
		var rf rfInput
		if err := json.Unmarshal([]byte(js), &rf); err != nil {
			return "", fmt.Errorf("parse error: %w", err)
		}

		// Add image entry
		result.Images = append(result.Images, coco.Image{
			ID:     imageID,
			File:   rf.Key,
			Width:  rf.Width,
			Height: rf.Height,
		})

		// Convert each annotation
		for _, b := range rf.Boxes {
			// category
			catID, ok := categoryIndex[b.Label]
			if !ok {
				catID = nextCategoryID
				nextCategoryID++
				categoryIndex[b.Label] = catID
				result.Categories = append(result.Categories, coco.Category{
					ID:   catID,
					Name: b.Label,
				})
			}

			// parse numeric values
			x, _ := strconv.ParseFloat(b.X, 64)
			y, _ := strconv.ParseFloat(b.Y, 64)
			w, _ := strconv.ParseFloat(b.Width, 64)
			h, _ := strconv.ParseFloat(b.Height, 64)

			annot := coco.Annotation{
				ID:         annotationID,
				ImageID:    imageID,
				CategoryID: catID,
				BBox:       []float64{x, y, w, h},
				IsCrowd:    0,
			}

			// polygon case
			if len(b.Points) > 0 {
				var flat []float64
				for _, p := range b.Points {
					flat = append(flat, p[0], p[1])
				}
				annot.Segmentation = [][]float64{flat}
				annot.Area = w * h
			} else {
				annot.Area = w * h
			}

			result.Annotations = append(result.Annotations, annot)
			annotationID++
		}

		imageID++
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(out), nil
}
