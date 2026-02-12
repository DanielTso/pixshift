package transform

import (
	"fmt"
	"image"
	"image/draw"
	"strconv"
	"strings"
)

// CropOptions controls how an image is cropped.
type CropOptions struct {
	Width       int    // Explicit crop width in pixels.
	Height      int    // Explicit crop height in pixels.
	AspectRatio string // Desired ratio as "W:H" (e.g. "16:9"). Overrides Width/Height.
	Gravity     string // Anchor point: center (default), north, south, east, west.
}

// Crop extracts a rectangular region from img according to opts.
// If AspectRatio is set it takes precedence over Width/Height, calculating
// the largest crop that fits the source image with the requested ratio.
// If the requested crop is larger than the image, the original is returned.
func Crop(img image.Image, opts CropOptions) image.Image {
	b := img.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()

	cropW, cropH := opts.Width, opts.Height

	if opts.AspectRatio != "" {
		rw, rh, err := parseAspectRatio(opts.AspectRatio)
		if err == nil {
			cropW, cropH = fitAspectRatio(srcW, srcH, rw, rh)
		}
	}

	if cropW <= 0 || cropH <= 0 {
		return img
	}
	if cropW >= srcW && cropH >= srcH {
		return img
	}
	if cropW > srcW {
		cropW = srcW
	}
	if cropH > srcH {
		cropH = srcH
	}

	x, y := gravityOffset(srcW, srcH, cropW, cropH, opts.Gravity)
	rect := image.Rect(b.Min.X+x, b.Min.Y+y, b.Min.X+x+cropW, b.Min.Y+y+cropH)

	// Try zero-copy SubImage first.
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := img.(subImager); ok {
		return si.SubImage(rect)
	}

	dst := image.NewRGBA(image.Rect(0, 0, cropW, cropH))
	draw.Draw(dst, dst.Bounds(), img, rect.Min, draw.Src)
	return dst
}

// parseAspectRatio parses a "W:H" string into two integers.
func parseAspectRatio(s string) (int, int, error) {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid aspect ratio %q", s)
	}
	w, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || w <= 0 {
		return 0, 0, fmt.Errorf("invalid aspect ratio width %q", parts[0])
	}
	h, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || h <= 0 {
		return 0, 0, fmt.Errorf("invalid aspect ratio height %q", parts[1])
	}
	return w, h, nil
}

// fitAspectRatio returns the largest crop dimensions that fit within srcW x srcH
// while matching the aspect ratio rw:rh.
func fitAspectRatio(srcW, srcH, rw, rh int) (int, int) {
	// Try fitting width first.
	w := srcW
	h := srcW * rh / rw
	if h > srcH {
		h = srcH
		w = srcH * rw / rh
	}
	return w, h
}

// gravityOffset returns the top-left (x, y) offset for placing a cropW x cropH
// rectangle within a srcW x srcH source image according to the gravity string.
func gravityOffset(srcW, srcH, cropW, cropH int, gravity string) (int, int) {
	switch strings.ToLower(strings.TrimSpace(gravity)) {
	case "north":
		return (srcW - cropW) / 2, 0
	case "south":
		return (srcW - cropW) / 2, srcH - cropH
	case "east":
		return srcW - cropW, (srcH - cropH) / 2
	case "west":
		return 0, (srcH - cropH) / 2
	default: // "center" or empty
		return (srcW - cropW) / 2, (srcH - cropH) / 2
	}
}
