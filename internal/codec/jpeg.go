package codec

import (
	"image"
	"image/jpeg"
	"io"
)

type jpegDecoder struct{}
type jpegEncoder struct{}

func (d *jpegDecoder) Decode(r io.ReadSeeker) (image.Image, error) {
	return jpeg.Decode(r)
}

func (d *jpegDecoder) Format() Format { return JPEG }

func (e *jpegEncoder) Encode(w io.Writer, img image.Image, quality int) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
}

func (e *jpegEncoder) EncodeWithOptions(w io.Writer, img image.Image, opts EncodeOptions) error {
	// Go's stdlib JPEG encoder does not support progressive or chroma subsampling.
	// These options are accepted for forward compatibility with alternative encoders.
	return jpeg.Encode(w, img, &jpeg.Options{Quality: opts.Quality})
}

func (e *jpegEncoder) Format() Format { return JPEG }

func registerJPEG(r *Registry) {
	r.RegisterDecoder(&jpegDecoder{})
	r.RegisterEncoder(&jpegEncoder{})
}
