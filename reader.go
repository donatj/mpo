// Copyright 2015-2025 Jesse G. Donat.
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
// # Nintendo 3DS Support
//
// DecodeAll optionally parses Nintendo 3DS-specific metadata from APP2/NINT
// segments when present. This metadata is stored in the MPO.Nintendo field
// and can be accessed using the HasNintendoMetadata() method. The raw bytes
// are preserved for custom parsing of Nintendo-specific stereoscopic parameters.
//
// Specification references:
//
//   - CIPA DC‑X007:2012 – Multi‑Picture Format (MPF)
//     https://www.cipa.jp/std/documents/e/DC-007-2012_E.pdf
//   - ISO/IEC 10918‑1 – JPEG Baseline coding and marker layout.
//   - JFIF 1.02 – APP0/JFIF segment details.
//   - Nintendo 3DS MPO Format – https://3dbrew.org/wiki/MPO
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

// NintendoMetadata contains Nintendo 3DS-specific metadata from APP2/NINT segments.
// This is optional metadata that may be present in MPO files created by Nintendo 3DS cameras.
//
// Nintendo 3DS cameras embed proprietary stereoscopic metadata in APP2 segments with
// the "NINT" identifier. This metadata can include parallax settings, camera calibration
// data, and 3D effect parameters.
//
// Reference: https://3dbrew.org/wiki/MPO
type NintendoMetadata struct {
	// Raw contains the raw bytes of the NINT segment data (after the "NINT" identifier)
	Raw []byte
}

// HasNintendoMetadata returns true if the MPO contains Nintendo 3DS-specific metadata.
func (m *MPO) HasNintendoMetadata() bool {
	return m.Nintendo != nil && len(m.Nintendo.Raw) > 0
}

// MPO represents the likely multiple images stored in a MPO file.
type MPO struct {
	Image []image.Image
	// Nintendo contains optional Nintendo 3DS-specific metadata, if present in the file.
	Nintendo *NintendoMetadata
}

const (
	mpojpgMKR = 0xFF
	mpojpgSOI = 0xD8 // Start of Image
	mpojpgEOI = 0xD9 // End of Image
	mpojpgAPP2 = 0xE2 // APP2 marker
)

// DecodeAll reads an MPO image from r and returns the sequential frames
func DecodeAll(rr io.Reader) (*MPO, error) {
	var rAt io.ReaderAt
	var rawData []byte
	if ra, ok := rr.(io.ReaderAt); ok {
		rAt = ra
		// Try to read the data to parse Nintendo metadata
		// For ReaderAt, we need to read the full content
		if seeker, ok := rr.(io.Seeker); ok {
			// Save current position
			if pos, err := seeker.Seek(0, io.SeekCurrent); err == nil {
				// Read all data
				if buf, err := io.ReadAll(rr); err == nil {
					rawData = buf
					// Restore position
					seeker.Seek(pos, io.SeekStart)
				}
			}
		}
	} else {
		// fallback: buffer entire data (for readers that lack ReaderAt)
		buf, err := io.ReadAll(rr)
		if err != nil {
			return nil, err
		}
		rawData = buf
		rAt = bytes.NewReader(buf)
	}

	r := io.NewSectionReader(rAt, 0, 1<<63-1)

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

	// Parse Nintendo metadata if we have raw data
	if len(rawData) > 0 {
		m.Nintendo = parseNintendoMetadata(rawData)
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

// parseNintendoMetadata scans the raw data for APP2/NINT segments and extracts Nintendo metadata.
// Returns nil if no Nintendo metadata is found.
func parseNintendoMetadata(data []byte) *NintendoMetadata {
	// Scan for APP2 markers with NINT identifier
	pos := 0
	for pos < len(data)-8 {
		// Look for FF E2 (APP2 marker)
		if data[pos] == mpojpgMKR && pos+1 < len(data) && data[pos+1] == mpojpgAPP2 {
			// Read segment length (big-endian)
			if pos+3 >= len(data) {
				break
			}
			segLen := int(data[pos+2])<<8 | int(data[pos+3])
			if segLen < 2 || pos+2+segLen > len(data) {
				pos++
				continue
			}
			
			// Check if this is a NINT segment
			if pos+8 <= len(data) && 
			   data[pos+4] == 'N' && data[pos+5] == 'I' && 
			   data[pos+6] == 'N' && data[pos+7] == 'T' {
				// Found Nintendo metadata
				// Extract the data after "NINT" identifier (4 bytes) up to segment length
				dataStart := pos + 8
				dataEnd := pos + 2 + segLen
				if dataEnd > len(data) {
					dataEnd = len(data)
				}
				
				raw := make([]byte, dataEnd-dataStart)
				copy(raw, data[dataStart:dataEnd])
				
				return &NintendoMetadata{
					Raw: raw,
				}
			}
			
			pos += 2 + segLen
		} else {
			pos++
		}
	}
	
	return nil
}

func init() {
	// \xff\xd8\xff clashes with jpeg but hopefully shouldn't cause issues
	image.RegisterFormat("mpo", "\xff\xd8\xff", Decode, DecodeConfig)
}
