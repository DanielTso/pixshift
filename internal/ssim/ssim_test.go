package ssim

import (
	"image"
	"image/color"
	"math"
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

func gradientImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8((x * 255) / w)
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func checkerboardImage(w, h, blockSize int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if ((x/blockSize)+(y/blockSize))%2 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}
	return img
}

func TestCompare_IdenticalImages_Returns1(t *testing.T) {
	img := gradientImage(64, 64)
	score := Compare(img, img)
	if math.Abs(score-1.0) > 0.001 {
		t.Errorf("identical images: SSIM = %f, want ~1.0", score)
	}
}

func TestCompare_IdenticalSolidImages_Returns1(t *testing.T) {
	img1 := solidImage(64, 64, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	img2 := solidImage(64, 64, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	score := Compare(img1, img2)
	if math.Abs(score-1.0) > 0.001 {
		t.Errorf("identical solid images: SSIM = %f, want ~1.0", score)
	}
}

func TestCompare_VeryDifferentImages_LowScore(t *testing.T) {
	white := solidImage(64, 64, color.White)
	checker := checkerboardImage(64, 64, 4)
	score := Compare(white, checker)
	if score > 0.5 {
		t.Errorf("very different images: SSIM = %f, want < 0.5", score)
	}
}

func TestCompare_SlightlyDifferent_HighScore(t *testing.T) {
	img1 := gradientImage(64, 64)
	img2 := gradientImage(64, 64)
	// Modify a few pixels
	for i := 0; i < 5; i++ {
		img2.Set(i, i, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}
	score := Compare(img1, img2)
	if score < 0.8 {
		t.Errorf("slightly different images: SSIM = %f, want > 0.8", score)
	}
}

func TestCompare_DifferentSizes_ResizesImg2(t *testing.T) {
	img1 := gradientImage(64, 64)
	img2 := gradientImage(128, 128) // Larger; should be resized down
	score := Compare(img1, img2)
	// After resize, the gradient content should be similar
	if score < 0.5 {
		t.Errorf("resized images: SSIM = %f, want > 0.5", score)
	}
}

func TestCompare_SmallImage_NoWindows(t *testing.T) {
	// Image smaller than window size (8x8) should return 1.0 (no windows)
	img1 := solidImage(4, 4, color.White)
	img2 := solidImage(4, 4, color.Black)
	score := Compare(img1, img2)
	// No windows fit, so count == 0 and function returns 1.0
	if score != 1.0 {
		t.Errorf("image smaller than window: SSIM = %f, want 1.0", score)
	}
}

func TestCompare_BlackVsWhite(t *testing.T) {
	black := solidImage(64, 64, color.Black)
	white := solidImage(64, 64, color.White)
	score := Compare(black, white)
	if score > 0.1 {
		t.Errorf("black vs white: SSIM = %f, want < 0.1", score)
	}
}

func TestRating_Excellent(t *testing.T) {
	if r := Rating(0.99); r != "Excellent" {
		t.Errorf("Rating(0.99) = %q, want Excellent", r)
	}
}

func TestRating_Good(t *testing.T) {
	if r := Rating(0.96); r != "Good" {
		t.Errorf("Rating(0.96) = %q, want Good", r)
	}
}

func TestRating_Acceptable(t *testing.T) {
	if r := Rating(0.91); r != "Acceptable" {
		t.Errorf("Rating(0.91) = %q, want Acceptable", r)
	}
}

func TestRating_Poor(t *testing.T) {
	if r := Rating(0.85); r != "Poor" {
		t.Errorf("Rating(0.85) = %q, want Poor", r)
	}
}

func TestRating_Bad(t *testing.T) {
	if r := Rating(0.5); r != "Bad" {
		t.Errorf("Rating(0.5) = %q, want Bad", r)
	}
}

func TestRating_BoundaryValues(t *testing.T) {
	tests := []struct {
		score float64
		want  string
	}{
		{1.0, "Excellent"},
		{0.98, "Good"},       // <= 0.98 is Good
		{0.981, "Excellent"}, // > 0.98 is Excellent
		{0.95, "Acceptable"}, // <= 0.95 is Acceptable
		{0.951, "Good"},      // > 0.95 is Good
		{0.90, "Poor"},       // <= 0.90 is Poor
		{0.901, "Acceptable"},
		{0.80, "Bad"},  // <= 0.80 is Bad
		{0.801, "Poor"},
		{0.0, "Bad"},
	}
	for _, tc := range tests {
		got := Rating(tc.score)
		if got != tc.want {
			t.Errorf("Rating(%g) = %q, want %q", tc.score, got, tc.want)
		}
	}
}

func TestCompare_Symmetric(t *testing.T) {
	img1 := gradientImage(64, 64)
	img2 := checkerboardImage(64, 64, 8)
	s1 := Compare(img1, img2)
	s2 := Compare(img2, img1)
	// SSIM is generally symmetric for same-size images
	if math.Abs(s1-s2) > 0.05 {
		t.Errorf("SSIM should be roughly symmetric: %f vs %f", s1, s2)
	}
}
