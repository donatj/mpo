package mpo

import (
	"image"
	"image/jpeg"
	"io"
)

// MPO represents the likely multiple images stored in a MPO file.
type MPO struct {
	Image []image.Image
}

const (
	mpojpgMKR = 0xFF
	mpojpgSOI = 0xD8
	mpojpgEOI = 0xD9
)

type MpoReader interface {
	io.Reader
	io.ReaderAt
}

// DecodeAll reads an MPO image from r and returns the sequential frames
func DecodeAll(r MpoReader) (*MPO, error) {
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
