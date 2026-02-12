package codec

import (
	"bytes"
	"testing"
)

func padTo(data []byte, size int) []byte {
	if len(data) >= size {
		return data
	}
	padded := make([]byte, size)
	copy(padded, data)
	return padded
}

func TestDetectFormat_JPEG(t *testing.T) {
	data := padTo([]byte{0xFF, 0xD8, 0xFF, 0xE0}, 32)
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != JPEG {
		t.Errorf("DetectFormat = %q, want %q", f, JPEG)
	}
}

func TestDetectFormat_PNG(t *testing.T) {
	data := padTo([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, 32)
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "photo.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != PNG {
		t.Errorf("DetectFormat = %q, want %q", f, PNG)
	}
}

func TestDetectFormat_GIF(t *testing.T) {
	data := padTo([]byte("GIF89a"), 32)
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "animation.gif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != GIF {
		t.Errorf("DetectFormat = %q, want %q", f, GIF)
	}
}

func TestDetectFormat_BMP(t *testing.T) {
	data := padTo([]byte("BM"), 32)
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "image.bmp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != BMP {
		t.Errorf("DetectFormat = %q, want %q", f, BMP)
	}
}

func TestDetectFormat_WebP(t *testing.T) {
	data := make([]byte, 32)
	copy(data[0:4], "RIFF")
	copy(data[8:12], "WEBP")
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "image.webp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != WebP {
		t.Errorf("DetectFormat = %q, want %q", f, WebP)
	}
}

func TestDetectFormat_TIFF(t *testing.T) {
	data := padTo([]byte("II\x2a\x00"), 32) // little-endian TIFF
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "image.tiff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != TIFF {
		t.Errorf("DetectFormat = %q, want %q", f, TIFF)
	}
}

func TestDetectFormat_ExtensionFallback(t *testing.T) {
	// Unknown magic bytes, fall back to extension
	data := padTo([]byte{0x00, 0x00, 0x00, 0x00}, 32)
	r := bytes.NewReader(data)
	f, err := DetectFormat(r, "photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f != JPEG {
		t.Errorf("DetectFormat = %q, want %q", f, JPEG)
	}
}

func TestDetectFormat_UnknownFormat(t *testing.T) {
	data := padTo([]byte{0x00, 0x00, 0x00, 0x00}, 32)
	r := bytes.NewReader(data)
	_, err := DetectFormat(r, "file.xyz")
	if err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}

func TestDetectFormat_SeekReset(t *testing.T) {
	data := padTo([]byte{0xFF, 0xD8, 0xFF, 0xE0}, 32)
	r := bytes.NewReader(data)
	_, err := DetectFormat(r, "photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After DetectFormat, reader should be reset to start
	pos, _ := r.Seek(0, 1) // current position
	if pos != 0 {
		t.Errorf("reader position after DetectFormat = %d, want 0", pos)
	}
}
