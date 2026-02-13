package codec

import (
	"image"
	"io"
	"strings"
)

// Format represents a supported image format.
type Format string

const (
	JPEG Format = "jpeg"
	PNG  Format = "png"
	GIF  Format = "gif"
	WebP Format = "webp"
	TIFF Format = "tiff"
	BMP  Format = "bmp"
	HEIC Format = "heic"
	AVIF Format = "avif"
	CR2  Format = "cr2"
	NEF  Format = "nef"
	DNG  Format = "dng"
	JXL  Format = "jxl"
	ARW  Format = "arw"  // Sony
	RAF  Format = "raf"  // Fujifilm
	ORF  Format = "orf"  // Olympus
	RW2  Format = "rw2"  // Panasonic
)

// Decoder can decode an image from a reader.
type Decoder interface {
	Decode(r io.ReadSeeker) (image.Image, error)
	Format() Format
}

// Encoder can encode an image to a writer.
type Encoder interface {
	Encode(w io.Writer, img image.Image, quality int) error
	Format() Format
}

// EncodeOptions holds format-specific encoding parameters.
type EncodeOptions struct {
	Quality     int
	Progressive bool   // JPEG: progressive encoding (reserved for future encoder)
	Subsample   string // JPEG: chroma subsampling "444", "422", "420" (reserved)
	Compression int    // PNG: 0=default, 1=none, 2=fast, 3=best
	WebPMethod  int    // WebP: encoding method 0-6 (speed vs quality)
	Lossless    bool   // WebP: lossless mode
}

// AdvancedEncoder extends Encoder with format-specific encoding options.
type AdvancedEncoder interface {
	Encoder
	EncodeWithOptions(w io.Writer, img image.Image, opts EncodeOptions) error
}

// AnimatedImage holds a multi-frame animation.
type AnimatedImage struct {
	Frames    []image.Image
	Delays    []int // centiseconds per frame
	Disposal  []byte
	LoopCount int
}

// MultiFrameDecoder can decode all frames of an animated image.
type MultiFrameDecoder interface {
	Decoder
	DecodeAll(r io.ReadSeeker) (*AnimatedImage, error)
}

// MultiFrameEncoder can encode multiple frames into an animated image.
type MultiFrameEncoder interface {
	Encoder
	EncodeAll(w io.Writer, anim *AnimatedImage) error
}

// IsRAW returns true if the format is a RAW camera format.
func IsRAW(f Format) bool {
	switch f {
	case CR2, NEF, DNG, ARW, RAF, ORF, RW2:
		return true
	}
	return false
}

// DefaultExtension returns the primary file extension for a format.
func DefaultExtension(f Format) string {
	switch f {
	case JPEG:
		return ".jpg"
	case PNG:
		return ".png"
	case GIF:
		return ".gif"
	case WebP:
		return ".webp"
	case TIFF:
		return ".tiff"
	case BMP:
		return ".bmp"
	case HEIC:
		return ".heic"
	case AVIF:
		return ".avif"
	case CR2:
		return ".cr2"
	case NEF:
		return ".nef"
	case DNG:
		return ".dng"
	case JXL:
		return ".jxl"
	case ARW:
		return ".arw"
	case RAF:
		return ".raf"
	case ORF:
		return ".orf"
	case RW2:
		return ".rw2"
	}
	return "." + string(f)
}

// IsSupportedExtension checks if a file extension belongs to any known format.
func IsSupportedExtension(ext string) bool {
	ext = strings.ToLower(ext)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".tif",
		".bmp", ".heic", ".heif", ".avif", ".cr2", ".nef", ".dng",
		".jxl", ".arw", ".raf", ".orf", ".rw2":
		return true
	}
	return false
}
