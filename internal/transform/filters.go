package transform

import (
	"image"
	"image/color"
	"image/draw"
	"math"
)

// Grayscale converts an image to grayscale using luminance formula.
func Grayscale(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			lum := uint8((0.299*float64(r) + 0.587*float64(g) + 0.114*float64(bl)) / 256)
			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{R: lum, G: lum, B: lum, A: uint8(a >> 8)})
		}
	}
	return dst
}

// Sepia applies a warm sepia tone. Intensity ranges from 0.0 (no effect) to 1.0 (full sepia).
func Sepia(img image.Image, intensity float64) image.Image {
	if intensity <= 0 {
		return img
	}
	if intensity > 1 {
		intensity = 1
	}
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			rf, gf, bf := float64(r>>8), float64(g>>8), float64(bl>>8)

			sr := rf*0.393 + gf*0.769 + bf*0.189
			sg := rf*0.349 + gf*0.686 + bf*0.168
			sb := rf*0.272 + gf*0.534 + bf*0.131

			nr := rf + (sr-rf)*intensity
			ng := gf + (sg-gf)*intensity
			nb := bf + (sb-bf)*intensity

			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{
				R: clampU8(nr),
				G: clampU8(ng),
				B: clampU8(nb),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// Brightness adjusts image brightness. Amount ranges from -100 (black) to +100 (white).
func Brightness(img image.Image, amount float64) image.Image {
	if amount == 0 {
		return img
	}
	adj := amount * 255 / 100
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{
				R: clampU8(float64(r>>8) + adj),
				G: clampU8(float64(g>>8) + adj),
				B: clampU8(float64(bl>>8) + adj),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// Contrast adjusts image contrast. Amount ranges from -100 (gray) to +100 (high contrast).
func Contrast(img image.Image, amount float64) image.Image {
	if amount == 0 {
		return img
	}
	factor := (259 * (amount + 255)) / (255 * (259 - amount))
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{
				R: clampU8(factor*(float64(r>>8)-128) + 128),
				G: clampU8(factor*(float64(g>>8)-128) + 128),
				B: clampU8(factor*(float64(bl>>8)-128) + 128),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

// Sharpen applies a 3x3 unsharp mask convolution kernel.
func Sharpen(img image.Image) image.Image {
	b := img.Bounds()
	src := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(src, src.Bounds(), img, b.Min, draw.Src)

	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	w, h := b.Dx(), b.Dy()

	// Sharpen kernel: [0 -1 0; -1 5 -1; 0 -1 0]
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if x == 0 || y == 0 || x == w-1 || y == h-1 {
				dst.SetRGBA(x, y, src.RGBAAt(x, y))
				continue
			}
			c := src.RGBAAt(x, y)
			t := src.RGBAAt(x, y-1)
			bo := src.RGBAAt(x, y+1)
			l := src.RGBAAt(x-1, y)
			r := src.RGBAAt(x+1, y)

			dst.SetRGBA(x, y, color.RGBA{
				R: clampU8(float64(c.R)*5 - float64(t.R) - float64(bo.R) - float64(l.R) - float64(r.R)),
				G: clampU8(float64(c.G)*5 - float64(t.G) - float64(bo.G) - float64(l.G) - float64(r.G)),
				B: clampU8(float64(c.B)*5 - float64(t.B) - float64(bo.B) - float64(l.B) - float64(r.B)),
				A: c.A,
			})
		}
	}
	return dst
}

// Blur applies a box blur with the given radius (in pixels).
func Blur(img image.Image, radius float64) image.Image {
	if radius <= 0 {
		return img
	}
	b := img.Bounds()
	src := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(src, src.Bounds(), img, b.Min, draw.Src)

	w, h := b.Dx(), b.Dy()
	r := int(math.Ceil(radius))
	size := (2*r + 1) * (2*r + 1)

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var sumR, sumG, sumB, sumA int
			for ky := -r; ky <= r; ky++ {
				for kx := -r; kx <= r; kx++ {
					sx := clampInt(x+kx, 0, w-1)
					sy := clampInt(y+ky, 0, h-1)
					c := src.RGBAAt(sx, sy)
					sumR += int(c.R)
					sumG += int(c.G)
					sumB += int(c.B)
					sumA += int(c.A)
				}
			}
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(sumR / size),
				G: uint8(sumG / size),
				B: uint8(sumB / size),
				A: uint8(sumA / size),
			})
		}
	}
	return dst
}

// Invert negates all color channels.
func Invert(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, color.RGBA{
				R: 255 - uint8(r>>8),
				G: 255 - uint8(g>>8),
				B: 255 - uint8(bl>>8),
				A: uint8(a >> 8),
			})
		}
	}
	return dst
}

func clampU8(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
