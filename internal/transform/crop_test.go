package transform

import (
	"image"
	"image/color"
	"testing"
)

func newCropTestImage(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 50,
				A: 255,
			})
		}
	}
	return img
}

func TestCrop_ExplicitWidthHeight(t *testing.T) {
	img := newCropTestImage(100, 80)
	out := Crop(img, CropOptions{Width: 50, Height: 40})
	b := out.Bounds()
	if b.Dx() != 50 || b.Dy() != 40 {
		t.Errorf("explicit crop: got %dx%d, want 50x40", b.Dx(), b.Dy())
	}
}

func TestCrop_AspectRatio16x9(t *testing.T) {
	img := newCropTestImage(160, 90)
	out := Crop(img, CropOptions{AspectRatio: "16:9"})
	b := out.Bounds()
	// 160x90 with 16:9 -> fit width first: 160x90, which fits perfectly
	if b.Dx() != 160 || b.Dy() != 90 {
		t.Errorf("16:9 crop of 160x90: got %dx%d, want 160x90", b.Dx(), b.Dy())
	}
}

func TestCrop_AspectRatio16x9_TallImage(t *testing.T) {
	img := newCropTestImage(160, 200)
	out := Crop(img, CropOptions{AspectRatio: "16:9"})
	b := out.Bounds()
	// 160x200 with 16:9 -> fit width: 160, h = 160*9/16 = 90
	if b.Dx() != 160 || b.Dy() != 90 {
		t.Errorf("16:9 crop of 160x200: got %dx%d, want 160x90", b.Dx(), b.Dy())
	}
}

func TestCrop_AspectRatio16x9_WideImage(t *testing.T) {
	img := newCropTestImage(320, 90)
	out := Crop(img, CropOptions{AspectRatio: "16:9"})
	b := out.Bounds()
	// 320x90 with 16:9 -> fit width: 320, h = 320*9/16 = 180 > 90
	// -> fit height: h=90, w = 90*16/9 = 160
	if b.Dx() != 160 || b.Dy() != 90 {
		t.Errorf("16:9 crop of 320x90: got %dx%d, want 160x90", b.Dx(), b.Dy())
	}
}

func TestCrop_GravityCenter(t *testing.T) {
	img := newCropTestImage(100, 100)
	out := Crop(img, CropOptions{Width: 50, Height: 50, Gravity: "center"})
	b := out.Bounds()
	if b.Dx() != 50 || b.Dy() != 50 {
		t.Errorf("center crop: got %dx%d, want 50x50", b.Dx(), b.Dy())
	}
}

func TestCrop_GravityNorth(t *testing.T) {
	// North: top-center region
	img := newCropTestImage(100, 100)
	out := Crop(img, CropOptions{Width: 40, Height: 30, Gravity: "north"})
	b := out.Bounds()
	if b.Dx() != 40 || b.Dy() != 30 {
		t.Errorf("north crop: got %dx%d, want 40x30", b.Dx(), b.Dy())
	}
}

func TestCrop_GravitySouth(t *testing.T) {
	img := newCropTestImage(100, 100)
	out := Crop(img, CropOptions{Width: 40, Height: 30, Gravity: "south"})
	b := out.Bounds()
	if b.Dx() != 40 || b.Dy() != 30 {
		t.Errorf("south crop: got %dx%d, want 40x30", b.Dx(), b.Dy())
	}
}

func TestCrop_GravityEast(t *testing.T) {
	img := newCropTestImage(100, 100)
	out := Crop(img, CropOptions{Width: 40, Height: 30, Gravity: "east"})
	b := out.Bounds()
	if b.Dx() != 40 || b.Dy() != 30 {
		t.Errorf("east crop: got %dx%d, want 40x30", b.Dx(), b.Dy())
	}
}

func TestCrop_GravityWest(t *testing.T) {
	img := newCropTestImage(100, 100)
	out := Crop(img, CropOptions{Width: 40, Height: 30, Gravity: "west"})
	b := out.Bounds()
	if b.Dx() != 40 || b.Dy() != 30 {
		t.Errorf("west crop: got %dx%d, want 40x30", b.Dx(), b.Dy())
	}
}

