package mpo

import (
	"image"
	"image/draw"
)

// Converts an MPO to StereoScopic image
func (m *MPO) ConvertToStereo() image.Image {
	mx := 0
	my := 0
	for _, i := range m.Image {
		mx += i.Bounds().Max.X
		if i.Bounds().Max.Y > my {
			my = i.Bounds().Max.Y
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, mx, my))

	dx := 0
	for _, i := range m.Image {
		b := i.Bounds()
		b = b.Add(image.Point{dx, 0})

		draw.Draw(img, b, i, image.Point{0, 0}, draw.Src)

		dx += i.Bounds().Max.X
	}

	return img
}
