package mpo

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
)

var ErrNoImages = errors.New("no images found in mpo image")

// MPO represents the likely multiple images stored in a MPO file.
type MPO struct {
	Image []image.Image
}

const (
	mpojpgMKR = 0xFF
	mpojpgSOI = 0xD8
	mpojpgEOI = 0xD9
)

// DecodeAll reads an MPO image from r and returns the sequential frames
func DecodeAll(rr io.Reader) (*MPO, error) {
	data, err := ioutil.ReadAll(rr)
	if err != nil {
		return nil, err
	}

	r := bytes.NewReader(data)

	sectReaders := make([]*io.SectionReader, 0)
	readData := make([]byte, 1)

	var depth uint8 = 0
	var imgStart int64 = 0
	var loc int64 = 0
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