func TestCrop_Noop_CropLargerThanImage(t *testing.T) {
	img := newCropTestImage(50, 40)
	out := Crop(img, CropOptions{Width: 100, Height: 100})
	// Crop >= src in both dimensions -> original returned
	if out != img {
		t.Error("crop larger than image should return original image pointer")
	}
}

func TestCrop_Noop_ZeroDimensions(t *testing.T) {
	img := newCropTestImage(50, 40)
	out := Crop(img, CropOptions{Width: 0, Height: 0})
	if out != img {
		t.Error("zero crop dimensions should return original image pointer")
	}
}

func TestCrop_SubImageZeroCopy(t *testing.T) {
	// *image.RGBA supports SubImage, so Crop should use that path.
	img := newCropTestImage(100, 80)
	out := Crop(img, CropOptions{Width: 50, Height: 40})
	// The output should be a sub-image (shares underlying pixel data)
	if _, ok := out.(*image.RGBA); !ok {
		t.Errorf("expected *image.RGBA from SubImage path, got %T", out)
	}
	b := out.Bounds()
	if b.Dx() != 50 || b.Dy() != 40 {
		t.Errorf("SubImage crop: got %dx%d, want 50x40", b.Dx(), b.Dy())
	}
}

func TestCrop_InvalidAspectRatio_Fallback(t *testing.T) {
	img := newCropTestImage(100, 80)
	// Invalid aspect ratio should be ignored; with no width/height, return original
	out := Crop(img, CropOptions{AspectRatio: "invalid"})
	if out != img {
		t.Error("invalid aspect ratio with no explicit dimensions should return original")
	}
}

func TestCrop_AspectRatioOverridesExplicitDimensions(t *testing.T) {
	img := newCropTestImage(160, 120)
	out := Crop(img, CropOptions{
		Width:       10,
		Height:      10,
		AspectRatio: "4:3",
	})
	b := out.Bounds()
	// AspectRatio takes precedence: 160x120 with 4:3 -> 160x120 (perfect fit)
	// Since crop == src, return original
	if out != img {
		t.Errorf("4:3 crop of 160x120 should be a no-op; got %dx%d", b.Dx(), b.Dy())
	}
}

func TestCrop_CropWidthExceedsSource(t *testing.T) {
	img := newCropTestImage(50, 80)
	out := Crop(img, CropOptions{Width: 100, Height: 40})
	b := out.Bounds()
	// Width clamped to 50, height 40
	if b.Dx() != 50 || b.Dy() != 40 {
		t.Errorf("clamped crop: got %dx%d, want 50x40", b.Dx(), b.Dy())
	}
}

func TestCrop_CropHeightExceedsSource(t *testing.T) {
	img := newCropTestImage(80, 50)
	out := Crop(img, CropOptions{Width: 40, Height: 100})
	b := out.Bounds()
	// Height clamped to 50, width 40
	if b.Dx() != 40 || b.Dy() != 50 {
		t.Errorf("clamped crop: got %dx%d, want 40x50", b.Dx(), b.Dy())
	}
}

func TestCrop_GravityNorth_PixelVerification(t *testing.T) {
	// North gravity with center-x: verify the top-left pixel comes from
	// the correct region.
	img := newCropTestImage(100, 100)
	// Mark pixel at (30, 0) - this should be in the north crop region
	marker := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	img.Set(30, 0, marker)

	out := Crop(img, CropOptions{Width: 100, Height: 50, Gravity: "north"})
	// Gravity north: x offset = (100-100)/2 = 0, y offset = 0
	r, g, b, _ := out.At(out.Bounds().Min.X+30, out.Bounds().Min.Y).RGBA()
	if r>>8 != 255 || g>>8 != 255 || b>>8 != 255 {
		t.Errorf("north crop pixel check failed: got (%d,%d,%d)", r>>8, g>>8, b>>8)
	}
}
