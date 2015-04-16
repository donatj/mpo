package mpo

import (
	"errors"
	"image"
	"image/color"
)

type colorType int

const (
	REDBLUE colorType = iota
	BLUERED

	REDGREEN
	GREENRED
)

func (m *Mpo) ConvertToAnaglyph(colType colorType) (image.Image, error) {
	if len(m.Images) != 2 {
		return nil, errors.New("Anaglph conversion only supports 2 image MPOs")
	}

	left := m.Images[0]
	right := m.Images[1]

	if !left.Bounds().Eq(right.Bounds()) {
		return nil, errors.New("MPO images must be the same size to convert to anaglyph")
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
			switch colType {
			case REDBLUE:
				c = color.RGBA64{
					R: uint16(lgs),
					G: uint16(rg),
					B: uint16(rb),
					A: 65535,
				}
			case BLUERED:
				c = color.RGBA64{
					R: uint16(rgs),
					G: uint16(lg),
					B: uint16(lb),
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
