// writer_test.go
package mpo_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/donatj/mpo"
)

func TestEncodeAll_NoImages(t *testing.T) {
	var buf bytes.Buffer
	err := mpo.EncodeAll(&buf, &mpo.MPO{Image: nil}, nil)
	if err == nil {
		t.Fatal("expected error when encoding zero images, got nil")
	}
}

func TestEncodeAll_RoundTrip(t *testing.T) {
	// Create two distinct 10×10 images
	img1 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img2 := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img1.Set(x, y, color.RGBA{255, 0, 0, 255})
			img2.Set(x, y, color.RGBA{0, 255, 0, 255})
		}
	}

	m := &mpo.MPO{Image: []image.Image{img1, img2}}
	var buf bytes.Buffer
	if err := mpo.EncodeAll(&buf, m, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("EncodeAll failed: %v", err)
	}

	// DecodeAll should return two frames
	decoded, err := mpo.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll failed: %v", err)
	}
	if got := len(decoded.Image); got != 2 {
		t.Fatalf("expected 2 images, got %d", got)
	}

	// DecodeConfig should report correct dimensions
	cfg, err := mpo.DecodeConfig(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeConfig failed: %v", err)
	}
	if cfg.Width != 10 || cfg.Height != 10 {
		t.Fatalf("unexpected dimensions: got %dx%d, want 10x10", cfg.Width, cfg.Height)
	}
}
