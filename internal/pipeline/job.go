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
	BackupOriginal   bool
}

// Result holds the outcome of a conversion job.
type Result struct {
	Job        Job
	Error      error
	InputSize  int64
	OutputSize int64
}
