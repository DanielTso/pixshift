package transform

import (
	"image"
	"image/color"
	"testing"
)

func colorImage(w, h int, c color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	return img
}

func TestGrayscale_PreservesDimensions(t *testing.T) {
	img := colorImage(100, 80, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	out := Grayscale(img)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("dimensions: got %dx%d, want 100x80", b.Dx(), b.Dy())
	}
}

func TestGrayscale_AllChannelsEqual(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	out := Grayscale(img)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R != c.G || c.G != c.B {
		t.Errorf("grayscale pixel not uniform: R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestSepia_ZeroIntensity_Unchanged(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Sepia(img, 0)
	if out != img {
		t.Error("sepia with intensity 0 should return original image")
	}
}

func TestSepia_PreservesDimensions(t *testing.T) {
	img := colorImage(50, 30, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out := Sepia(img, 0.8)
	b := out.Bounds()
	if b.Dx() != 50 || b.Dy() != 30 {
		t.Errorf("dimensions: got %dx%d, want 50x30", b.Dx(), b.Dy())
	}
}

func TestSepia_ProducesDifferentOutput(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out := Sepia(img, 0.8)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R == 100 && c.G == 100 && c.B == 100 {
		t.Error("sepia should modify pixel values")
	}
}

func TestBrightness_Zero_Unchanged(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Brightness(img, 0)
	if out != img {
		t.Error("brightness 0 should return original image")
	}
}

func TestBrightness_Positive(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out := Brightness(img, 50)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R <= 100 {
		t.Errorf("positive brightness should increase R: got %d", c.R)
	}
}

func TestBrightness_Negative(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 100, G: 100, B: 100, A: 255})
	out := Brightness(img, -50)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R >= 100 {
		t.Errorf("negative brightness should decrease R: got %d", c.R)
	}
}

func TestContrast_Zero_Unchanged(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Contrast(img, 0)
	if out != img {
		t.Error("contrast 0 should return original image")
	}
}

func TestContrast_PreservesDimensions(t *testing.T) {
	img := colorImage(100, 50, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Contrast(img, 50)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 50 {
		t.Errorf("dimensions: got %dx%d, want 100x50", b.Dx(), b.Dy())
	}
}

func TestSharpen_PreservesDimensions(t *testing.T) {
	img := colorImage(100, 80, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Sharpen(img)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("dimensions: got %dx%d, want 100x80", b.Dx(), b.Dy())
	}
}

func TestSharpen_ProducesDifferentOutput(t *testing.T) {
	// Create an image with a non-linear pattern (step edge) for the kernel to work on.
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			var v uint8
			if x < 10 {
				v = 50
			} else {
				v = 200
			}
			img.SetRGBA(x, y, color.RGBA{R: v, G: v, B: v, A: 255})
		}
	}
	out := Sharpen(img)
	// Pixels near the edge should differ from original due to sharpening.
	rgba := out.(*image.RGBA)
	orig := img.RGBAAt(10, 10) // Right at the edge
	sharp := rgba.RGBAAt(10, 10)
	if orig.R == sharp.R && orig.G == sharp.G && orig.B == sharp.B {
		t.Error("sharpen should modify pixels near a step edge")
	}
}

func TestBlur_ZeroRadius_Unchanged(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Blur(img, 0)
	if out != img {
		t.Error("blur with radius 0 should return original image")
	}
}

func TestBlur_PreservesDimensions(t *testing.T) {
	img := colorImage(100, 80, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Blur(img, 2)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("dimensions: got %dx%d, want 100x80", b.Dx(), b.Dy())
	}
}

func TestBlur_SmoothsSharpEdge(t *testing.T) {
	// Half black, half white image.
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			if x < 10 {
				img.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
			}
		}
	}
	out := Blur(img, 2)
	rgba := out.(*image.RGBA)
	// Pixel at the edge should be between 0 and 255.
	c := rgba.RGBAAt(10, 10)
	if c.R == 0 || c.R == 255 {
		t.Errorf("blur should smooth edge pixel, got R=%d", c.R)
	}
}

func TestInvert_PreservesDimensions(t *testing.T) {
	img := colorImage(100, 80, color.RGBA{R: 128, G: 128, B: 128, A: 255})
	out := Invert(img)
	b := out.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("dimensions: got %dx%d, want 100x80", b.Dx(), b.Dy())
	}
}

func TestInvert_InvertsValues(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 200, G: 100, B: 50, A: 255})
	out := Invert(img)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R != 55 || c.G != 155 || c.B != 205 {
		t.Errorf("invert: got R=%d G=%d B=%d, want R=55 G=155 B=205", c.R, c.G, c.B)
	}
}

func TestInvert_DoubleInvert_RestoresOriginal(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 42, G: 137, B: 200, A: 255})
	out := Invert(Invert(img))
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R != 42 || c.G != 137 || c.B != 200 {
		t.Errorf("double invert: got R=%d G=%d B=%d, want R=42 G=137 B=200", c.R, c.G, c.B)
	}
}

func TestBrightness_ClampMax(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 250, G: 250, B: 250, A: 255})
	out := Brightness(img, 100)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R != 255 || c.G != 255 || c.B != 255 {
		t.Errorf("max brightness clamp: got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}

func TestBrightness_ClampMin(t *testing.T) {
	img := colorImage(10, 10, color.RGBA{R: 5, G: 5, B: 5, A: 255})
	out := Brightness(img, -100)
	rgba := out.(*image.RGBA)
	c := rgba.RGBAAt(5, 5)
	if c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("min brightness clamp: got R=%d G=%d B=%d", c.R, c.G, c.B)
	}
}
