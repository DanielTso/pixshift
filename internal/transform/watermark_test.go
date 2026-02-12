package transform

import (
	"image"
	"image/color"
	"testing"
)

func solidImage(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestApplyWatermark_EmptyText_ReturnsOriginal(t *testing.T) {
	img := solidImage(100, 80, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: ""})
	if out != img {
		t.Error("empty text should return the original image pointer")
	}
}

func TestApplyWatermark_ValidText_ProducesDifferentImage(t *testing.T) {
	img := solidImage(200, 200, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "Test", Opacity: 0.8})

	// The output should be different from the input (watermark drawn)
	b := out.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("watermarked image dimensions: got %dx%d, want 200x200", b.Dx(), b.Dy())
	}

	// Check that at least some pixels differ from the solid white input
	differs := false
	for y := b.Min.Y; y < b.Max.Y && !differs; y++ {
		for x := b.Min.X; x < b.Max.X && !differs; x++ {
			r, g, bb, _ := out.At(x, y).RGBA()
			if r != 0xFFFF || g != 0xFFFF || bb != 0xFFFF {
				differs = true
			}
		}
	}
	if !differs {
		t.Error("watermarked image should have at least some non-white pixels")
	}
}

func TestApplyWatermark_OpacityZero_DefaultsToHalf(t *testing.T) {
	// Opacity 0 should default to 0.5; the watermark should still be visible.
	img := solidImage(200, 200, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "Test", Opacity: 0})
	b := out.Bounds()
	differs := false
	for y := b.Min.Y; y < b.Max.Y && !differs; y++ {
		for x := b.Min.X; x < b.Max.X && !differs; x++ {
			r, g, bb, _ := out.At(x, y).RGBA()
			if r != 0xFFFF || g != 0xFFFF || bb != 0xFFFF {
				differs = true
			}
		}
	}
	if !differs {
		t.Error("opacity 0 (defaulting to 0.5) should still produce visible watermark")
	}
}

