package transform

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	xdraw "golang.org/x/image/draw"
)

// WatermarkOptions controls watermark appearance and placement.
type WatermarkOptions struct {
	Text     string  // Text to render.
	Position string  // Placement: bottom-right (default), bottom-left, top-right, top-left, center.
	Opacity  float64 // 0-1, default 0.5.
	FontSize float64 // Scale factor: 1.0 = 13px (base), 2.0 = 26px, etc. Default 1.0.
	Color    string  // Hex color for text, e.g. "#FF0000" (default: white).
	BgColor  string  // Hex color for background, e.g. "#000000" (default: black).
}

// ApplyWatermark draws a semi-transparent text watermark on img and returns
// a new image. The text is rendered over a dark background pad for readability.
// When FontSize is not set (0), the watermark auto-scales to ~3% of the
// shorter image dimension so it remains visible on any image size.
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

	textColor := parseHexColor(opts.Color, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	bgColor := parseHexColor(opts.BgColor, color.RGBA{R: 0, G: 0, B: 0, A: 255})

	b := img.Bounds()
	imgW, imgH := b.Dx(), b.Dy()

	// Copy source into a mutable RGBA.
	dst := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)

	face := basicfont.Face7x13
	metrics := face.Metrics()
	textW := font.MeasureString(face, opts.Text).Ceil()
	textH := metrics.Ascent.Ceil() + metrics.Descent.Ceil()

	const padX = 4
	const padY = 2

	// Calculate scale factor.
	scale := opts.FontSize
	if scale <= 0 {
		// Auto-scale: target font height at ~3% of the shorter image dimension.
		minDim := float64(imgW)
		if float64(imgH) < minDim {
			minDim = float64(imgH)
		}
		scale = minDim * 0.03 / float64(textH)
		if scale < 1.0 {
			scale = 1.0
		}
	}

	// Render text at base size, then scale up with smooth interpolation.
	baseW := textW + padX*2
	baseH := textH + padY*2

	base := image.NewRGBA(image.Rect(0, 0, baseW, baseH))
	bgC := color.NRGBA{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: 255}
	draw.Draw(base, base.Bounds(), &image.Uniform{C: bgC}, image.Point{}, draw.Src)

	txtC := color.NRGBA{R: textColor.R, G: textColor.G, B: textColor.B, A: 255}
	dot := fixed.Point26_6{
		X: fixed.I(padX),
		Y: fixed.I(padY + metrics.Ascent.Ceil()),
	}
	drawer := &font.Drawer{
		Dst:  base,
		Src:  &image.Uniform{C: txtC},
		Face: face,
		Dot:  dot,
	}
	drawer.DrawString(opts.Text)

	// Scale with CatmullRom for smooth text edges.
	scaledW := int(float64(baseW) * scale)
	scaledH := int(float64(baseH) * scale)
	if scaledW < 1 {
		scaledW = 1
	}
	if scaledH < 1 {
		scaledH = 1
	}
	scaled := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	xdraw.CatmullRom.Scale(scaled, scaled.Bounds(), base, base.Bounds(), xdraw.Over, nil)

	// Scale margin proportionally.
	margin := int(10 * scale)
	if margin < 10 {
		margin = 10
	}

	// Position and composite with opacity.
	x, y := watermarkPosition(imgW, imgH, scaledW, scaledH, opts.Position, margin)
	opAlpha := opts.Opacity
	for sy := 0; sy < scaledH; sy++ {
		for sx := 0; sx < scaledW; sx++ {
			dx, dy := x+sx, y+sy
			if dx < 0 || dy < 0 || dx >= imgW || dy >= imgH {
				continue
			}
			sc := scaled.RGBAAt(sx, sy)
			dc := dst.RGBAAt(dx, dy)
			a := float64(sc.A) / 255 * opAlpha
			dst.SetRGBA(dx, dy, color.RGBA{
				R: uint8(float64(sc.R)*a + float64(dc.R)*(1-a)),
				G: uint8(float64(sc.G)*a + float64(dc.G)*(1-a)),
				B: uint8(float64(sc.B)*a + float64(dc.B)*(1-a)),
				A: dc.A,
			})
		}
	}
	return dst
}

// watermarkPosition returns the top-left corner for a watermark box of size
// (w, h) within an image of size (imgW, imgH), respecting the given margin.
func watermarkPosition(imgW, imgH, w, h int, position string, margin int) (int, int) {
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

// parseHexColor parses a hex color string like "#FF0000" or "FF0000".
// Returns the fallback color on parse failure.
func parseHexColor(s string, fallback color.RGBA) color.RGBA {
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return fallback
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return fallback
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return fallback
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return fallback
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// ParseHexColor is the exported version for testing.
func ParseHexColor(s string, fallback color.RGBA) color.RGBA {
	return parseHexColor(s, fallback)
}

// FormatHexColor formats a color as a hex string.
func FormatHexColor(c color.RGBA) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}
