package mpo

import (
	"errors"
	"fmt"
	"image"
	"image/color"
)

type colorType int

const (
	// RedCyan is Red on left eye, cyan on right
	RedCyan colorType = iota

	// CyanRed is Cyan on left eye, red on right
	CyanRed

	// RedGreen is Red on left eye, green on right
	RedGreen

	// GreenRed is Green on left eye, red on right
	GreenRed
)

// ErrInvalidImageCount indicates that incorrect number of images were found
// during the anaglyph conversion process.
var ErrInvalidImageCount = errors.New("anaglyph conversion only supports 2 images")

// ErrInconsistentBounds indicates that not all images within the MPO file were
// found to be the same size, which is a requirement for the anaglyph conversion.
var ErrInconsistentBounds = errors.New("anaglyph images must be the same size")

// ErrUnsupportedColorType indicates that the color type requested is not
// supported by the anaglyph conversion process.
var ErrUnsupportedColorType = errors.New("unsupported color type")

// ConvertToAnaglyph converts an MPO to the anaglyph format specified by ct colorType constant
// and returns the resulting image.
//
// ErrInconsistentBounds is returned if the images within the MPO are not the same size.
// ErrInvalidImageCount is returned if the number of images in the MPO is not exactly 2.
// ErrUnsupportedColorType is returned if the color type requested is not supported.
func (m *MPO) ConvertToAnaglyph(ct colorType) (image.Image, error) {
	if len(m.Image) != 2 {
		return nil, ErrInvalidImageCount
	}

	left := m.Image[0]
	right := m.Image[1]

	b := left.Bounds()

	if !left.Bounds().Eq(right.Bounds()) {
		return nil, ErrInconsistentBounds
	}

	img := image.NewRGBA(b)

	for x := b.Min.X; x < b.Max.X; x++ {
		for y := b.Min.Y; y < b.Max.Y; y++ {
			lr, lg, lb, _ := left.At(x, y).RGBA()
			rr, rg, rb, _ := right.At(x, y).RGBA()

			lgs := (((float32(lr) / 65535) * .229) * 65535) +
				(((float32(lg) / 65535) * .587) * 65535) +
				(((float32(lb) / 65535) * .144) * 65535)

			rgs := (((float32(rr) / 65535) * .229) * 65535) +
				(((float32(rg) / 65535) * .587) * 65535) +
				(((float32(rb) / 65535) * .144) * 65535)

			var c color.RGBA64
			switch ct {
			case RedCyan:
				c = color.RGBA64{
					R: uint16(lgs),
					G: uint16(rg),
					B: uint16(rb),
					A: 65535,
				}
			case CyanRed:
				c = color.RGBA64{
					R: uint16(rgs),
					G: uint16(lg),
					B: uint16(lb),
					A: 65535,
				}
			case RedGreen:
				c = color.RGBA64{
					R: uint16(lgs),
					G: uint16(rgs),
					B: 65535 / 2,
					A: 65535,
				}
			case GreenRed:
				c = color.RGBA64{
					R: uint16(rgs),
					G: uint16(lgs),
					B: 65535 / 2,
					A: 65535,
				}
			default:
				return nil, fmt.Errorf("unsupported color type %d: %w", ct, ErrUnsupportedColorType)
			}

			img.Set(x, y, c)
		}
	}

	return img, nil
}
