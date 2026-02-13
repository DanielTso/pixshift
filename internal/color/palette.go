package color

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"os"
	"sort"

	"github.com/DanielTso/pixshift/internal/codec"
)

// Color represents a dominant color extracted from an image.
type Color struct {
	R          uint8   `json:"r"`
	G          uint8   `json:"g"`
	B          uint8   `json:"b"`
	Hex        string  `json:"hex"`
	Percentage float64 `json:"percentage"`
}

// pixel holds an RGB value sampled from the image.
type pixel struct {
	r, g, b float64
}

// ExtractPalette extracts the N most dominant colors from an image using K-means clustering.
func ExtractPalette(img image.Image, count int) []Color {
	if count <= 0 {
		return nil
	}

	pixels := samplePixels(img)
	if len(pixels) == 0 {
		return nil
	}

	// Clamp count to number of unique-ish pixels available.
	if count > len(pixels) {
		count = len(pixels)
	}

	centroids := initCentroids(pixels, count)

	// Run K-means for 10 iterations.
	assignments := make([]int, len(pixels))
	for iter := 0; iter < 10; iter++ {
		// Assign each pixel to the nearest centroid.
		for i, p := range pixels {
			best := 0
			bestDist := math.MaxFloat64
			for j, c := range centroids {
				d := distSq(p, c)
				if d < bestDist {
					bestDist = d
					best = j
				}
			}
			assignments[i] = best
		}

		// Recalculate centroids as the mean of assigned pixels.
		sums := make([]pixel, count)
		counts := make([]float64, count)
		for i, p := range pixels {
			a := assignments[i]
			sums[a].r += p.r
			sums[a].g += p.g
			sums[a].b += p.b
			counts[a]++
		}
		for j := range centroids {
			if counts[j] > 0 {
				centroids[j] = pixel{
					r: sums[j].r / counts[j],
					g: sums[j].g / counts[j],
					b: sums[j].b / counts[j],
				}
			}
		}
	}

	// Count pixels per cluster.
	clusterCounts := make([]int, count)
	for _, a := range assignments {
		clusterCounts[a]++
	}

	total := float64(len(pixels))
	type clusterInfo struct {
		centroid pixel
		count    int
	}
	clusters := make([]clusterInfo, count)
	for i := range clusters {
		clusters[i] = clusterInfo{centroid: centroids[i], count: clusterCounts[i]}
	}

	// Sort by count descending.
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].count > clusters[j].count
	})

	colors := make([]Color, 0, count)
	for _, cl := range clusters {
		if cl.count == 0 {
			continue
		}
		r := uint8(math.Round(cl.centroid.r))
		g := uint8(math.Round(cl.centroid.g))
		b := uint8(math.Round(cl.centroid.b))
		colors = append(colors, Color{
			R:          r,
			G:          g,
			B:          b,
			Hex:        fmt.Sprintf("#%02X%02X%02X", r, g, b),
			Percentage: float64(cl.count) / total * 100,
		})
	}
	return colors
}

// DominantColor returns the single most dominant color in an image.
func DominantColor(img image.Image) Color {
	colors := ExtractPalette(img, 1)
	if len(colors) > 0 {
		return colors[0]
	}
	return Color{}
}

// ExtractPaletteFromFile opens an image file, decodes it, and extracts colors.
func ExtractPaletteFromFile(path string, count int, reg *codec.Registry) ([]Color, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	format, err := codec.DetectFormat(f, path)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	dec, err := reg.Decoder(format)
	if err != nil {
		return nil, fmt.Errorf("decoder for %s: %w", format, err)
	}

	img, err := dec.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}

	return ExtractPalette(img, count), nil
}

// samplePixels samples pixels from an image, taking every Nth pixel
// where N = max(1, totalPixels/10000) to keep processing fast.
func samplePixels(img image.Image) []pixel {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	total := w * h
	if total == 0 {
		return nil
	}

	step := total / 10000
	if step < 1 {
		step = 1
	}

	pixels := make([]pixel, 0, min(total, 10000))
	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if i%step == 0 {
				r, g, b, _ := img.At(x, y).RGBA()
				pixels = append(pixels, pixel{
					r: float64(r >> 8),
					g: float64(g >> 8),
					b: float64(b >> 8),
				})
			}
			i++
		}
	}
	return pixels
}

// initCentroids picks count random pixels as initial centroids.
func initCentroids(pixels []pixel, count int) []pixel {
	centroids := make([]pixel, count)
	perm := rand.Perm(len(pixels))
	for i := 0; i < count; i++ {
		centroids[i] = pixels[perm[i]]
	}
	return centroids
}

// distSq returns the squared Euclidean distance between two pixels in RGB space.
func distSq(a, b pixel) float64 {
	dr := a.r - b.r
	dg := a.g - b.g
	db := a.b - b.b
	return dr*dr + dg*dg + db*db
}
