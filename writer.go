package mpo

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image/jpeg"
	"io"
)

// EncodeAll encodes all images in m into a Baseline‑MP MPO and writes it to w.
func EncodeAll(w io.Writer, m *MPO, o *jpeg.Options) error {
	if o == nil {
		o = &jpeg.Options{Quality: 90}
	}

	// ── JPEG‑encode every image ────────────────────────────────────────────────
	bufs := make([][]byte, len(m.Image))
	sizes := make([]uint32, len(m.Image))
	for i, img := range m.Image {
		var b bytes.Buffer
		if err := jpeg.Encode(&b, img, o); err != nil {
			return err
		}
		bufs[i] = b.Bytes()
		sizes[i] = uint32(b.Len())
	}
	if len(bufs) == 0 {
		return errors.New("no images to encode")
	}

	first := bufs[0]
	if !bytes.HasPrefix(first, []byte{0xFF, 0xD8}) { // SOI marker
		return errors.New("first image missing SOI")
	}

	// ── build MPF segment once we know its size --------------------------------
	tmp, _ := buildMPFSegment(make([]uint32, len(sizes)), sizes)
	mpfSize := len(tmp)

	// offsets are relative to MP Endian field (see spec §5.2.3.3.3)
	jfifLen := uint32(findJFIFEnd(first[2:])) // 0 if none
	posEndian := uint32(2) + jfifLen + 8      // SOI + JFIF + 8 bytes
	offsets := make([]uint32, len(bufs))
	filePos := uint32(len(first)) + uint32(mpfSize) // size of first JPEG + MPF
	for i := 1; i < len(bufs); i++ {
		offsets[i] = filePos - posEndian
		filePos += uint32(len(bufs[i]))
	}
	// first image must be 0
	offsets[0] = 0

	mpfSeg, err := buildMPFSegment(offsets, sizes)
	if err != nil {
		return err
	}

	// ── write final MPO stream --------------------------------------------------
	if _, err := w.Write(first[:2]); err != nil { // SOI
		return err
	}
	if jfifLen > 0 {
		if _, err := w.Write(first[2 : 2+jfifLen]); err != nil {
			return err
		}
	}
	if _, err := w.Write(mpfSeg); err != nil { // APP2/MPF
		return err
	}
	startRest := 2 + int(jfifLen)
	if _, err := w.Write(first[startRest:]); err != nil { // rest of first JPEG
		return err
	}
	for i := 1; i < len(bufs); i++ { // remaining images
		if _, err := w.Write(bufs[i]); err != nil {
			return err
		}
	}
	return nil
}

const (
	tagMPFVersion  = 0xB000
	tagNumImages   = 0xB001
	tagMPImageList = 0xB002
	typeUNDEFINED  = 7
	typeLONG       = 4
	tiffHeaderSize = 8
)

const (
	flagRepresentative = 0x20000000
	mpTypeBaseline     = 0x00030000 // Baseline MP primary image
)

// buildMPFSegment constructs a valid APP2/MPF segment.
func buildMPFSegment(offsets, sizes []uint32) ([]byte, error) {
	if len(offsets) != len(sizes) {
		return nil, errors.New("offset and size counts differ")
	}

	numImg := uint32(len(offsets))
	numTags := uint16(3)

	b := new(bytes.Buffer)
	// APP2 marker & length placeholder
	b.Write([]byte{0xFF, 0xE2, 0x00, 0x00})
	// "MPF\0"
	b.Write([]byte{'M', 'P', 'F', 0x00})

	// TIFF header (little‑endian)
	b.Write([]byte("II"))
	binary.Write(b, binary.LittleEndian, uint16(0x002A))
	binary.Write(b, binary.LittleEndian, uint32(8)) // first IFD after header

	// IFD entry count
	binary.Write(b, binary.LittleEndian, numTags)

	// ── tag 0xb000 – MPFVersion ("0100") inline ――――――――――――――――――――――――――――――
	binary.Write(b, binary.LittleEndian, uint16(tagMPFVersion))
	binary.Write(b, binary.LittleEndian, uint16(typeUNDEFINED))
	binary.Write(b, binary.LittleEndian, uint32(4))
	b.Write([]byte{'0', '1', '0', '0'})

	// ── tag 0xb001 – NumberOfImages ―――――――――――――――――――――――――――――――――――――
	binary.Write(b, binary.LittleEndian, uint16(tagNumImages))
	binary.Write(b, binary.LittleEndian, uint16(typeLONG))
	binary.Write(b, binary.LittleEndian, uint32(1))
	binary.Write(b, binary.LittleEndian, numImg)

	// ── tag 0xb002 – MPImageList (offset to 16‑byte entries) ――――――――――――――――
	entryOffset := uint32(tiffHeaderSize + 2 + uint32(numTags)*12 + 4)
	binary.Write(b, binary.LittleEndian, uint16(tagMPImageList))
	binary.Write(b, binary.LittleEndian, uint16(typeUNDEFINED))
	binary.Write(b, binary.LittleEndian, uint32(numImg*16))
	binary.Write(b, binary.LittleEndian, entryOffset)

	// next‑IFD offset = 0
	binary.Write(b, binary.LittleEndian, uint32(0))

	// ── MP Entry array ―――――――――――――――――――――――――――――――――――――――――――――――――――
	for i := range offsets {
		attr := mpTypeBaseline
		if i == 0 {
			attr |= flagRepresentative
		}
		binary.Write(b, binary.LittleEndian, uint32(attr))
		binary.Write(b, binary.LittleEndian, sizes[i])
		binary.Write(b, binary.LittleEndian, offsets[i])
		binary.Write(b, binary.LittleEndian, uint16(0)) // Dep‑1
		binary.Write(b, binary.LittleEndian, uint16(0)) // Dep‑2
	}

	// fill in APP2 length (bytes after marker)
	data := b.Bytes()
	segLen := len(data) - 2
	data[2] = byte(segLen >> 8)
	data[3] = byte(segLen)

	return data, nil
}

// findJFIFEnd returns the length of an APP0/JFIF segment immediately after SOI.
func findJFIFEnd(d []byte) int {
	if len(d) < 4 || d[0] != 0xFF || d[1] != 0xE0 { // APP0?
		return 0
	}
	l := int(d[2])<<8 | int(d[3])
	if l >= 2 && len(d) >= l {
		return l
	}
	return 0
}
