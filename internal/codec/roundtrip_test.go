package codec

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func testImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 25), G: uint8(y * 25), B: 128, A: 255})
		}
	}
	return img
}

func TestRoundTrip_JPEG(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	var buf bytes.Buffer
	enc, err := reg.Encoder(JPEG)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}
	if err := enc.Encode(&buf, img, 90); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec, err := reg.Decoder(JPEG)
	if err != nil {
		t.Fatalf("get decoder: %v", err)
	}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestRoundTrip_PNG(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	var buf bytes.Buffer
	enc, err := reg.Encoder(PNG)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}
	if err := enc.Encode(&buf, img, 0); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec, err := reg.Decoder(PNG)
	if err != nil {
		t.Fatalf("get decoder: %v", err)
	}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestRoundTrip_GIF(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	var buf bytes.Buffer
	enc, err := reg.Encoder(GIF)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}
	if err := enc.Encode(&buf, img, 0); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec, err := reg.Decoder(GIF)
	if err != nil {
		t.Fatalf("get decoder: %v", err)
	}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestRoundTrip_BMP(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	var buf bytes.Buffer
	enc, err := reg.Encoder(BMP)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}
	if err := enc.Encode(&buf, img, 0); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec, err := reg.Decoder(BMP)
	if err != nil {
		t.Fatalf("get decoder: %v", err)
	}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestRoundTrip_TIFF(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	var buf bytes.Buffer
	enc, err := reg.Encoder(TIFF)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}
	if err := enc.Encode(&buf, img, 0); err != nil {
		t.Fatalf("encode: %v", err)
	}

	dec, err := reg.Decoder(TIFF)
	if err != nil {
		t.Fatalf("get decoder: %v", err)
	}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}

func TestJPEG_QualityAffectsSize(t *testing.T) {
	reg := DefaultRegistry()
	img := testImage(10, 10)

	enc, err := reg.Encoder(JPEG)
	if err != nil {
		t.Fatalf("get encoder: %v", err)
	}

	var bufLow bytes.Buffer
	if err := enc.Encode(&bufLow, img, 10); err != nil {
		t.Fatalf("encode low quality: %v", err)
	}

	var bufHigh bytes.Buffer
	if err := enc.Encode(&bufHigh, img, 95); err != nil {
		t.Fatalf("encode high quality: %v", err)
	}

	if bufHigh.Len() <= bufLow.Len() {
		t.Errorf("high quality size (%d) should be > low quality size (%d)",
			bufHigh.Len(), bufLow.Len())
	}
}
