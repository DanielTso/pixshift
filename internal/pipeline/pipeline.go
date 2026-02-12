package pipeline

import (
	"fmt"
	"io"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/metadata"
	"github.com/DanielTso/pixshift/internal/resize"
	"github.com/DanielTso/pixshift/internal/transform"
)

// Pipeline executes the detect -> decode -> transform -> encode -> metadata inject flow.
type Pipeline struct {
	Registry *codec.Registry
}

// NewPipeline creates a pipeline with the given codec registry.
func NewPipeline(reg *codec.Registry) *Pipeline {
	return &Pipeline{Registry: reg}
}

// Execute runs a single conversion job and returns file sizes.
func (p *Pipeline) Execute(job Job) (inputSize, outputSize int64, err error) {
	// Get input file size
	info, statErr := os.Stat(job.InputPath)
	if statErr == nil {
		inputSize = info.Size()
	}

	// Backup original if requested
	if job.BackupOriginal {
		backupPath := job.InputPath + ".bak"
		if err := copyFile(job.InputPath, backupPath); err != nil {
			return inputSize, 0, fmt.Errorf("backup %s: %w", job.InputPath, err)
		}
	}

	// Open input file
	f, err := os.Open(job.InputPath)
	if err != nil {
		return inputSize, 0, fmt.Errorf("open %s: %w", job.InputPath, err)
	}
	defer f.Close()

	// Detect input format if not set
	inputFormat := job.InputFormat
	if inputFormat == "" {
		inputFormat, err = codec.DetectFormat(f, job.InputPath)
		if err != nil {
			return inputSize, 0, fmt.Errorf("detect format: %w", err)
		}
	}

	// Extract metadata before decoding (if requested and not stripping)
	var meta *metadata.Metadata
	if job.PreserveMetadata && !job.StripMetadata {
		meta, err = metadata.Extract(f, inputFormat)
		if err != nil {
			// Non-fatal: warn but continue without metadata
			meta = nil
		}
		// Reset file position after metadata extraction
		if _, err := f.Seek(0, 0); err != nil {
			return inputSize, 0, fmt.Errorf("seek: %w", err)
		}
	}

	// Get decoder
	dec, err := p.Registry.Decoder(inputFormat)
	if err != nil {
		return inputSize, 0, err
	}

	// Decode
	img, err := dec.Decode(f)
	if err != nil {
		return inputSize, 0, fmt.Errorf("decode %s: %w", inputFormat, err)
	}

	// Auto-rotate based on EXIF orientation
	if job.AutoRotate && job.EXIFOrientation > 1 {
		img = transform.AutoRotate(img, job.EXIFOrientation)
	}

	// Crop if requested
	if job.CropWidth > 0 || job.CropHeight > 0 || job.CropAspectRatio != "" {
		img = transform.Crop(img, transform.CropOptions{
			Width:       job.CropWidth,
			Height:      job.CropHeight,
			AspectRatio: job.CropAspectRatio,
			Gravity:     job.CropGravity,
		})
	}

	// Resize if requested
	if job.Width > 0 || job.Height > 0 || job.MaxDim > 0 {
		img = resize.Resize(img, resize.ResizeOptions{
			Width:  job.Width,
			Height: job.Height,
			MaxDim: job.MaxDim,
		})
	}

	// Watermark if requested
	if job.WatermarkText != "" {
		img = transform.ApplyWatermark(img, transform.WatermarkOptions{
			Text:     job.WatermarkText,
			Position: job.WatermarkPos,
			Opacity:  job.WatermarkOpacity,
		})
	}

	// Get encoder
	enc, err := p.Registry.Encoder(job.OutputFormat)
	if err != nil {
		return inputSize, 0, err
	}

	// Create output file
	out, err := os.Create(job.OutputPath)
	if err != nil {
		return inputSize, 0, fmt.Errorf("create %s: %w", job.OutputPath, err)
	}
	defer out.Close()

	// Encode
	if err := enc.Encode(out, img, job.Quality); err != nil {
		os.Remove(job.OutputPath)
		return inputSize, 0, fmt.Errorf("encode %s: %w", job.OutputFormat, err)
	}

	// Close the file before metadata injection (needs to re-read/write)
	out.Close()

	// Inject metadata if available (and not stripping)
	if !job.StripMetadata && meta.HasEXIF() {
		if err := metadata.Inject(job.OutputPath, job.OutputFormat, meta); err != nil {
			// Non-fatal: file was converted but metadata not preserved
			return inputSize, 0, fmt.Errorf("metadata inject (file converted OK): %w", err)
		}
	}

	// Get output file size
	outInfo, statErr := os.Stat(job.OutputPath)
	if statErr == nil {
		outputSize = outInfo.Size()
	}

	return inputSize, outputSize, nil
}

// copyFile copies src to dst.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
