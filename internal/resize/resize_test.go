package resize

import (
	"image"
	"image/color"
	"testing"
)

func testImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	return img
}

func TestResize_WidthOnly(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Width: 400})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(Width=400) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_HeightOnly(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Height: 300})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(Height=300) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_MaxDim(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{MaxDim: 400})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(MaxDim=400) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_BothDimensions(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Width: 400, Height: 400})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 400 {
		t.Errorf("Resize(Width=400, Height=400) on 800x600 = %dx%d, want 400x400",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_AllZeros(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{})
	// Should return the same image pointer
	if result != img {
		t.Error("Resize with all zeros should return original image (same pointer)")
	}
}

func TestResize_NoUpscale(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{MaxDim: 1000})
	// Should return the same image since 800 < 1000
	if result != img {
		t.Error("Resize(MaxDim=1000) on 800x600 should return original (no upscale)")
	}
}

func TestResize_Interpolation_Nearest(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Width: 400, Interpolation: "nearest"})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(nearest) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_Interpolation_Bilinear(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Width: 400, Interpolation: "bilinear"})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(bilinear) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_Interpolation_Default(t *testing.T) {
	img := testImage(800, 600)
	result := Resize(img, ResizeOptions{Width: 400, Interpolation: ""})
	bounds := result.Bounds()
	if bounds.Dx() != 400 || bounds.Dy() != 300 {
		t.Errorf("Resize(default) on 800x600 = %dx%d, want 400x300",
			bounds.Dx(), bounds.Dy())
	}
}

func TestResize_MaxDim_Portrait(t *testing.T) {
	img := testImage(600, 800)
	result := Resize(img, ResizeOptions{MaxDim: 400})
	bounds := result.Bounds()
	if bounds.Dx() != 300 || bounds.Dy() != 400 {
		t.Errorf("Resize(MaxDim=400) on 600x800 = %dx%d, want 300x400",
			bounds.Dx(), bounds.Dy())
	}
}
