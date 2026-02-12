package dedup

import (
	"fmt"
	"image"
	"math/bits"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"golang.org/x/image/draw"
)

// DHash computes a 64-bit difference hash for perceptual image comparison.
// It resizes the image to 9x8 grayscale and compares adjacent horizontal pixels.
func DHash(img image.Image) uint64 {
	gray := image.NewGray(image.Rect(0, 0, 9, 8))
	draw.NearestNeighbor.Scale(gray, gray.Bounds(), img, img.Bounds(), draw.Over, nil)

	var hash uint64
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			left := gray.GrayAt(x, y).Y
			right := gray.GrayAt(x+1, y).Y
			if left > right {
				hash |= 1 << uint(y*8+x)
			}
		}
	}
	return hash
}

// HammingDistance returns the number of differing bits between two hashes.
func HammingDistance(a, b uint64) int {
	return bits.OnesCount64(a ^ b)
}

// IsDuplicate returns true if the Hamming distance between two hashes
// is less than or equal to the threshold.
func IsDuplicate(a, b uint64, threshold int) bool {
	return HammingDistance(a, b) <= threshold
}

// HashFile opens a file, decodes it using the registry, and computes its DHash.
func HashFile(path string, registry *codec.Registry) (uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	format, err := codec.DetectFormat(f, path)
	if err != nil {
		return 0, fmt.Errorf("detect format %s: %w", path, err)
	}

	dec, err := registry.Decoder(format)
	if err != nil {
		return 0, fmt.Errorf("decoder for %s: %w", format, err)
	}

	img, err := dec.Decode(f)
	if err != nil {
		return 0, fmt.Errorf("decode %s: %w", path, err)
	}

	return DHash(img), nil
}
