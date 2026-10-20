package mpo_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/donatj/mpo"
)

// TestNintendoMetadata_NotPresent verifies that MPO files without Nintendo metadata
// work correctly (backward compatibility)
func TestNintendoMetadata_NotPresent(t *testing.T) {
	// Create a simple MPO without Nintendo metadata
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	m := &mpo.MPO{Image: []image.Image{img}}
	var buf bytes.Buffer
	if err := mpo.EncodeAll(&buf, m, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("EncodeAll failed: %v", err)
	}

	// Decode and verify no Nintendo metadata
	decoded, err := mpo.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll failed: %v", err)
	}

	if decoded.Nintendo != nil {
		t.Error("expected Nintendo metadata to be nil for standard MPO")
	}
}

// TestNintendoMetadata_Present verifies that Nintendo metadata is correctly parsed
func TestNintendoMetadata_Present(t *testing.T) {
	// Create a mock MPO with Nintendo APP2/NINT segment
	// Build a minimal JPEG with Nintendo metadata
	var buf bytes.Buffer

	// SOI marker
	buf.Write([]byte{0xFF, 0xD8})

	// APP2/NINT segment
	buf.Write([]byte{0xFF, 0xE2}) // APP2 marker
	nintData := []byte("Test Nintendo Data")
	segLen := 2 + 4 + len(nintData)                    // length field + "NINT" + data
	buf.Write([]byte{byte(segLen >> 8), byte(segLen)}) // segment length
	buf.Write([]byte{'N', 'I', 'N', 'T'})              // NINT identifier
	buf.Write(nintData)                                // Nintendo-specific data

	// Add a minimal valid JPEG (1x1 red pixel)
	// This is a simplified JPEG structure - enough to pass basic parsing
	// For a real test, we'd need a complete valid JPEG
	buf.Write([]byte{
		0xFF, 0xDB, 0x00, 0x43, 0x00, // DQT marker and basic quantization table
	})
	// Add dummy quantization table
	for i := 0; i < 64; i++ {
		buf.WriteByte(0x10)
	}

	// SOF0 marker - start of frame
	buf.Write([]byte{
		0xFF, 0xC0, 0x00, 0x0B, // SOF0 marker, length
		0x08,       // bits per component
		0x00, 0x01, // height = 1
		0x00, 0x01, // width = 1
		0x01,             // number of components
		0x01, 0x11, 0x00, // component 1 spec
	})

	// DHT marker - Huffman table (minimal)
	buf.Write([]byte{
		0xFF, 0xC4, 0x00, 0x14, 0x00, // DHT marker, length, table info
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	})

	// SOS marker - start of scan
	buf.Write([]byte{
		0xFF, 0xDA, 0x00, 0x08, // SOS marker, length
		0x01,       // components in scan
		0x01, 0x00, // component selector and tables
		0x00, 0x3F, 0x00, // spectral selection
	})

	// Minimal scan data
	buf.Write([]byte{0x00, 0x01})

	// EOI marker
	buf.Write([]byte{0xFF, 0xD9})

	// Try to decode - it might fail on JPEG decode but should parse Nintendo metadata
	decoded, err := mpo.DecodeAll(bytes.NewReader(buf.Bytes()))
	// We expect this might fail during JPEG decode, but we should still get Nintendo metadata
	if err != nil {
		// If JPEG decoding fails, that's okay for this test - we're testing metadata parsing
		// Let's verify the error is JPEG-related and try a different approach
		t.Logf("JPEG decode failed (expected for minimal JPEG): %v", err)

		// Test with a real encoded JPEG instead
		testWithRealJPEG(t)
		return
	}

	// If we got here, check Nintendo metadata
	if decoded.Nintendo == nil {
		t.Fatal("expected Nintendo metadata to be present")
	}

	if !bytes.Equal(decoded.Nintendo.Raw, nintData) {
		t.Errorf("Nintendo metadata mismatch: got %q, want %q", decoded.Nintendo.Raw, nintData)
	}
}

// testWithRealJPEG creates a test with a real JPEG that has Nintendo metadata injected
func testWithRealJPEG(t *testing.T) {
	// Create a real JPEG first
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}

	// Now inject Nintendo metadata after SOI
	jpegBytes := jpegBuf.Bytes()
	if len(jpegBytes) < 2 || jpegBytes[0] != 0xFF || jpegBytes[1] != 0xD8 {
		t.Fatal("invalid JPEG SOI")
	}

	// Build new buffer with Nintendo metadata
	var withNintendo bytes.Buffer
	withNintendo.Write(jpegBytes[0:2]) // SOI

	// Add Nintendo APP2/NINT segment
	nintData := []byte("Test Nintendo Data")
	withNintendo.Write([]byte{0xFF, 0xE2}) // APP2 marker
	segLen := 2 + 4 + len(nintData)
	withNintendo.Write([]byte{byte(segLen >> 8), byte(segLen)})
	withNintendo.Write([]byte{'N', 'I', 'N', 'T'})
	withNintendo.Write(nintData)

	// Add rest of JPEG
	withNintendo.Write(jpegBytes[2:])

	// Decode
	decoded, err := mpo.DecodeAll(bytes.NewReader(withNintendo.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll failed: %v", err)
	}

	if decoded.Nintendo == nil {
		t.Fatal("expected Nintendo metadata to be present")
	}

	if !bytes.Equal(decoded.Nintendo.Raw, nintData) {
		t.Errorf("Nintendo metadata mismatch: got %q, want %q", decoded.Nintendo.Raw, nintData)
	}

	// Verify we still got the image
	if len(decoded.Image) != 1 {
		t.Errorf("expected 1 image, got %d", len(decoded.Image))
	}
}

// TestHasNintendoMetadata tests the HasNintendoMetadata helper method
func TestHasNintendoMetadata(t *testing.T) {
	tests := []struct {
		name     string
		mpo      *mpo.MPO
		expected bool
	}{
		{
			name:     "nil Nintendo field",
			mpo:      &mpo.MPO{Nintendo: nil},
			expected: false,
		},
		{
			name:     "empty Raw data",
			mpo:      &mpo.MPO{Nintendo: &mpo.NintendoMetadata{Raw: []byte{}}},
			expected: false,
		},
		{
			name:     "nil Raw data",
			mpo:      &mpo.MPO{Nintendo: &mpo.NintendoMetadata{Raw: nil}},
			expected: false,
		},
		{
			name:     "valid Nintendo data",
			mpo:      &mpo.MPO{Nintendo: &mpo.NintendoMetadata{Raw: []byte("test")}},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mpo.HasNintendoMetadata()
			if got != tt.expected {
				t.Errorf("HasNintendoMetadata() = %v, want %v", got, tt.expected)
			}
		})
	}
}
