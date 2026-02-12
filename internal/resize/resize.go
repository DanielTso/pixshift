package resize

import (
	"image"

	"golang.org/x/image/draw"
)

// ResizeOptions controls how an image is resized.
type ResizeOptions struct {
	Width  int // Target width. If Height is 0, aspect ratio is preserved.
	Height int // Target height. If Width is 0, aspect ratio is preserved.
	MaxDim int // Maximum dimension. Scales so max(width, height) <= MaxDim.
}

// Resize scales img according to opts using high-quality CatmullRom interpolation.
// If no resize is needed (all options zero) or the image is already within the
// target dimensions, the original image is returned unchanged.
func Resize(img image.Image, opts ResizeOptions) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()

	newW, newH := targetDimensions(srcW, srcH, opts)

	// Nothing to do — all options are zero.
	if newW == 0 && newH == 0 {
		return img
	}

	// Don't upscale: if the image already fits, return as-is.
	if newW >= srcW && newH >= srcH {
		return img
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)
	return dst
}

// targetDimensions computes the final width and height from the source
// dimensions and the provided options.
func targetDimensions(srcW, srcH int, opts ResizeOptions) (int, int) {
	if opts.MaxDim > 0 {
		maxSrc := srcW
		if srcH > maxSrc {
			maxSrc = srcH
		}
		if maxSrc <= opts.MaxDim {
			return srcW, srcH
		}
		if srcW >= srcH {
			return opts.MaxDim, srcH * opts.MaxDim / srcW
		}
		return srcW * opts.MaxDim / srcH, opts.MaxDim
	}

	w, h := opts.Width, opts.Height

	switch {
	case w > 0 && h > 0:
		// Both set: use exact dimensions.
		return w, h
	case w > 0:
		// Only width set: preserve aspect ratio.
		return w, srcH * w / srcW
	case h > 0:
		// Only height set: preserve aspect ratio.
		return srcW * h / srcH, h
	default:
		// All zero — no resize.
		return 0, 0
	}
}
