package mpo

import (
	"image"
	"image/jpeg"
	"io"
	"os"
)

type Mpo struct {
	Images []*image.Image
}

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

const (
	mpojpgMKR = 0xFF
	mpojpgSOI = 0xD8
	mpojpgEOI = 0xD9
)

func Decode(filename string) (*Mpo, error) {
	r, err := os.Open(filename)

	if err != nil {
		return nil, err
	}

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

	m := &Mpo{
		Images: make([]*image.Image, 0),
	}

	for _, s := range sectReaders {
		img, _, err := image.Decode(s)
		if err != nil {
			return nil, err
		}

		m.Images = append(m.Images, &img)
	}

	return m, nil

}
