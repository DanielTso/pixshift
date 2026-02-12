package transform

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// WatermarkOptions controls watermark appearance and placement.
type WatermarkOptions struct {
	Text     string  // Text to render.
	Position string  // Placement: bottom-right (default), bottom-left, top-right, top-left, center.
	Opacity  float64 // 0-1, default 0.5.
	FontSize float64 // Unused with basicfont; reserved for future scalable fonts.
}

// ApplyWatermark draws a semi-transparent text watermark on img and returns
// a new image. The text is rendered in white over a dark background pad for
// readability.
func ApplyWatermark(img image.Image, opts WatermarkOptions) image.Image {
	if opts.Text == "" {
		return img
	}
	if opts.Opacity <= 0 {
		opts.Opacity = 0.5
	}
	if opts.Opacity > 1 {
		opts.Opacity = 1
	}

	b := img.Bounds()

	// Copy source into a mutable RGBA.
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)

	face := basicfont.Face7x13
	metrics := face.Metrics()
	textW := font.MeasureString(face, opts.Text).Ceil()
	textH := metrics.Ascent.Ceil() + metrics.Descent.Ceil()

	const padX = 4
	const padY = 2

	bgW := textW + padX*2
	bgH := textH + padY*2

	x, y := watermarkPosition(dst.Bounds().Dx(), dst.Bounds().Dy(), bgW, bgH, opts.Position)

	// Draw dark background rectangle with opacity.
	alpha := uint8(opts.Opacity * 255)
	bgColor := color.NRGBA{R: 0, G: 0, B: 0, A: alpha}
	bgRect := image.Rect(x, y, x+bgW, y+bgH)
	draw.Draw(dst, bgRect, &image.Uniform{C: bgColor}, image.Point{}, draw.Over)

	// Draw white text with opacity.
	textColor := color.NRGBA{R: 255, G: 255, B: 255, A: alpha}
	dot := fixed.Point26_6{
		X: fixed.I(x + padX),
		Y: fixed.I(y + padY + metrics.Ascent.Ceil()),
	}
	drawer := &font.Drawer{
		Dst:  dst,
		Src:  &image.Uniform{C: textColor},
		Face: face,
		Dot:  dot,
	}
	drawer.DrawString(opts.Text)

	return dst
}

// watermarkPosition returns the top-left corner for a watermark box of size
// (w, h) within an image of size (imgW, imgH), respecting a 10px margin.
func watermarkPosition(imgW, imgH, w, h int, position string) (int, int) {
	const margin = 10
	switch strings.ToLower(strings.TrimSpace(position)) {
	case "top-left":
		return margin, margin
	case "top-right":
		return imgW - w - margin, margin
	case "bottom-left":
		return margin, imgH - h - margin
	case "center":
		return (imgW - w) / 2, (imgH - h) / 2
	default: // "bottom-right" or empty
		return imgW - w - margin, imgH - h - margin
	}
}
