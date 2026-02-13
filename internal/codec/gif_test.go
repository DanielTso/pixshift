package codec

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func TestGifDecoder_ImplementsMultiFrameDecoder(t *testing.T) {
	reg := DefaultRegistry()
	dec, err := reg.Decoder(GIF)
	if err != nil {
		t.Fatalf("get GIF decoder: %v", err)
	}
	if _, ok := dec.(MultiFrameDecoder); !ok {
		t.Error("GIF decoder should implement MultiFrameDecoder")
	}
}

func TestGifEncoder_ImplementsMultiFrameEncoder(t *testing.T) {
	reg := DefaultRegistry()
	enc, err := reg.Encoder(GIF)
	if err != nil {
		t.Fatalf("get GIF encoder: %v", err)
	}
	if _, ok := enc.(MultiFrameEncoder); !ok {
		t.Error("GIF encoder should implement MultiFrameEncoder")
	}
}

// buildAnimatedGIF creates a simple 3-frame animated GIF in a buffer.
func buildAnimatedGIF(t *testing.T) *bytes.Buffer {
	t.Helper()
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	}

	g := &gif.GIF{
		LoopCount: 0,
	}
	for i := 0; i < 3; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, 10, 10), palette)
		// Fill each frame with a different color
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				frame.SetColorIndex(x, y, uint8(i+1))
			}
		}
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, 10)
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("build animated GIF: %v", err)
	}
	return &buf
}

func TestDecodeAll_AnimatedGIF(t *testing.T) {
	buf := buildAnimatedGIF(t)

	dec := &gifDecoder{}
	anim, err := dec.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}

	if len(anim.Frames) != 3 {
		t.Fatalf("expected 3 frames, got %d", len(anim.Frames))
	}

	if len(anim.Delays) != 3 {
		t.Errorf("expected 3 delays, got %d", len(anim.Delays))
	}

	// Each frame should be 10x10
	for i, frame := range anim.Frames {
		bounds := frame.Bounds()
		if bounds.Dx() != 10 || bounds.Dy() != 10 {
			t.Errorf("frame %d: size = %dx%d, want 10x10", i, bounds.Dx(), bounds.Dy())
		}
	}

	// Frames should be converted to NRGBA
	for i, frame := range anim.Frames {
		if _, ok := frame.(*image.NRGBA); !ok {
			t.Errorf("frame %d: expected *image.NRGBA, got %T", i, frame)
		}
	}
}

func TestEncodeAll_RoundTrip(t *testing.T) {
	// Create an AnimatedImage with 3 frames
	anim := &AnimatedImage{
		Frames:    make([]image.Image, 3),
		Delays:    []int{10, 20, 30},
		LoopCount: 0,
	}

	colors := []color.NRGBA{
		{255, 0, 0, 255},
		{0, 255, 0, 255},
		{0, 0, 255, 255},
	}

	for i := 0; i < 3; i++ {
		nrgba := image.NewNRGBA(image.Rect(0, 0, 8, 8))
		for y := 0; y < 8; y++ {
			for x := 0; x < 8; x++ {
				nrgba.Set(x, y, colors[i])
			}
		}
		anim.Frames[i] = nrgba
	}

	// Encode
	enc := &gifEncoder{}
	var buf bytes.Buffer
	if err := enc.EncodeAll(&buf, anim); err != nil {
		t.Fatalf("EncodeAll: %v", err)
	}

	// Decode back
	dec := &gifDecoder{}
	anim2, err := dec.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll (round-trip): %v", err)
	}

	if len(anim2.Frames) != 3 {
		t.Fatalf("round-trip: expected 3 frames, got %d", len(anim2.Frames))
	}

	if len(anim2.Delays) != 3 {
		t.Errorf("round-trip: expected 3 delays, got %d", len(anim2.Delays))
	}

	// Verify delays are preserved
	for i, delay := range anim2.Delays {
		if delay != anim.Delays[i] {
			t.Errorf("round-trip delay[%d] = %d, want %d", i, delay, anim.Delays[i])
		}
	}

	// Verify frame dimensions
	for i, frame := range anim2.Frames {
		bounds := frame.Bounds()
		if bounds.Dx() != 8 || bounds.Dy() != 8 {
			t.Errorf("round-trip frame %d: size = %dx%d, want 8x8", i, bounds.Dx(), bounds.Dy())
		}
	}
}

func TestDecodeAll_SingleFrame(t *testing.T) {
	// Build a single-frame GIF
	palette := color.Palette{
		color.RGBA{0, 0, 0, 255},
		color.RGBA{255, 255, 255, 255},
	}
	g := &gif.GIF{
		Image: []*image.Paletted{image.NewPaletted(image.Rect(0, 0, 5, 5), palette)},
		Delay: []int{0},
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("encode single-frame GIF: %v", err)
	}

	dec := &gifDecoder{}
	anim, err := dec.DecodeAll(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("DecodeAll: %v", err)
	}

	if len(anim.Frames) != 1 {
		t.Errorf("expected 1 frame, got %d", len(anim.Frames))
	}
}

func TestSingleFrameEncode_BackwardCompatible(t *testing.T) {
	// Verify the single-frame Encode method still works
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{128, 64, 32, 255})
		}
	}

	enc := &gifEncoder{}
	var buf bytes.Buffer
	if err := enc.Encode(&buf, img, 0); err != nil {
		t.Fatalf("Encode: %v", err)
	}

	// Verify it's a valid GIF
	dec := &gifDecoder{}
	decoded, err := dec.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 10 || bounds.Dy() != 10 {
		t.Errorf("decoded size = %dx%d, want 10x10", bounds.Dx(), bounds.Dy())
	}
}
