package transform

import (
	"image"
	"image/color"
	"testing"
)

// newTestImage creates a w x h RGBA image with distinct pixel values so
// orientation transforms can be verified by checking pixel positions.
func newTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 60) % 256),
				G: uint8((y * 80) % 256),
				B: 100,
				A: 255,
			})
		}
	}
	return img
}

func TestAutoRotate_Orientation0_Unchanged(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 0)
	if out != img {
		t.Error("orientation 0 should return the same image pointer")
	}
}

func TestAutoRotate_Orientation1_Unchanged(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 1)
	if out != img {
		t.Error("orientation 1 should return the same image pointer")
	}
}

func TestAutoRotate_NegativeOrientation_Unchanged(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, -1)
	if out != img {
		t.Error("negative orientation should return the same image pointer")
	}
}

func TestAutoRotate_Orientation2_FlipH(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 2)
	b := out.Bounds()
	if b.Dx() != 4 || b.Dy() != 3 {
		t.Errorf("orientation 2: got %dx%d, want 4x3", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation3_Rotate180(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 3)
	b := out.Bounds()
	if b.Dx() != 4 || b.Dy() != 3 {
		t.Errorf("orientation 3: got %dx%d, want 4x3", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation4_FlipV(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 4)
	b := out.Bounds()
	if b.Dx() != 4 || b.Dy() != 3 {
		t.Errorf("orientation 4: got %dx%d, want 4x3", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation5_FlipH_Rotate90CW(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 5)
	b := out.Bounds()
	// Orientation 5 = flipH(rotate90CW) -> dimensions swap
	if b.Dx() != 3 || b.Dy() != 4 {
		t.Errorf("orientation 5: got %dx%d, want 3x4", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation6_Rotate90CW(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 6)
	b := out.Bounds()
	// 90CW swaps width and height
	if b.Dx() != 3 || b.Dy() != 4 {
		t.Errorf("orientation 6: got %dx%d, want 3x4", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation7_FlipH_Rotate90CCW(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 7)
	b := out.Bounds()
	// Orientation 7 = flipH(rotate90CCW) -> dimensions swap
	if b.Dx() != 3 || b.Dy() != 4 {
		t.Errorf("orientation 7: got %dx%d, want 3x4", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation8_Rotate90CCW(t *testing.T) {
	img := newTestImage(4, 3)
	out := AutoRotate(img, 8)
	b := out.Bounds()
	// 90CCW swaps width and height
	if b.Dx() != 3 || b.Dy() != 4 {
		t.Errorf("orientation 8: got %dx%d, want 3x4", b.Dx(), b.Dy())
	}
}

func TestAutoRotate_Orientation6_PixelMapping(t *testing.T) {
	// Verify that a 90CW rotation actually moves pixels correctly.
	// For a 4x3 image rotated 90CW to 3x4:
	// src(0,0) -> dst(2,0) in the rotate90CW implementation:
	//   dst.Set(b.Max.Y-1-y, x, img.At(x, y))
	//   dst.Set(2-0, 0, ...) = dst.Set(2, 0)
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	marker := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	img.Set(0, 0, marker)

	out := AutoRotate(img, 6)
	r, g, b, a := out.At(2, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
		t.Errorf("pixel at (2,0) after 90CW should be red marker, got RGBA(%d,%d,%d,%d)",
			r>>8, g>>8, b>>8, a>>8)
	}
}

func TestAutoRotate_SquareImage(t *testing.T) {
	// For a square image, all orientations should produce same dimensions
	img := newTestImage(5, 5)
	for orient := 0; orient <= 8; orient++ {
		out := AutoRotate(img, orient)
		b := out.Bounds()
		if b.Dx() != 5 || b.Dy() != 5 {
			t.Errorf("orientation %d with square: got %dx%d, want 5x5",
				orient, b.Dx(), b.Dy())
		}
	}
}
