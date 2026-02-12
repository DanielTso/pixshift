package codec

import (
	"image"
	"image/png"
	"io"
)

type pngDecoder struct{}
type pngEncoder struct{}

func (d *pngDecoder) Decode(r io.ReadSeeker) (image.Image, error) {
	return png.Decode(r)
}

func (d *pngDecoder) Format() Format { return PNG }

func (e *pngEncoder) Encode(w io.Writer, img image.Image, _ int) error {
	return png.Encode(w, img)
}

func (e *pngEncoder) EncodeWithOptions(w io.Writer, img image.Image, opts EncodeOptions) error {
	enc := &png.Encoder{}
	switch opts.Compression {
	case 1:
		enc.CompressionLevel = png.NoCompression
	case 2:
		enc.CompressionLevel = png.BestSpeed
	case 3:
		enc.CompressionLevel = png.BestCompression
	default:
		enc.CompressionLevel = png.DefaultCompression
	}
	return enc.Encode(w, img)
}

func (e *pngEncoder) Format() Format { return PNG }

func registerPNG(r *Registry) {
	r.RegisterDecoder(&pngDecoder{})
	r.RegisterEncoder(&pngEncoder{})
}
