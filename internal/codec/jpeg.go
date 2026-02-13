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
	return e.EncodeWithOptions(w, img, EncodeOptions{Quality: quality})
}

// EncodeWithOptions encodes a JPEG image with optional encoding parameters.
//
// Progressive JPEG encoding is not currently supported. The opts.Progressive
// flag is accepted but silently falls back to baseline (sequential) JPEG.
//
// Reason: progressive encoding requires libjpeg-turbo via CGO, but on systems
// where other CGO dependencies (libheif, libavif) link against a different
// libjpeg-turbo ABI version (e.g. headers say version 80 but runtime library
// is version 62), the resulting struct layout mismatch causes a fatal crash:
// "Wrong JPEG library version: library is 62, caller expects 80". This is an
// ABI incompatibility that cannot be resolved without rebuilding the system
// libraries. The stdlib encoder produces correct, well-tested baseline JPEGs.
func (e *jpegEncoder) EncodeWithOptions(w io.Writer, img image.Image, opts EncodeOptions) error {
	return jpeg.Encode(w, img, &jpeg.Options{Quality: opts.Quality})
}

func (e *jpegEncoder) Format() Format { return JPEG }

func registerJPEG(r *Registry) {
	r.RegisterDecoder(&jpegDecoder{})
	r.RegisterEncoder(&jpegEncoder{})
}