func TestApplyWatermark_OpacityNegative_DefaultsToHalf(t *testing.T) {
	img := solidImage(200, 200, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "Hello", Opacity: -0.5})
	b := out.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("negative opacity: got %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_OpacityAboveOne_Clamped(t *testing.T) {
	img := solidImage(200, 200, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "Hello", Opacity: 2.5})
	b := out.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("clamped opacity: got %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_PositionBottomRight(t *testing.T) {
	img := solidImage(300, 300, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "WM", Position: "bottom-right", Opacity: 1.0})
	b := out.Bounds()
	if b.Dx() != 300 || b.Dy() != 300 {
		t.Errorf("bottom-right: got %dx%d, want 300x300", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_PositionBottomLeft(t *testing.T) {
	img := solidImage(300, 300, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "WM", Position: "bottom-left", Opacity: 1.0})
	b := out.Bounds()
	if b.Dx() != 300 || b.Dy() != 300 {
		t.Errorf("bottom-left: got %dx%d, want 300x300", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_PositionTopRight(t *testing.T) {
	img := solidImage(300, 300, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "WM", Position: "top-right", Opacity: 1.0})
	b := out.Bounds()
	if b.Dx() != 300 || b.Dy() != 300 {
		t.Errorf("top-right: got %dx%d, want 300x300", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_PositionTopLeft(t *testing.T) {
	img := solidImage(300, 300, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "WM", Position: "top-left", Opacity: 1.0})
	b := out.Bounds()
	if b.Dx() != 300 || b.Dy() != 300 {
		t.Errorf("top-left: got %dx%d, want 300x300", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_PositionCenter(t *testing.T) {
	img := solidImage(300, 300, color.White)
	out := ApplyWatermark(img, WatermarkOptions{Text: "WM", Position: "center", Opacity: 1.0})
	b := out.Bounds()
	if b.Dx() != 300 || b.Dy() != 300 {
		t.Errorf("center: got %dx%d, want 300x300", b.Dx(), b.Dy())
	}
}

func TestApplyWatermark_DifferentPositions_ProduceDifferentResults(t *testing.T) {
	positions := []string{"top-left", "top-right", "bottom-left", "bottom-right", "center"}
	results := make([]image.Image, len(positions))
	for i, pos := range positions {
		img := solidImage(300, 300, color.White)
		results[i] = ApplyWatermark(img, WatermarkOptions{
			Text:     "Watermark",
			Position: pos,
			Opacity:  1.0,
		})
	}

	// At least some pairs should differ in pixel content
	allSame := true
	for i := 1; i < len(results); i++ {
		if !imagesEqual(results[0], results[i]) {
			allSame = false
			break
		}
	}
	if allSame {
		t.Error("different positions should produce different pixel content")
	}
}

func TestApplyWatermark_FontScale(t *testing.T) {
	// Scale 2 should produce a larger watermark area than scale 1.
	img1 := solidImage(400, 400, color.White)
	out1 := ApplyWatermark(img1, WatermarkOptions{Text: "Scale", Opacity: 1.0, FontSize: 1.0})
	img2 := solidImage(400, 400, color.White)
	out2 := ApplyWatermark(img2, WatermarkOptions{Text: "Scale", Opacity: 1.0, FontSize: 3.0})

	// Count non-white pixels in each.
	count1 := countNonWhitePixels(out1)
	count2 := countNonWhitePixels(out2)
	if count2 <= count1 {
		t.Errorf("FontSize 3.0 should produce more non-white pixels (%d) than 1.0 (%d)", count2, count1)
	}
}

func TestApplyWatermark_CustomColor(t *testing.T) {
	img := solidImage(200, 200, color.White)
	out := ApplyWatermark(img, WatermarkOptions{
		Text:    "Red",
		Opacity: 1.0,
		Color:   "#FF0000",
		BgColor: "#00FF00",
	})
	b := out.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("custom color: got %dx%d, want 200x200", b.Dx(), b.Dy())
	}
	// Check that at least some pixels are not white.
	differs := false
	for y := b.Min.Y; y < b.Max.Y && !differs; y++ {
		for x := b.Min.X; x < b.Max.X && !differs; x++ {
			r, g, bl, _ := out.At(x, y).RGBA()
			if r != 0xFFFF || g != 0xFFFF || bl != 0xFFFF {
				differs = true
			}
		}
	}
	if !differs {
		t.Error("custom color watermark should produce visible pixels")
	}
}

func TestParseHexColor(t *testing.T) {
	fallback := color.RGBA{R: 0, G: 0, B: 0, A: 255}

	tests := []struct {
		input string
		want  color.RGBA
	}{
		{"#FF0000", color.RGBA{R: 255, G: 0, B: 0, A: 255}},
		{"00FF00", color.RGBA{R: 0, G: 255, B: 0, A: 255}},
		{"#0000ff", color.RGBA{R: 0, G: 0, B: 255, A: 255}},
		{"", fallback},
		{"invalid", fallback},
		{"#GG0000", fallback},
	}

	for _, tc := range tests {
		got := ParseHexColor(tc.input, fallback)
		if got != tc.want {
			t.Errorf("ParseHexColor(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func countNonWhitePixels(img image.Image) int {
	count := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := img.At(x, y).RGBA()
			if r != 0xFFFF || g != 0xFFFF || bl != 0xFFFF {
				count++
			}
		}
	}
	return count
}

func imagesEqual(a, b image.Image) bool {
	ab := a.Bounds()
	bb := b.Bounds()
	if ab.Dx() != bb.Dx() || ab.Dy() != bb.Dy() {
		return false
	}
	for y := 0; y < ab.Dy(); y++ {
		for x := 0; x < ab.Dx(); x++ {
			r1, g1, b1, a1 := a.At(x+ab.Min.X, y+ab.Min.Y).RGBA()
			r2, g2, b2, a2 := b.At(x+bb.Min.X, y+bb.Min.Y).RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				return false
			}
		}
	}
	return true
}
