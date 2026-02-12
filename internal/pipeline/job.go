package pipeline

import "github.com/DanielTso/pixshift/internal/codec"

// Job describes a single conversion task.
type Job struct {
	InputPath        string
	OutputPath       string
	InputFormat      codec.Format
	OutputFormat     codec.Format
	Quality          int
	PreserveMetadata bool
	StripMetadata    bool
	Width            int
	Height           int
	MaxDim           int

	// v0.3.0 fields
	AutoRotate       bool
	EXIFOrientation  int
	CropWidth        int
	CropHeight       int
	CropAspectRatio  string
	CropGravity      string
	WatermarkText    string
	WatermarkPos     string
	WatermarkOpacity float64
	WatermarkSize    float64
	WatermarkColor   string
	WatermarkBg      string
	BackupOriginal   bool

	// v0.4.0 fields
	Grayscale     bool
	Sepia         float64 // 0=off, 0.0-1.0 intensity
	Brightness    float64 // -100 to +100
	Contrast      float64 // -100 to +100
	Sharpen       bool
	Blur          float64 // 0=off, radius in pixels
	Invert        bool
	Interpolation string // "nearest", "bilinear", "catmullrom" (default)
	EncodeOpts    codec.EncodeOptions
}

// Result holds the outcome of a conversion job.
type Result struct {
	Job        Job
	Error      error
	InputSize  int64
	OutputSize int64
}
