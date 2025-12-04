package formatcov

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
)

func JPGToPNG(srcPath, dstPath string) error {
	// open source JPG
	in, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()

	// decode JPG
	img, err := jpeg.Decode(in)
	if err != nil {
		return fmt.Errorf("decode jpg: %w", err)
	}

	// create destination PNG file
	out, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer out.Close()

	// encode as PNG
	if err := png.Encode(out, img); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}

	return nil
}
