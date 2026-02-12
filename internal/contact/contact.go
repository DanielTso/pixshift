package contact

import (
	"image"
	"image/color"
	"image/draw"
	"path/filepath"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/math/fixed"
)

// ImageEntry represents a single image to include in the contact sheet.
type ImageEntry struct {
	Path  string
	Image image.Image
	Label string
}

// Options controls contact sheet layout and appearance.
type Options struct {
	Columns     int
	ThumbSize   int
	Padding     int
	Background  color.Color
	ShowLabels  bool
	LabelHeight int
}

// DefaultOptions returns sensible defaults for contact sheet generation.
func DefaultOptions() Options {
	return Options{
		Columns:     4,
		ThumbSize:   200,
		Padding:     10,
		Background:  color.White,
		ShowLabels:  true,
		LabelHeight: 20,
	}
}

// Generate creates a contact sheet image from the given images and options.
func Generate(images []ImageEntry, opts Options) image.Image {
	if len(images) == 0 {
		return image.NewRGBA(image.Rect(0, 0, 1, 1))
	}

	cols := opts.Columns
	if cols <= 0 {
		cols = 4
	}
	rows := (len(images) + cols - 1) / cols

	labelH := 0
	if opts.ShowLabels {
		labelH = opts.LabelHeight
	}

	totalW := cols*(opts.ThumbSize+opts.Padding) + opts.Padding
	totalH := rows*(opts.ThumbSize+labelH+opts.Padding) + opts.Padding

	canvas := image.NewRGBA(image.Rect(0, 0, totalW, totalH))

	// Fill background
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{opts.Background}, image.Point{}, draw.Src)

	face := basicfont.Face7x13

	for i, entry := range images {
		col := i % cols
		row := i / cols

		cellX := opts.Padding + col*(opts.ThumbSize+opts.Padding)
		cellY := opts.Padding + row*(opts.ThumbSize+labelH+opts.Padding)

		// Resize image to fit within thumb size
		thumb := fitResize(entry.Image, opts.ThumbSize, opts.ThumbSize)
		thumbBounds := thumb.Bounds()
		tw := thumbBounds.Dx()
		th := thumbBounds.Dy()

		// Center thumbnail in cell
		offsetX := (opts.ThumbSize - tw) / 2
		offsetY := (opts.ThumbSize - th) / 2

		dstRect := image.Rect(
			cellX+offsetX,
			cellY+offsetY,
			cellX+offsetX+tw,
			cellY+offsetY+th,
		)
		draw.Draw(canvas, dstRect, thumb, thumbBounds.Min, draw.Over)

		// Draw label
		if opts.ShowLabels {
			label := entry.Label
			if label == "" {
				label = filepath.Base(entry.Path)
			}
			// Truncate label to fit cell width
			maxChars := opts.ThumbSize / 7 // Face7x13 is ~7px wide per char
			if len(label) > maxChars {
				label = label[:maxChars]
			}
			labelY := cellY + opts.ThumbSize + labelH - 4
			d := &font.Drawer{
				Dst:  canvas,
				Src:  image.NewUniform(color.Black),
				Face: face,
				Dot:  fixed.P(cellX, labelY),
			}
			d.DrawString(label)
		}
	}

	return canvas
}

// fitResize scales img to fit within maxW x maxH while preserving aspect ratio.
func fitResize(img image.Image, maxW, maxH int) image.Image {
	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()

	if srcW <= maxW && srcH <= maxH {
		return img
	}

	// Calculate scale factor to fit within bounds
	scaleW := float64(maxW) / float64(srcW)
	scaleH := float64(maxH) / float64(srcH)
	scale := scaleW
	if scaleH < scaleW {
		scale = scaleH
	}

	newW := int(float64(srcW) * scale)
	newH := int(float64(srcH) * scale)
	if newW < 1 {
		newW = 1
	}
	if newH < 1 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
	return dst
}
