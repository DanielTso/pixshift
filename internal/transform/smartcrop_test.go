package transform

import (
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// fillSolid fills the given rectangle of img with a solid color.
func fillSolid(img *image.RGBA, r image.Rectangle, c color.Color) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

// fillRandom fills the given rectangle of img with random colored pixels.
func fillRandom(img *image.RGBA, r image.Rectangle, rng *rand.Rand) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(rng.Intn(256)),
				G: uint8(rng.Intn(256)),
				B: uint8(rng.Intn(256)),
				A: 255,
			})
		}
	}
}

func TestSmartCrop_InterestRegion(t *testing.T) {
	// Left half: solid gray (low entropy). Right half: random colors (high entropy).
	// SmartCrop to 100x100 should select the right half.
	rng := rand.New(rand.NewSource(42))
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	fillSolid(img, image.Rect(0, 0, 100, 100), color.RGBA{R: 128, G: 128, B: 128, A: 255})
	fillRandom(img, image.Rect(100, 0, 200, 100), rng)

	out := SmartCrop(img, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Fatalf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}

	// The selected region should start at x>=100 (the high-entropy right half).
	// Check that the top-left pixel of the result is NOT the solid gray.
	r, g, bl, _ := out.At(b.Min.X, b.Min.Y).RGBA()
	gray := uint32(128) << 8
	if r == gray && g == gray && bl == gray {
		t.Error("smart crop selected the solid gray region instead of the high-entropy region")
	}
}

func TestSmartCrop_CenterInterest(t *testing.T) {
	// 200x200 image: white background, center 100x100 has random colors.
	// SmartCrop to 100x100 should select approximately the center.
	rng := rand.New(rand.NewSource(99))
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	fillSolid(img, image.Rect(0, 0, 200, 200), color.White)
	fillRandom(img, image.Rect(50, 50, 150, 150), rng)

	out := SmartCrop(img, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Fatalf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}

	// The center pixel of the output should be a random color, not white.
	cx := b.Min.X + b.Dx()/2
	cy := b.Min.Y + b.Dy()/2
	r, g, bl, _ := out.At(cx, cy).RGBA()
	if r == 0xffff && g == 0xffff && bl == 0xffff {
		t.Error("center pixel is white; expected smart crop to capture the colorful center")
	}
}

func TestSmartCrop_Passthrough(t *testing.T) {
	// SmartCrop on a 100x100 image with target 100x100 should return same image.
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	out := SmartCrop(img, 100, 100)
	if out != img {
		t.Error("same-size smart crop should return the original image")
	}
}

func TestSmartCrop_LargerTarget(t *testing.T) {
	// SmartCrop on a 50x50 image with target 100x100 should return unchanged.
	img := image.NewRGBA(image.Rect(0, 0, 50, 50))
	out := SmartCrop(img, 100, 100)
	if out != img {
		t.Error("target larger than source should return original image")
	}
}

func TestSmartCrop_ExactDimensions(t *testing.T) {
	// SmartCrop on a 400x200 image to 100x100 should return exactly 100x100.
	rng := rand.New(rand.NewSource(7))
	img := image.NewRGBA(image.Rect(0, 0, 400, 200))
	fillRandom(img, image.Rect(0, 0, 400, 200), rng)

	out := SmartCrop(img, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestSmartCrop_VerticalInterest(t *testing.T) {
	// 100x200 image: top half solid blue, bottom half random.
	// SmartCrop to 100x100 should pick the bottom (high entropy).
	rng := rand.New(rand.NewSource(12))
	img := image.NewRGBA(image.Rect(0, 0, 100, 200))
	fillSolid(img, image.Rect(0, 0, 100, 100), color.RGBA{B: 200, A: 255})
	fillRandom(img, image.Rect(0, 100, 100, 200), rng)

	out := SmartCrop(img, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Fatalf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}

	// The majority of the crop should be from the random region (bottom half).
	// Check a pixel well into the crop — at the vertical midpoint it should
	// be random, not the solid blue from the top.
	midY := b.Min.Y + b.Dy()/2
	r, g, bl, _ := out.At(b.Min.X+50, midY).RGBA()
	solidBlue := color.RGBA{B: 200, A: 255}
	er, eg, eb, _ := solidBlue.RGBA()
	if r == er && g == eg && bl == eb {
		t.Error("smart crop selected the solid blue region instead of the high-entropy bottom")
	}
}

func TestSmartCrop_SmallImage(t *testing.T) {
	// Image smaller than block size should use center crop.
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	out := SmartCrop(img, 5, 5)
	b := out.Bounds()
	if b.Dx() != 5 || b.Dy() != 5 {
		t.Errorf("expected 5x5, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestSmartCrop_WidthOnlyCrop(t *testing.T) {
	// Target height matches source, only width needs cropping.
	rng := rand.New(rand.NewSource(55))
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	fillSolid(img, image.Rect(0, 0, 200, 100), color.RGBA{R: 100, G: 100, B: 100, A: 255})
	fillRandom(img, image.Rect(50, 0, 150, 100), rng)

	out := SmartCrop(img, 100, 100)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 100 {
		t.Errorf("expected 100x100, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestSmartCrop_PartiallyLargerTarget(t *testing.T) {
	// Target wider than source but shorter: should clamp width.
	rng := rand.New(rand.NewSource(33))
	img := image.NewRGBA(image.Rect(0, 0, 64, 128))
	fillRandom(img, img.Bounds(), rng)

	out := SmartCrop(img, 100, 64)
	b := out.Bounds()
	if b.Dx() != 64 || b.Dy() != 64 {
		t.Errorf("expected 64x64, got %dx%d", b.Dx(), b.Dy())
	}
}
