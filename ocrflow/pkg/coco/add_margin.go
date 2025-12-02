package coco

import "errors"

type AddMargin struct {
	Side     Side
	Margin   float64
	Category string
}

type addMarginBuilder struct {
	addMargin *AddMargin
}

func AddMarginBuilder() *addMarginBuilder {
	return &addMarginBuilder{
		addMargin: &AddMargin{},
	}
}

func (b *addMarginBuilder) Side(side Side) *addMarginBuilder {
	b.addMargin.Side = side
	return b
}

func (b *addMarginBuilder) Margin(margin float64) *addMarginBuilder {
	b.addMargin.Margin = margin
	return b
}

func (b *addMarginBuilder) Category(category string) *addMarginBuilder {
	b.addMargin.Category = category
	return b
}

func (b *addMarginBuilder) Build() (*AddMargin, error) {
	if b.addMargin.Side == "" {
		return nil, errors.New("side is required")
	}
	if b.addMargin.Margin <= 0 {
		return nil, errors.New("margin must be greater than 0")
	}
	if b.addMargin.Category == "" {
		return nil, errors.New("category is required")
	}
	return b.addMargin, nil
}

func ApplyAddMargin(c *Root, am *AddMargin) error {
	if c == nil {
		return errors.New("root is nil")
	}
	if am == nil {
		return errors.New("add margin is nil")
	}
	if am.Side == "" {
		return errors.New("side is required")
	}
	if am.Margin <= 0 {
		return errors.New("margin must be greater than 0")
	}
	if am.Category == "" {
		return errors.New("category is required")
	}

	// category name -> id
	catByName := make(map[string]int, len(c.Categories))
	for _, cat := range c.Categories {
		catByName[cat.Name] = cat.ID
	}

	catID, ok := catByName[am.Category]
	if !ok {
		return errors.New("category not found in root categories: " + am.Category)
	}

	// image id -> image for clamping
	imgByID := make(map[int]Image, len(c.Images))
	for _, img := range c.Images {
		imgByID[img.ID] = img
	}

	for i := range c.Annotations {
		a := &c.Annotations[i]
		if a.CategoryID != catID {
			continue
		}
		if len(a.BBox) < 4 {
			continue
		}

		x, y, w, h := a.BBox[0], a.BBox[1], a.BBox[2], a.BBox[3]
		if w <= 0 || h <= 0 {
			continue
		}

		newX, newY, newW, newH := x, y, w, h

		if am.Side == SideLeft || am.Side == SideRight {
			newW = w + am.Margin
		}
		if am.Side == SideHorizontal || am.Side == SideAll {
			newW = w + 2*am.Margin
		}
		if am.Side == SideTop || am.Side == SideBottom {
			newH = h + am.Margin
		}
		if am.Side == SideVertical || am.Side == SideAll {
			newH = h + 2*am.Margin
		}
		if am.Side == SideLeft || am.Side == SideHorizontal || am.Side == SideAll {
			newX = x - am.Margin
		}
		if am.Side == SideTop || am.Side == SideVertical || am.Side == SideAll {
			newY = y - am.Margin
		}

		// Clamp to image bounds if we know them
		if img, ok := imgByID[a.ImageID]; ok {
			right := newX + newW
			bottom := newY + newH

			if newX < 0 {
				newX = 0
			}
			if newY < 0 {
				newY = 0
			}
			if right > float64(img.Width) {
				right = float64(img.Width)
			}
			if bottom > float64(img.Height) {
				bottom = float64(img.Height)
			}

			newW = right - newX
			newH = bottom - newY
		}

		// Skip if clamping produced a degenerate box
		if newW <= 0 || newH <= 0 {
			continue
		}

		a.BBox[0] = newX
		a.BBox[1] = newY
		a.BBox[2] = newW
		a.BBox[3] = newH
		a.Area = newW * newH
	}

	return nil
}
