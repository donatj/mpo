package mpo

import (
	"errors"
	"image"
	"image/color"
)

type colorType int

const (
	RedCyan colorType = iota
	CyanRed

	RedGreen
	GreenRed
)

// Converts an MPO to the anaglyph format specified by ct colorType constant
func (m *MPO) ConvertToAnaglyph(ct colorType) (image.Image, error) {
	if len(m.Image) != 2 {
		return nil, errors.New("anaglph conversion only supports 2 image")
	}

	left := m.Image[0]
	right := m.Image[1]

	if !left.Bounds().Eq(right.Bounds()) {
		return nil, errors.New("anaglyph images must be the same size")
	}

	img := image.NewRGBA(left.Bounds())

	for x := 0; x <= left.Bounds().Max.X; x++ {
		for y := 0; y <= left.Bounds().Max.Y; y++ {
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
				return nil, errors.New("Unsupported color type")
			}

			img.Set(x, y, c)
		}
	}

	return img, nil
}
