package pipeline

import (
	"fmt"
	"image"
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

	// Extract metadata before decoding (for preservation or auto-rotate)
	var meta *metadata.Metadata
	needMeta := (job.PreserveMetadata && !job.StripMetadata) || job.AutoRotate
	if needMeta {
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

	// Populate EXIF orientation from extracted metadata
	if meta != nil {
		if orient := meta.Orientation(); orient > 0 {
			job.EXIFOrientation = orient
		}
	}

	// Don't inject metadata if we only extracted it for auto-rotate
	injectMeta := job.PreserveMetadata && !job.StripMetadata

	// Get decoder
	dec, err := p.Registry.Decoder(inputFormat)
	if err != nil {
		return inputSize, 0, err
	}

	// Get encoder
	enc, err := p.Registry.Encoder(job.OutputFormat)
	if err != nil {
		return inputSize, 0, err
	}

	// Check for multi-frame support
	mfDec, isMultiFrame := dec.(codec.MultiFrameDecoder)
	mfEnc, canEncodeMultiFrame := enc.(codec.MultiFrameEncoder)

	if isMultiFrame && canEncodeMultiFrame {
		// Multi-frame path
		anim, decErr := mfDec.DecodeAll(f)
		if decErr != nil {
			return inputSize, 0, fmt.Errorf("decode animated %s: %w", inputFormat, decErr)
		}

		if len(anim.Frames) > 1 {
			// Process each frame
			for i, frame := range anim.Frames {
				anim.Frames[i] = transformImage(frame, job)
			}

			// Create output file
			out, createErr := os.Create(job.OutputPath)
			if createErr != nil {
				return inputSize, 0, fmt.Errorf("create %s: %w", job.OutputPath, createErr)
			}
			defer out.Close()

			if encErr := mfEnc.EncodeAll(out, anim); encErr != nil {
				os.Remove(job.OutputPath)
				return inputSize, 0, fmt.Errorf("encode animated %s: %w", job.OutputFormat, encErr)
			}
			out.Close()

			// Get output file size
			outInfo, outStatErr := os.Stat(job.OutputPath)
			if outStatErr == nil {
				outputSize = outInfo.Size()
			}
			return inputSize, outputSize, nil
		}
		// Single frame animated image - fall through to normal path
		// Reset file position for normal decode
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return inputSize, 0, fmt.Errorf("seek: %w", seekErr)
		}
	}

	// Normal single-frame path
	img, err := dec.Decode(f)
	if err != nil {
		return inputSize, 0, fmt.Errorf("decode %s: %w", inputFormat, err)
	}

	img = transformImage(img, job)

	// Create output file
	out, err := os.Create(job.OutputPath)
	if err != nil {
		return inputSize, 0, fmt.Errorf("create %s: %w", job.OutputPath, err)
	}
	defer out.Close()

	// Encode (use AdvancedEncoder if available and options are set)
	opts := job.EncodeOpts
	useAdvanced := false
	if _, ok := enc.(codec.AdvancedEncoder); ok {
		if opts.Progressive || opts.Compression != 0 || opts.WebPMethod != 0 || opts.Lossless {
			useAdvanced = true
		}
	}
	if useAdvanced {
		if opts.Quality == 0 {
			opts.Quality = job.Quality
		}
		if err := enc.(codec.AdvancedEncoder).EncodeWithOptions(out, img, opts); err != nil {
			os.Remove(job.OutputPath)
			return inputSize, 0, fmt.Errorf("encode %s: %w", job.OutputFormat, err)
		}
	} else {
		if err := enc.Encode(out, img, job.Quality); err != nil {
			os.Remove(job.OutputPath)
			return inputSize, 0, fmt.Errorf("encode %s: %w", job.OutputFormat, err)
		}
	}

	// Close the file before metadata injection (needs to re-read/write)
	out.Close()

	// Inject metadata if available (and preservation was requested)
	if injectMeta && meta.HasEXIF() {
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

// transformImage applies all transforms to a single image frame.
func transformImage(img image.Image, job Job) image.Image {
	// Auto-rotate based on EXIF orientation
	if job.AutoRotate && job.EXIFOrientation > 1 {
		img = transform.AutoRotate(img, job.EXIFOrientation)
	}

	// Smart crop (entropy-based)
	if job.SmartCropWidth > 0 && job.SmartCropHeight > 0 {
		img = transform.SmartCrop(img, job.SmartCropWidth, job.SmartCropHeight)
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
			Width:         job.Width,
			Height:        job.Height,
			MaxDim:        job.MaxDim,
			Interpolation: job.Interpolation,
		})
	}

	// Watermark if requested
	if job.WatermarkText != "" {
		img = transform.ApplyWatermark(img, transform.WatermarkOptions{
			Text:     job.WatermarkText,
			Position: job.WatermarkPos,
			Opacity:  job.WatermarkOpacity,
			FontSize: job.WatermarkSize,
			Color:    job.WatermarkColor,
			BgColor:  job.WatermarkBg,
		})
	}

	// Apply filters (fixed order: brightness -> contrast -> sharpen -> blur -> grayscale -> sepia -> invert)
	if job.Brightness != 0 {
		img = transform.Brightness(img, job.Brightness)
	}
	if job.Contrast != 0 {
		img = transform.Contrast(img, job.Contrast)
	}
	if job.Sharpen {
		img = transform.Sharpen(img)
	}
	if job.Blur > 0 {
		img = transform.Blur(img, job.Blur)
	}
	if job.Grayscale {
		img = transform.Grayscale(img)
	}
	if job.Sepia > 0 {
		img = transform.Sepia(img, job.Sepia)
	}
	if job.Invert {
		img = transform.Invert(img)
	}

	return img
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
