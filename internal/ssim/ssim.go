package ssim

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"golang.org/x/image/draw"
)

const (
	c1         = 6.5025  // (0.01 * 255)^2
	c2         = 58.5225 // (0.03 * 255)^2
	windowSize = 8
)

// Compare computes the Structural Similarity Index between two images.
// Returns a value between 0.0 (completely different) and 1.0 (identical).
// If the images differ in size, img2 is resized to match img1.
func Compare(img1, img2 image.Image) float64 {
	px1, w, h := toGrayFloat(img1)

	b2 := img2.Bounds()
	if b2.Dx() != w || b2.Dy() != h {
		resized := image.NewNRGBA(image.Rect(0, 0, w, h))
		draw.BiLinear.Scale(resized, resized.Bounds(), img2, img2.Bounds(), draw.Over, nil)
		img2 = resized
	}
	px2, _, _ := toGrayFloat(img2)

	var sum float64
	var count int

	for y := 0; y <= h-windowSize; y++ {
		for x := 0; x <= w-windowSize; x++ {
			var sumX, sumY, sumXX, sumYY, sumXY float64

			for wy := 0; wy < windowSize; wy++ {
				for wx := 0; wx < windowSize; wx++ {
					idx := (y+wy)*w + (x + wx)
					vx := px1[idx]
					vy := px2[idx]
					sumX += vx
					sumY += vy
					sumXX += vx * vx
					sumYY += vy * vy
					sumXY += vx * vy
				}
			}

			n := float64(windowSize * windowSize)
			muX := sumX / n
			muY := sumY / n
			sigmaXX := sumXX/n - muX*muX
			sigmaYY := sumYY/n - muY*muY
			sigmaXY := sumXY/n - muX*muY

			numerator := (2*muX*muY + c1) * (2*sigmaXY + c2)
			denominator := (muX*muX + muY*muY + c1) * (sigmaXX + sigmaYY + c2)

			sum += numerator / denominator
			count++
		}
	}

	if count == 0 {
		return 1.0
	}
	return math.Max(0, math.Min(1, sum/float64(count)))
}

// CompareFiles decodes two image files and computes their SSIM.
func CompareFiles(path1, path2 string, registry *codec.Registry) (float64, error) {
	img1, err := decodeFile(path1, registry)
	if err != nil {
		return 0, err
	}
	img2, err := decodeFile(path2, registry)
	if err != nil {
		return 0, err
	}
	return Compare(img1, img2), nil
}

// Rating returns a human-readable quality label for an SSIM score.
func Rating(score float64) string {
	switch {
	case score > 0.98:
		return "Excellent"
	case score > 0.95:
		return "Good"
	case score > 0.90:
		return "Acceptable"
	case score > 0.80:
		return "Poor"
	default:
		return "Bad"
	}
}

// toGrayFloat converts an image to a flat grayscale float64 slice.
func toGrayFloat(img image.Image) ([]float64, int, int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	px := make([]float64, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x+img.Bounds().Min.X, y+img.Bounds().Min.Y).RGBA()
			gray := color.GrayModel.Convert(color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: 255,
			}).(color.Gray)
			px[y*w+x] = float64(gray.Y)
		}
	}
	return px, w, h
}

func decodeFile(path string, registry *codec.Registry) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	format, err := codec.DetectFormat(f, path)
	if err != nil {
		return nil, fmt.Errorf("detect format %s: %w", path, err)
	}

	dec, err := registry.Decoder(format)
	if err != nil {
		return nil, fmt.Errorf("decoder for %s: %w", format, err)
	}

	img, err := dec.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return img, nil
}
