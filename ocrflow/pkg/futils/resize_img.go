package futils

import (
	"image"
	"image/color"
)

func ResizeImageToMaxSide(src image.Image, maxSide int) image.Image {
	bounds := src.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= 0 || srcH <= 0 || maxSide <= 0 {
		return src
	}

	longestSide := max(srcW, srcH)
	scale := 1.0
	if longestSide > maxSide {
		scale = float64(maxSide) / float64(longestSide)
	}

	dstW := max(1, int(float64(srcW)*scale+0.5))
	dstH := max(1, int(float64(srcH)*scale+0.5))
	if dstW == srcW && dstH == srcH {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		srcY := bounds.Min.Y + y*srcH/dstH
		for x := 0; x < dstW; x++ {
			srcX := bounds.Min.X + x*srcW/dstW
			dst.Set(x, y, flattenOnWhite(src.At(srcX, srcY)))
		}
	}
	return dst
}

func flattenOnWhite(c color.Color) color.Color {
	r16, g16, b16, a16 := c.RGBA()
	if a16 == 0xffff {
		return color.RGBA{
			R: uint8(r16 >> 8),
			G: uint8(g16 >> 8),
			B: uint8(b16 >> 8),
			A: 0xff,
		}
	}
	if a16 == 0 {
		return color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	}

	a := float64(a16) / 65535.0
	blend := func(v uint32) uint8 {
		fg := float64(v) / 65535.0
		return uint8((fg*a+(1-a))*255.0 + 0.5)
	}

	return color.RGBA{
		R: blend(r16),
		G: blend(g16),
		B: blend(b16),
		A: 0xff,
	}
}
