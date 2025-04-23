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
