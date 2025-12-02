package coco

import (
	"errors"
	"fmt"
)

type StretchTowardsCategory struct {
	StretchCategory string
	Towards         string
	ContactType     ContactType
	ContactSide     Side
}

type stretchTowardsCategoryBuilder struct {
	stretchTowardsCategory *StretchTowardsCategory
}

func (b *stretchTowardsCategoryBuilder) Stretch(category string) *stretchTowardsCategoryBuilder {
	b.stretchTowardsCategory.StretchCategory = category
	return b
}

func (b *stretchTowardsCategoryBuilder) Towards(towards string, contactType ContactType, contactSide Side) *stretchTowardsCategoryBuilder {
	b.stretchTowardsCategory.Towards = towards
	b.stretchTowardsCategory.ContactType = contactType
	b.stretchTowardsCategory.ContactSide = contactSide
	return b
}

func (b *stretchTowardsCategoryBuilder) Build() (*StretchTowardsCategory, error) {
	if b.stretchTowardsCategory.StretchCategory == "" {
		return nil, errors.New("stretch category is required")
	}
	if b.stretchTowardsCategory.Towards == "" {
		return nil, errors.New("towards is required")
	}
	if b.stretchTowardsCategory.ContactType == "" {
		return nil, errors.New("contact type is required")
	}
	if b.stretchTowardsCategory.ContactSide == "" {
		return nil, errors.New("contact side is required")
	}
	return b.stretchTowardsCategory, nil
}

func StretchTowardsCategoryBuilder() *stretchTowardsCategoryBuilder {
	return &stretchTowardsCategoryBuilder{
		stretchTowardsCategory: &StretchTowardsCategory{},
	}
}

func ApplyStretchTowardsCategory(c *Root, stc *StretchTowardsCategory) error {
	// Build category name -> ID map
	catByName := make(map[string]int, len(c.Categories))
	for _, cat := range c.Categories {
		catByName[cat.Name] = cat.ID
	}

	srcCatID, ok := catByName[stc.StretchCategory]
	if !ok {
		return fmt.Errorf("stretch category %q not found in categories", stc.StretchCategory)
	}

	tgtCatID, ok := catByName[stc.Towards]
	if !ok {
		return fmt.Errorf("towards category %q not found in categories", stc.Towards)
	}

	byImage := make(map[int]*imgAnns)

	// We need pointers, but c.Annotations is a slice of values.
	for i := range c.Annotations {
		a := &c.Annotations[i]
		entry, ok := byImage[a.ImageID]
		if !ok {
			entry = &imgAnns{}
			byImage[a.ImageID] = entry
		}
		switch a.CategoryID {
		case srcCatID:
			entry.src = append(entry.src, a)
		case tgtCatID:
			entry.tgt = append(entry.tgt, a)
		}
	}

	for imgID, entry := range byImage {
		if entry == nil || len(entry.src) == 0 {
			continue
		}

		for _, src := range entry.src {
			tgt := chooseTarget(src, entry, stc.ContactSide, stc.ContactType)
			if tgt == nil {
				// Nothing to stretch towards for this image.
				continue
			}

			if len(src.BBox) < 4 || len(tgt.BBox) < 4 {
				// Skip malformed bboxes.
				continue
			}

			sx, sy, sw, sh := src.BBox[0], src.BBox[1], src.BBox[2], src.BBox[3]
			tx, ty, tw, th := tgt.BBox[0], tgt.BBox[1], tgt.BBox[2], tgt.BBox[3]

			if sw <= 0 || sh <= 0 || tw <= 0 || th <= 0 {
				continue
			}

			newX, newY, newW, newH := sx, sy, sw, sh

			switch stc.ContactType {
			case ContactTypeInner:
				switch stc.ContactSide {
				case SideLeft:
					// Move left edge of src to left edge of tgt, keep right edge fixed.
					right := sx + sw
					newX = tx
					newW = right - newX
				case SideRight:
					// Move right edge of src to right edge of tgt, keep left edge fixed.
					right := tx + tw
					newW = right - sx
				case SideTop:
					// Move top edge of src to top edge of tgt, keep bottom edge fixed.
					bottom := sy + sh
					newY = ty
					newH = bottom - newY
				case SideBottom:
					// Move bottom edge of src to bottom edge of tgt, keep top edge fixed.
					bottom := ty + th
					newH = bottom - sy
				default:
					return fmt.Errorf("unsupported contact side %q for image %d", stc.ContactSide, imgID)
				}
			case ContactTypeOuter:
				switch stc.ContactSide {
				case SideLeft:
					// Make the right edge of src coincide with the left edge of tgt, keep left edge fixed.
					newW = tx - sx
				case SideRight:
					// Make the left edge of src coincide with the right edge of tgt, keep right edge fixed.
					right := sx + sw
					left := tx + tw
					newW = right - left
					newX = left
				case SideTop:
					// Make the bottom edge of src coincide with the top edge of tgt, keep top edge fixed.
					newH = ty - sy
				case SideBottom:
					// Make the top edge of src coincide with the bottom edge of tgt, keep bottom edge fixed.
					bottom := sy + sh
					top := ty + th
					newH = bottom - top
					newY = top
				default:
					return fmt.Errorf("unsupported contact side %q for image %d", stc.ContactSide, imgID)
				}
			default:
				return fmt.Errorf("unsupported contact type %q", stc.ContactType)
			}

			// Guard against invalid geometry.
			if newW <= 0 || newH <= 0 {
				// If result is invalid, skip modifying this annotation.
				continue
			}

			src.BBox[0] = newX
			src.BBox[1] = newY
			src.BBox[2] = newW
			src.BBox[3] = newH
			src.Area = newW * newH
		}
	}

	return nil
}

