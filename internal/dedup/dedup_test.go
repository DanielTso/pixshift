package dedup

import (
	"image"
	"image/color"
	"testing"
)

func solidImg(w, h int, c color.Color) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func checkerImg(w, h int) *image.RGBA {
	// Vertical stripes: alternating black and white columns. After
	// downsampling to 9x8, adjacent pixels will alternate bright/dark,
	// producing a non-zero DHash.
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if (x/(w/9+1))%2 == 0 {
				img.Set(x, y, color.White)
			} else {
				img.Set(x, y, color.Black)
			}
		}
	}
	return img
}

func gradientImg(w, h int) *image.RGBA {
	// Create a strong horizontal gradient so that adjacent pixels at any
	// resolution have clearly different brightness values.
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := uint8(x * 255 / w)
			img.Set(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	return img
}

func TestDHash_IdenticalImages_DistanceZero(t *testing.T) {
	img := solidImg(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	h1 := DHash(img)
	h2 := DHash(img)
	if h1 != h2 {
		t.Errorf("identical images should produce same hash: %x != %x", h1, h2)
	}
	dist := HammingDistance(h1, h2)
	if dist != 0 {
		t.Errorf("distance between identical hashes should be 0, got %d", dist)
	}
}

func TestDHash_DifferentImages_DistanceGreaterZero(t *testing.T) {
	// Use a checkerboard pattern: alternating bright/dark columns produce
	// different left-vs-right comparisons than a solid white image.
	white := solidImg(100, 100, color.White)
	checker := checkerImg(100, 100)
	h1 := DHash(white)
	h2 := DHash(checker)
	dist := HammingDistance(h1, h2)
	if dist == 0 {
		t.Errorf("different images should have distance > 0, got 0 (h1=%x, h2=%x)", h1, h2)
	}
}

func TestDHash_SimilarImages_SmallDistance(t *testing.T) {
	// Two nearly identical images (one pixel different) should have small distance.
	img1 := solidImg(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	img2 := solidImg(100, 100, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	img2.Set(50, 50, color.RGBA{R: 200, G: 200, B: 200, A: 255})

	h1 := DHash(img1)
	h2 := DHash(img2)
	dist := HammingDistance(h1, h2)
	// One pixel change in a 100x100 image, after downsampling to 9x8, should
	// produce a very small distance (likely 0 or 1).
	if dist > 5 {
		t.Errorf("nearly identical images should have small distance, got %d", dist)
	}
}

func TestHammingDistance_KnownValues(t *testing.T) {
	tests := []struct {
		a, b uint64
		want int
	}{
		{0, 0, 0},
		{0xFFFFFFFFFFFFFFFF, 0, 64},
		{0b1010, 0b0101, 4},
		{0b1111, 0b1110, 1},
		{0b1100, 0b1100, 0},
	}
	for _, tc := range tests {
		got := HammingDistance(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("HammingDistance(%x, %x) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestIsDuplicate_BelowThreshold(t *testing.T) {
	if !IsDuplicate(0b1100, 0b1100, 5) {
		t.Error("identical hashes should be duplicates at any threshold")
	}
}

func TestIsDuplicate_AtThreshold(t *testing.T) {
	// Distance 1, threshold 1 -> duplicate
	if !IsDuplicate(0b1110, 0b1111, 1) {
		t.Error("distance == threshold should be duplicate")
	}
}

func TestIsDuplicate_AboveThreshold(t *testing.T) {
	// Distance 4, threshold 2 -> not duplicate
	if IsDuplicate(0b1010, 0b0101, 2) {
		t.Error("distance > threshold should not be duplicate")
	}
}

func TestIsDuplicate_ThresholdZero(t *testing.T) {
	if !IsDuplicate(42, 42, 0) {
		t.Error("identical hashes with threshold 0 should be duplicate")
	}
	if IsDuplicate(42, 43, 0) {
		t.Error("different hashes with threshold 0 should not be duplicate")
	}
}

func TestDHash_Deterministic(t *testing.T) {
	img := gradientImg(200, 150)
	h1 := DHash(img)
	h2 := DHash(img)
	h3 := DHash(img)
	if h1 != h2 || h2 != h3 {
		t.Errorf("DHash should be deterministic: %x, %x, %x", h1, h2, h3)
	}
}

func TestDHash_BlackVsWhite(t *testing.T) {
	black := solidImg(50, 50, color.Black)
	white := solidImg(50, 50, color.White)
	hBlack := DHash(black)
	hWhite := DHash(white)
	// Solid images have no adjacent pixel differences, so both hashes should be 0.
	if hBlack != 0 {
		t.Errorf("solid black DHash should be 0, got %x", hBlack)
	}
	if hWhite != 0 {
		t.Errorf("solid white DHash should be 0, got %x", hWhite)
	}
}
