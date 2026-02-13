package color

import (
	"image"
	"image/color"
	"testing"
)

// newSolidImage creates a uniform single-color image.
func newSolidImage(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

// newTwoColorImage creates an image split vertically: left half is c1, right half is c2.
func newTwoColorImage(w, h int, c1, c2 color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	mid := w / 2
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x < mid {
				img.Set(x, y, c1)
			} else {
				img.Set(x, y, c2)
			}
		}
	}
	return img
}

func TestExtractPalette_TwoColors(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	img := newTwoColorImage(100, 100, red, blue)

	colors := ExtractPalette(img, 2)
	if len(colors) != 2 {
		t.Fatalf("expected 2 colors, got %d", len(colors))
	}

	foundRed := false
	foundBlue := false
	for _, c := range colors {
		if c.R > 200 && c.G < 50 && c.B < 50 {
			foundRed = true
			if c.Percentage < 40 || c.Percentage > 60 {
				t.Errorf("red percentage %.1f%% not within 40-60%% range", c.Percentage)
			}
		}
		if c.B > 200 && c.R < 50 && c.G < 50 {
			foundBlue = true
			if c.Percentage < 40 || c.Percentage > 60 {
				t.Errorf("blue percentage %.1f%% not within 40-60%% range", c.Percentage)
			}
		}
	}
	if !foundRed {
		t.Error("red color not found in palette")
	}
	if !foundBlue {
		t.Error("blue color not found in palette")
	}
}

func TestExtractPalette_HexFormat(t *testing.T) {
	img := newSolidImage(50, 50, color.RGBA{R: 255, G: 128, B: 0, A: 255})
	colors := ExtractPalette(img, 1)
	if len(colors) == 0 {
		t.Fatal("expected at least 1 color")
	}
	hex := colors[0].Hex
	if len(hex) != 7 || hex[0] != '#' {
		t.Errorf("invalid hex format: %q", hex)
	}
}

func TestExtractPalette_SolidColor(t *testing.T) {
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	img := newSolidImage(100, 100, green)

	colors := ExtractPalette(img, 3)
	if len(colors) == 0 {
		t.Fatal("expected at least 1 color")
	}

	// The dominant color should be green at ~100%.
	c := colors[0]
	if c.G < 240 {
		t.Errorf("expected green channel near 255, got %d", c.G)
	}
	if c.Percentage < 90 {
		t.Errorf("expected percentage near 100%%, got %.1f%%", c.Percentage)
	}
}

func TestExtractPalette_CountZero(t *testing.T) {
	img := newSolidImage(10, 10, color.White)
	colors := ExtractPalette(img, 0)
	if colors != nil {
		t.Errorf("expected nil for count=0, got %v", colors)
	}
}

func TestExtractPalette_CountOne(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	img := newTwoColorImage(100, 100, red, blue)

	colors := ExtractPalette(img, 1)
	if len(colors) != 1 {
		t.Fatalf("expected 1 color, got %d", len(colors))
	}
}

func TestDominantColor(t *testing.T) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	img := newSolidImage(50, 50, white)

	c := DominantColor(img)
	if c.R < 240 || c.G < 240 || c.B < 240 {
		t.Errorf("expected white, got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
	if c.Hex == "" {
		t.Error("expected non-empty hex string")
	}
}

func TestDominantColor_EmptyImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 0, 0))
	c := DominantColor(img)
	if c.Hex != "" {
		t.Errorf("expected empty Color for empty image, got %+v", c)
	}
}

func TestExtractPalette_PercentagesSum(t *testing.T) {
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	blue := color.RGBA{R: 0, G: 0, B: 255, A: 255}
	img := newTwoColorImage(100, 100, red, blue)

	colors := ExtractPalette(img, 2)
	sum := 0.0
	for _, c := range colors {
		sum += c.Percentage
	}
	if sum < 99.0 || sum > 101.0 {
		t.Errorf("percentages sum to %.1f%%, expected ~100%%", sum)
	}
}
