package codec

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"
	"sort"
)

type gifDecoder struct{}
type gifEncoder struct{}

func (d *gifDecoder) Decode(r io.ReadSeeker) (image.Image, error) {
	return gif.Decode(r)
}

func (d *gifDecoder) Format() Format { return GIF }

// DecodeAll decodes all frames from an animated GIF.
func (d *gifDecoder) DecodeAll(r io.ReadSeeker) (*AnimatedImage, error) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		return nil, err
	}

	anim := &AnimatedImage{
		Frames:    make([]image.Image, len(g.Image)),
		Delays:    g.Delay,
		LoopCount: g.LoopCount,
	}

	if len(g.Disposal) > 0 {
		anim.Disposal = g.Disposal
	}

	// Convert paletted images to NRGBA for uniform processing
	for i, frame := range g.Image {
		bounds := frame.Bounds()
		nrgba := image.NewNRGBA(bounds)
		draw.Draw(nrgba, bounds, frame, bounds.Min, draw.Over)
		anim.Frames[i] = nrgba
	}

	return anim, nil
}

func (e *gifEncoder) Encode(w io.Writer, img image.Image, _ int) error {
	return gif.Encode(w, img, nil)
}

func (e *gifEncoder) Format() Format { return GIF }

// EncodeAll encodes multiple frames into an animated GIF.
func (e *gifEncoder) EncodeAll(w io.Writer, anim *AnimatedImage) error {
	g := &gif.GIF{
		Image:     make([]*image.Paletted, len(anim.Frames)),
		Delay:     anim.Delays,
		LoopCount: anim.LoopCount,
	}

	if len(anim.Disposal) > 0 {
		g.Disposal = anim.Disposal
	}

	for i, frame := range anim.Frames {
		bounds := frame.Bounds()
		palette := buildPalette(frame)
		palettedImg := image.NewPaletted(bounds, palette)
		draw.FloydSteinberg.Draw(palettedImg, bounds, frame, bounds.Min)
		g.Image[i] = palettedImg
	}

	// Ensure delays array matches frames
	for len(g.Delay) < len(g.Image) {
		g.Delay = append(g.Delay, 10) // default 100ms
	}

	return gif.EncodeAll(w, g)
}

// buildPalette creates a 256-color palette from an image using color frequency.
func buildPalette(img image.Image) color.Palette {
	bounds := img.Bounds()
	colorMap := make(map[color.RGBA]int)

	// Sample pixels (every Nth pixel for large images)
	step := 1
	totalPixels := bounds.Dx() * bounds.Dy()
	if totalPixels > 10000 {
		step = totalPixels / 10000
		if step < 1 {
			step = 1
		}
	}

	idx := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			idx++
			if idx%step != 0 {
				continue
			}
			r, g, b, a := img.At(x, y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			colorMap[c]++
		}
	}

	// Sort by frequency and pick top 255 + transparent
	type colorFreq struct {
		c     color.RGBA
		count int
	}
	sorted := make([]colorFreq, 0, len(colorMap))
	for c, count := range colorMap {
		sorted = append(sorted, colorFreq{c, count})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	palette := make(color.Palette, 0, 256)
	palette = append(palette, color.RGBA{0, 0, 0, 0}) // transparent
	for i := 0; i < len(sorted) && len(palette) < 256; i++ {
		palette = append(palette, sorted[i].c)
	}

	// Fill remaining slots if palette is too small
	for len(palette) < 2 {
		palette = append(palette, color.RGBA{0, 0, 0, 255})
	}

	return palette
}

func registerGIF(r *Registry) {
	r.RegisterDecoder(&gifDecoder{})
	r.RegisterEncoder(&gifEncoder{})
}
