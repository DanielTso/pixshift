package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
	pixcolor "github.com/DanielTso/pixshift/internal/color"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/ssim"
)

// Format type aliases for the public API.
type Format = codec.Format

// Format constants.
const (
	JPEG Format = codec.JPEG
	PNG  Format = codec.PNG
	WebP Format = codec.WebP
	AVIF Format = codec.AVIF
	HEIC Format = codec.HEIC
	TIFF Format = codec.TIFF
	GIF  Format = codec.GIF
	JXL  Format = codec.JXL
)

// Color represents a dominant color.
type Color = pixcolor.Color

// ImageInfo holds metadata about an image file.
type ImageInfo struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Format Format `json:"format"`
	Size   int64  `json:"size"`
}

// Convert converts an image file to the specified format and writes the result.
func Convert(input, output string, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	reg := codec.DefaultRegistry()

	// Detect output format from extension if not set.
	outputFormat := cfg.format
	if outputFormat == "" {
		ext := strings.TrimPrefix(filepath.Ext(output), ".")
		f, err := codec.ParseFormat(ext)
		if err != nil {
			return fmt.Errorf("cannot determine output format from extension %q: %w", ext, err)
		}
		outputFormat = f
	}

	job := pipeline.Job{
		InputPath:        input,
		OutputPath:       output,
		OutputFormat:     outputFormat,
		Quality:          cfg.quality,
		Width:            cfg.width,
		Height:           cfg.height,
		MaxDim:           cfg.maxDim,
		Grayscale:        cfg.grayscale,
		Sharpen:          cfg.sharpen,
		Blur:             cfg.blur,
		Invert:           cfg.invert,
		StripMetadata:    cfg.stripMetadata,
		PreserveMetadata: cfg.preserveMetadata,
		WatermarkText:    cfg.watermarkText,
		WatermarkPos:     cfg.watermarkPos,
		WatermarkOpacity: cfg.watermarkOpacity,
		SmartCropWidth:   cfg.smartCropW,
		SmartCropHeight:  cfg.smartCropH,
	}

	pipe := pipeline.NewPipeline(reg)
	_, _, err := pipe.Execute(job)
	return err
}

// ConvertBytes converts image bytes to the specified format.
func ConvertBytes(data []byte, outputFormat Format, opts ...Option) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "pixshift-sdk-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input")
	if err := os.WriteFile(inputPath, data, 0600); err != nil {
		return nil, fmt.Errorf("write input: %w", err)
	}

	outExt := codec.DefaultExtension(outputFormat)
	outputPath := filepath.Join(tmpDir, "output"+outExt)

	opts = append([]Option{WithFormat(outputFormat)}, opts...)
	if err := Convert(inputPath, outputPath, opts...); err != nil {
		return nil, err
	}

	return os.ReadFile(outputPath)
}

// Analyze returns metadata about an image file.
func Analyze(path string) (*ImageInfo, error) {
	reg := codec.DefaultRegistry()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	format, err := codec.DetectFormat(f, path)
	if err != nil {
		return nil, fmt.Errorf("detect format: %w", err)
	}

	if _, err := f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}

	dec, err := reg.Decoder(format)
	if err != nil {
		return nil, fmt.Errorf("no decoder for %s: %w", format, err)
	}

	img, err := dec.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	bounds := img.Bounds()
	return &ImageInfo{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: format,
		Size:   info.Size(),
	}, nil
}

// Palette extracts the N most dominant colors from an image.
func Palette(path string, count int) ([]Color, error) {
	reg := codec.DefaultRegistry()
	return pixcolor.ExtractPaletteFromFile(path, count, reg)
}

// Compare computes the Structural Similarity Index between two images.
// Returns a score between 0 and 1, where 1 means identical.
func Compare(path1, path2 string) (float64, error) {
	reg := codec.DefaultRegistry()
	return ssim.CompareFiles(path1, path2, reg)
}
