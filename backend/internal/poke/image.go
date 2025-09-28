package poke

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// ToSilhouette converts an image to a black silhouette with transparent background
func ToSilhouette(src image.Image) ([]byte, error) {
	b := src.Bounds()
	dst := image.NewNRGBA(b)
	draw.Draw(dst, b, image.Transparent, image.Point{}, draw.Src)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a > 0 { // if not fully transparent
				dst.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
			}
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
