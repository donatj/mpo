package mpo_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/donatj/mpo"
)

func TestConvertToStereo(t *testing.T) {
	// Two 1×1 images: left=red, right=blue
	img1 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img1.Set(0, 0, color.RGBA{255, 0, 0, 255})
	img2 := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img2.Set(0, 0, color.RGBA{0, 0, 255, 255})

	m := &mpo.MPO{Image: []image.Image{img1, img2}}
	stereo := m.ConvertToStereo()
	b := stereo.Bounds()
	if dx, dy := b.Dx(), b.Dy(); dx != 2 || dy != 1 {
		t.Fatalf("stereo bounds = %dx%d, want 2x1", dx, dy)
	}

	if c := stereo.At(0, 0); c != img1.At(0, 0) {
		t.Errorf("pixel 0,0 = %v, want %v", c, img1.At(0, 0))
	}
	if c := stereo.At(1, 0); c != img2.At(0, 0) {
		t.Errorf("pixel 1,0 = %v, want %v", c, img2.At(0, 0))
	}
}

func TestConvertToAnaglyph_UnsupportedCount(t *testing.T) {
	// Only one frame => error
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	m := &mpo.MPO{Image: []image.Image{img}}
	_, err := m.ConvertToAnaglyph(mpo.RedCyan)
	if err == nil {
		t.Fatal("expected error for single-image anaglyph, got nil")
	}
}

func TestConvertToAnaglyph_LuminanceCoefficients(t *testing.T) {
	tests := []struct {
		name  string
		left  color.RGBA
		coeff float32
	}{
		{
			name:  "red channel weight",
			left:  color.RGBA{255, 0, 0, 255},
			coeff: .299,
		},
		{
			name:  "blue channel weight",
			left:  color.RGBA{0, 0, 255, 255},
			coeff: .114,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			left := image.NewRGBA(image.Rect(0, 0, 1, 1))
			left.Set(0, 0, tc.left)

			right := image.NewRGBA(image.Rect(0, 0, 1, 1))
			right.Set(0, 0, color.RGBA{0, 0, 0, 255})

			m := &mpo.MPO{Image: []image.Image{left, right}}
			anaglyph, err := m.ConvertToAnaglyph(mpo.RedCyan)
			if err != nil {
				t.Fatalf("ConvertToAnaglyph failed: %v", err)
			}

			// Convert the 16-bit luminance value through RGBA's 8-bit storage path.
			rawLuminance := uint16(float32(65535) * tc.coeff)
			expectedR8 := uint16(uint8(rawLuminance >> 8))
			expectedR := expectedR8<<8 | expectedR8

			got := color.RGBA64Model.Convert(anaglyph.At(0, 0)).(color.RGBA64)
			if got.R != expectedR {
				t.Fatalf("red channel = %d, want %d", got.R, expectedR)
			}
		})
	}
}
