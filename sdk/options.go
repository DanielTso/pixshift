package sdk

type config struct {
	format           Format
	quality          int
	width, height    int
	maxDim           int
	grayscale        bool
	sharpen          bool
	blur             float64
	invert           bool
	stripMetadata    bool
	preserveMetadata bool
	watermarkText    string
	watermarkPos     string
	watermarkOpacity float64
	smartCropW       int
	smartCropH       int
}

func defaultConfig() config {
	return config{
		quality:          92,
		watermarkOpacity: 0.5,
	}
}

// Option configures a conversion operation.
type Option func(*config)

// WithFormat sets the output format explicitly, overriding extension-based detection.
func WithFormat(f Format) Option { return func(c *config) { c.format = f } }

// WithQuality sets the encoding quality (1-100).
func WithQuality(q int) Option { return func(c *config) { c.quality = q } }

// WithResize sets the target width and height. Zero means auto-scale that dimension.
func WithResize(w, h int) Option { return func(c *config) { c.width = w; c.height = h } }

// WithMaxDim constrains the largest dimension while preserving aspect ratio.
func WithMaxDim(d int) Option { return func(c *config) { c.maxDim = d } }

// WithGrayscale converts the image to grayscale.
func WithGrayscale() Option { return func(c *config) { c.grayscale = true } }

// WithSharpen applies a sharpening filter.
func WithSharpen() Option { return func(c *config) { c.sharpen = true } }

// WithBlur applies a Gaussian blur with the given radius in pixels.
func WithBlur(r float64) Option { return func(c *config) { c.blur = r } }

// WithInvert inverts the image colors.
func WithInvert() Option { return func(c *config) { c.invert = true } }

// WithStripMetadata removes all EXIF metadata from the output.
func WithStripMetadata() Option { return func(c *config) { c.stripMetadata = true } }

// WithPreserveMetadata copies EXIF metadata from the input to the output.
func WithPreserveMetadata() Option { return func(c *config) { c.preserveMetadata = true } }

// WithSmartCrop crops to the given dimensions using entropy-based region selection.
func WithSmartCrop(w, h int) Option {
	return func(c *config) { c.smartCropW = w; c.smartCropH = h }
}

// WithWatermark overlays text on the image at the given position with the given opacity.
func WithWatermark(text, pos string, opacity float64) Option {
	return func(c *config) {
		c.watermarkText = text
		c.watermarkPos = pos
		c.watermarkOpacity = opacity
	}
}
