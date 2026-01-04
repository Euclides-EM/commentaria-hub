package geo

import "math"

type Rectangle struct {
	MinX float64
	MinY float64
	MaxX float64
	MaxY float64
}

func (r *Rectangle) Width() float64  { return math.Max(0, r.MaxX-r.MinX) }
func (r *Rectangle) Height() float64 { return math.Max(0, r.MaxY-r.MinY) }
func (r *Rectangle) Area() float64   { return r.Width() * r.Height() }

func (r *Rectangle) Expand(p float64) *Rectangle {
	return &Rectangle{
		MinX: r.MinX - p,
		MinY: r.MinY - p,
		MaxX: r.MaxX + p,
		MaxY: r.MaxY + p,
	}
}

func (r *Rectangle) Intersect(b *Rectangle) *Rectangle {
	return &Rectangle{
		MinX: math.Max(r.MinX, b.MinX),
		MinY: math.Max(r.MinY, b.MinY),
		MaxX: math.Min(r.MaxX, b.MaxX),
		MaxY: math.Min(r.MaxY, b.MaxY),
	}
}

func (r *Rectangle) Contains(inner *Rectangle) bool {
	return inner.MinX >= r.MinX &&
		inner.MinY >= r.MinY &&
		inner.MaxX <= r.MaxX &&
		inner.MaxY <= r.MaxY
}

// OverlapRatio returns intersection area / line area.
// If line area is 0, returns 0.
func (r *Rectangle) OverlapRatio(block *Rectangle) float64 {
	ia := r.Intersect(block).Area()
	la := r.Area()
	if la <= 0 {
		return 0
	}
	return ia / la
}

func RectangleFromLeftBottomCorner(x, y, width, height float64) *Rectangle {
	return &Rectangle{
		MinX: x,
		MinY: y,
		MaxX: x + width,
		MaxY: y + height,
	}
}
