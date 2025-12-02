package formatcov

import (
	"encoding/json"
	"fmt"
	coco2 "github.com/MiaMish/elements-dh/ocrflow/pkg/coco"
	"sort"
	"strings"
	"time"
)

type roboflowResult struct {
	InferenceID string  `json:"inference_id"`
	Time        float64 `json:"time"`
	Image       struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"image"`
	Predictions []struct {
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Width      float64 `json:"width"`
		Height     float64 `json:"height"`
		Confidence float64 `json:"confidence"`
		Class      string  `json:"class"`
		ClassID    int     `json:"class_id"`
		DetectID   string  `json:"detection_id"`
	} `json:"predictions"`
}

// Roboflow2Coco converts one or more Roboflow JSON strings into a COCO JSON string.
func Roboflow2Coco(imageNameToRB map[string]string, categories []string) (string, error) {
	var coco coco2.Root

	catMap := make(map[int]coco2.Category) // class_id -> category
	nextImageID := 1
	nextAnnID := 1

	coco.Info = coco2.Info{
		Description: fmt.Sprintf("Converted from Roboflow format with %d images", len(imageNameToRB)),
		Version:     "1.0",
		Year:        2025,
		Contributor: "ocrflow",
		DateCreated: time.Now().Format("2006/01/02"),
	}

	coco.Licenses = []coco2.License{
		{
			URL:  "http://creativecommons.org/licenses/by-nc/2.0/",
			ID:   1,
			Name: "Attribution-NonCommercial License",
		},
	}

	for imgFile, js := range imageNameToRB {
		if strings.TrimSpace(js) == "" {
			continue
		}

		var rf roboflowResult
		if err := json.Unmarshal([]byte(js), &rf); err != nil {
			return "", fmt.Errorf("decode roboflow json: %w", err)
		}

		imageID := nextImageID
		nextImageID++

		fileName := imgFile
		if rf.InferenceID == "" {
			fileName = fmt.Sprintf("image_%d.jpg", imageID)
		}

		coco.Images = append(coco.Images, coco2.Image{
			ID:       imageID,
			FileName: fileName,
			Width:    rf.Image.Width,
			Height:   rf.Image.Height,
		})

		for _, p := range rf.Predictions {
			// Register category if new
			if _, ok := catMap[p.ClassID]; !ok {
				catMap[p.ClassID] = coco2.Category{
					ID:   p.ClassID,
					Name: p.Class,
				}
			}

			// Roboflow gives center-based boxes, COCO wants top-left
			xMin := p.X - p.Width/2.0
			yMin := p.Y - p.Height/2.0
			if xMin < 0 {
				xMin = 0
			}
			if yMin < 0 {
				yMin = 0
			}

			area := p.Width * p.Height

			coco.Annotations = append(coco.Annotations, coco2.Annotation{
				ID:         nextAnnID,
				ImageID:    imageID,
				CategoryID: p.ClassID,
				BBox:       []float64{xMin, yMin, p.Width, p.Height},
				Area:       area,
				IsCrowd:    0,
				Score:      p.Confidence,
			})
			nextAnnID++
		}
	}

	// if categories are provided, make sure they are included in catMap
	if len(categories) > 0 {
		for id, name := range categories {
			found := false
			for _, cat := range catMap {
				if cat.Name == name {
					found = true
					if cat.ID != id {
						return "", fmt.Errorf("category ID/name mismatch: ID %d is '%s' but existing ID %d is '%s'", id+1, name, cat.ID, cat.Name)
					}
					break
				}
			}
			if !found {
				catMap[id] = coco2.Category{
					ID:   id,
					Name: name,
				}
			}
		}
	}
	// Flatten categories in stable order
	if len(catMap) > 0 {
		var ids []int
		for id := range catMap {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		for _, id := range ids {
			coco.Categories = append(coco.Categories, catMap[id])
		}
	}

	out, err := json.MarshalIndent(coco, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal coco json: %w", err)
	}

	return string(out), nil
}
