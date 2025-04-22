// Copyright 2015 Jesse G. Donat.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE.md file.

// Package mpo provides simple read and write support for
// Multi‑Picture Object (MPO) files, a format that stores two or more JPEG
// frames in a single byte stream, and is used in some 3D cameras.
//
// The package is designed to be simple and easy to use, with a focus on
// extracting and writing the JPEG frames. It does not attempt to
// implement the full specification of the MPO format, but rather provides
// a basic set of functions to work with the most common use cases.
//
// The package offers:
//
//   - DecodeAll  – extract every JPEG frame present in an MPO.
//   - EncodeAll  – write a Baseline‑MP MPO from a slice of image.Image.
//   - ConvertToStereo   – merge the first two frames side‑by‑side.
//   - ConvertToAnaglyph – create red/cyan or similar anaglyphs.
//
// EncodeAll produces only the subset required for a Baseline‑MP file: the
// first frame is flagged as the representative image and is given MP type
// 0x00030000. DecodeAll imposes no such restriction and simply returns every
// JPEG it finds.
//
// Specification references:
//
//   - CIPA DC‑X007:2012 – Multi‑Picture Format (MPF)
//     https://www.cipa.jp/std/documents/e/DC-007-2012_E.pdf
//   - ISO/IEC 10918‑1 – JPEG Baseline coding and marker layout.
//   - JFIF 1.02 – APP0/JFIF segment details.
//
// Offsets in the MP Image List are measured relative to the TIFF endian
// marker inside the APP2/MPF segment, as required by DC‑X007 §5.2.3.3.
// The code relies only on the Go standard library and is safe for pure‑Go
// builds.
package mpo

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"
)

// ErrNoImages indicates that no images were found in the specified file.
var ErrNoImages = errors.New("no images found in mpo image")

// MPO represents the likely multiple images stored in a MPO file.
type MPO struct {
	Image []image.Image
}

const (
	mpojpgMKR = 0xFF
	mpojpgSOI = 0xD8 // Start of Image
	mpojpgEOI = 0xD9 // End of Image
)

// DecodeAll reads an MPO image from r and returns the sequential frames
func DecodeAll(rr io.Reader) (*MPO, error) {
	data, err := io.ReadAll(rr)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)

	sectReaders := make([]*io.SectionReader, 0)
	readData := make([]byte, 1)

	var (
		depth    uint8
		imgStart int64
		loc      int64
	)

	for {
		i1, err := r.Read(readData)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		loc += int64(i1)

		if readData[0] == mpojpgMKR {
			i2, err := r.Read(readData)
			if err == io.EOF {
				break
			} else if err != nil {
				return nil, err
			}
			loc += int64(i2)

			if readData[0] == mpojpgSOI {
				if depth == 0 {
					imgStart = loc - 2
				}

				depth++
			} else if readData[0] == mpojpgEOI {
				depth--
				if depth == 0 {
					sectReaders = append(sectReaders, io.NewSectionReader(r, imgStart, loc))
				}

			}
		}
	}

	m := &MPO{
		Image: make([]image.Image, 0),
	}

	for _, s := range sectReaders {
		img, err := jpeg.Decode(s)
		if err != nil {
			return nil, err
		}

		m.Image = append(m.Image, img)
	}

	return m, nil
}

// Decode reads a MPO image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	all, err := DecodeAll(r)
	if err != nil {
		return nil, err
	}

	if len(all.Image) < 1 {
		return nil, ErrNoImages
	}

	return all.Image[0], nil
}

// DecodeConfig returns the color model and dimensions of an MPO image without
// decoding the entire image.
//
// TODO Optimize this - possibly just falling back to jpeg.DecodeConfig
func DecodeConfig(r io.Reader) (image.Config, error) {
	all, err := DecodeAll(r)
	if err != nil {
		return image.Config{}, err
	}

	if len(all.Image) < 1 {
		return image.Config{}, ErrNoImages
	}

	return image.Config{
		ColorModel: all.Image[0].ColorModel(),
		Width:      all.Image[0].Bounds().Max.X,
		Height:     all.Image[0].Bounds().Max.Y,
	}, nil
}

func init() {
	// \xff\xd8\xff clashes with jpeg but hopefully shouldn't cause issues
	image.RegisterFormat("mpo", "\xff\xd8\xff", Decode, DecodeConfig)
}
