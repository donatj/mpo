package mpo

import (
	"image"
	"image/draw"
	"math"
)

func (m *Mpo) ConvertToStereo() image.Image {
	mx := 0
	my := 0
	for _, i := range m.Images {
		mx += i.Bounds().Max.X
		my = int(math.Max(float64(my), float64(i.Bounds().Max.Y)))
	}

	img := image.NewRGBA(image.Rect(0, 0, mx, my))

	dx := 0
	for _, i := range m.Images {
		b := i.Bounds()
		b = b.Add(image.Point{dx, 0})

		draw.Draw(img, b, i, image.Point{0, 0}, draw.Src)

		dx += i.Bounds().Max.X
	}

	return img
}