// Group annotations by image and by category for quick lookup.
type imgAnns struct {
	src []*Annotation
	tgt []*Annotation
}

// intervalsIntersect reports whether [a1,a2] and [b1,b2] intersect
// with positive length on one axis.
func intervalsIntersect(a1, a2, b1, b2 float64) bool {
	return a1 < b2 && b1 < a2
}

// chooseTarget picks the best target for a given src, side and contact type.
//
// For ContactTypeInner:
//   - SideTop:    target spans src's top (tTop < sTop < tBottom) and overlaps on x
//   - SideBottom: target spans src's bottom (tTop < sBottom < tBottom) and overlaps on x
//   - SideLeft:   target spans src's left  (tLeft < sLeft < tRight) and overlaps on y
//   - SideRight:  target spans src's right (tLeft < sRight < tRight) and overlaps on y
//     Among all that match, pick the closest along the relevant axis.
//
// For ContactTypeOuter:
//   - SideTop:    target is strictly above or touching src (tBottom <= sTop) and overlaps on x
//   - SideBottom: target is strictly below or touching src (tTop >= sBottom) and overlaps on x
//   - SideLeft:   target is strictly to the left or touching (tRight <= sLeft) and overlaps on y
//   - SideRight:  target is strictly to the right or touching (tLeft >= sRight) and overlaps on y
//     Again pick the closest one along the relevant axis.
func chooseTarget(src *Annotation, entry *imgAnns, side Side, contactType ContactType) *Annotation {
	if entry == nil || len(entry.tgt) == 0 || src == nil {
		return nil
	}
	if len(src.BBox) < 4 {
		return nil
	}

	sx, sy, sw, sh := src.BBox[0], src.BBox[1], src.BBox[2], src.BBox[3]
	if sw <= 0 || sh <= 0 {
		return nil
	}
	sLeft, sTop := sx, sy
	sRight, sBottom := sx+sw, sy+sh

	var best *Annotation
	bestDist := 0.0
	hasBest := false

	for _, tgt := range entry.tgt {
		if tgt == nil || len(tgt.BBox) < 4 {
			continue
		}

		tx, ty, tw, th := tgt.BBox[0], tgt.BBox[1], tgt.BBox[2], tgt.BBox[3]
		if tw <= 0 || th <= 0 {
			continue
		}
		tLeft, tTop := tx, ty
		tRight, tBottom := tx+tw, ty+th

		match := false
		dist := 0.0

		switch contactType {
		case ContactTypeInner:
			switch side {
			case SideTop:
				// Your example:
				// - src top between target top and bottom
				// - x ranges intersect
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop < sTop && sTop < tBottom {
					match = true
					// distance along vertical axis (want smallest sTop - tTop > 0)
					dist = sTop - tTop
					if dist < 0 {
						match = false
					}
				}
			case SideBottom:
				// target spans src bottom
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop < sBottom && sBottom < tBottom {
					match = true
					dist = tBottom - sBottom
					if dist < 0 {
						match = false
					}
				}
			case SideLeft:
				// target spans src left
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft < sLeft && sLeft < tRight {
					match = true
					dist = sLeft - tLeft
					if dist < 0 {
						match = false
					}
				}
			case SideRight:
				// target spans src right
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft < sRight && sRight < tRight {
					match = true
					dist = tRight - sRight
					if dist < 0 {
						match = false
					}
				}
			}

		case ContactTypeOuter:
			switch side {
			case SideTop:
				// target above or touching src, overlap on x
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tBottom <= sTop {
					match = true
					// distance from src top down to target bottom
					dist = sTop - tBottom
					if dist < 0 {
						match = false
					}
				}
			case SideBottom:
				// target below or touching src, overlap on x
				if intervalsIntersect(sLeft, sRight, tLeft, tRight) &&
					tTop >= sBottom {
					match = true
					dist = tTop - sBottom
					if dist < 0 {
						match = false
					}
				}
			case SideLeft:
				// target to the left or touching, overlap on y
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tRight <= sLeft {
					match = true
					dist = sLeft - tRight
					if dist < 0 {
						match = false
					}
				}
			case SideRight:
				// target to the right or touching, overlap on y
				if intervalsIntersect(sTop, sBottom, tTop, tBottom) &&
					tLeft >= sRight {
					match = true
					dist = tLeft - sRight
					if dist < 0 {
						match = false
					}
				}
			}
		}

		if !match {
			continue
		}

		if !hasBest || dist < bestDist {
			hasBest = true
			bestDist = dist
			best = tgt
		}
	}

	return best
}
