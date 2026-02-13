package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/metadata"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/ssim"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Tool definitions

func convertImageTool() mcp.Tool {
	return mcp.NewTool("convert_image",
		mcp.WithDescription("Convert an image between formats with optional transforms (resize, crop, filters, watermark)"),
		mcp.WithString("input_path", mcp.Required(), mcp.Description("Absolute path to the input image file")),
		mcp.WithString("output_format", mcp.Required(), mcp.Description("Target format: jpeg, png, gif, webp, tiff, bmp, heic, avif, jxl")),
		mcp.WithString("output_path", mcp.Description("Output file path (default: input path with new extension)")),
		mcp.WithNumber("quality", mcp.Description("Encoding quality 1-100 (default: 92)"), mcp.Min(1), mcp.Max(100)),
		mcp.WithNumber("width", mcp.Description("Target width in pixels (preserves aspect ratio)")),
		mcp.WithNumber("height", mcp.Description("Target height in pixels (preserves aspect ratio)")),
		mcp.WithBoolean("grayscale", mcp.Description("Convert to grayscale")),
		mcp.WithBoolean("sharpen", mcp.Description("Apply sharpen filter")),
		mcp.WithNumber("blur", mcp.Description("Blur radius in pixels (0 = off)")),
		mcp.WithString("watermark", mcp.Description("Text watermark to overlay on the image")),
	)
}

func getFormatsTool() mcp.Tool {
	return mcp.NewTool("get_formats",
		mcp.WithDescription("List all supported image formats for decoding and encoding"),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
		}),
	)
}

func analyzeImageTool() mcp.Tool {
	return mcp.NewTool("analyze_image",
		mcp.WithDescription("Analyze an image file: detect format, dimensions, file size, and EXIF metadata"),
		mcp.WithString("path", mcp.Required(), mcp.Description("Absolute path to the image file")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
		}),
	)
}

func compareImagesTool() mcp.Tool {
	return mcp.NewTool("compare_images",
		mcp.WithDescription("Compare two images using Structural Similarity Index (SSIM). Returns a score from 0.0 (completely different) to 1.0 (identical)."),
		mcp.WithString("path1", mcp.Required(), mcp.Description("Absolute path to the first image")),
		mcp.WithString("path2", mcp.Required(), mcp.Description("Absolute path to the second image")),
		mcp.WithToolAnnotation(mcp.ToolAnnotation{
			ReadOnlyHint:    mcp.ToBoolPtr(true),
			DestructiveHint: mcp.ToBoolPtr(false),
		}),
	)
}

// Tool handlers

func (s *Server) handleConvertImage() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		inputPath, err := request.RequireString("input_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		outputFormatStr, err := request.RequireString("output_format")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		outputFormat, err := codec.ParseFormat(outputFormatStr)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("invalid output_format: %v", err)), nil
		}

		// Build output path
		outputPath := request.GetString("output_path", "")
		if outputPath == "" {
			ext := codec.DefaultExtension(outputFormat)
			base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
			outputPath = filepath.Join(filepath.Dir(inputPath), base+ext)
		}

		quality := request.GetInt("quality", 92)
		width := request.GetInt("width", 0)
		height := request.GetInt("height", 0)
		grayscale := request.GetBool("grayscale", false)
		sharpen := request.GetBool("sharpen", false)
		blur := request.GetFloat("blur", 0)
		watermark := request.GetString("watermark", "")

		job := pipeline.Job{
			InputPath:    inputPath,
			OutputPath:   outputPath,
			OutputFormat: outputFormat,
			Quality:      quality,
			Width:        width,
			Height:       height,
			Grayscale:    grayscale,
			Sharpen:      sharpen,
			Blur:         blur,
		}

		if watermark != "" {
			job.WatermarkText = watermark
			job.WatermarkPos = "bottom-right"
			job.WatermarkOpacity = 0.5
		}

		pipe := pipeline.NewPipeline(s.registry)
		start := time.Now()
		inputSize, outputSize, err := pipe.Execute(job)
		durationMs := time.Since(start).Milliseconds()

		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("conversion failed: %v", err)), nil
		}

		result := map[string]any{
			"output_path":  outputPath,
			"input_size":   inputSize,
			"output_size":  outputSize,
			"duration_ms":  durationMs,
			"input_format": "auto-detected",
			"output_format": string(outputFormat),
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (s *Server) handleGetFormats() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		decoders := s.registry.SupportedDecoders()
		encoders := s.registry.SupportedEncoders()

		decodeFormats := make([]string, len(decoders))
		for i, f := range decoders {
			decodeFormats[i] = string(f)
		}
		sort.Strings(decodeFormats)

		encodeFormats := make([]string, len(encoders))
		for i, f := range encoders {
			encodeFormats[i] = string(f)
		}
		sort.Strings(encodeFormats)

		result := map[string]any{
			"decode": decodeFormats,
			"encode": encodeFormats,
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (s *Server) handleAnalyzeImage() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		info, statErr := os.Stat(path)
		if statErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot access file: %v", statErr)), nil
		}

		f, err := os.Open(path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("open: %v", err)), nil
		}
		defer f.Close()

		format, err := codec.DetectFormat(f, path)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("detect format: %v", err)), nil
		}

		// Get dimensions using image.DecodeConfig for efficiency
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("seek: %v", seekErr)), nil
		}

		var width, height int
		dec, decErr := s.registry.Decoder(format)
		if decErr == nil {
			img, imgErr := dec.Decode(f)
			if imgErr == nil {
				bounds := img.Bounds()
				width = bounds.Dx()
				height = bounds.Dy()
			}
		}

		// Extract EXIF metadata
		if _, seekErr := f.Seek(0, 0); seekErr != nil {
			return mcp.NewToolResultError(fmt.Sprintf("seek: %v", seekErr)), nil
		}

		var hasExif bool
		var orientation int
		meta, metaErr := metadata.Extract(f, format)
		if metaErr == nil && meta.HasEXIF() {
			hasExif = true
			orientation = meta.Orientation()
		}

		result := map[string]any{
			"path":       path,
			"format":     string(format),
			"size_bytes": info.Size(),
			"has_exif":   hasExif,
		}

		if width > 0 && height > 0 {
			result["width"] = width
			result["height"] = height
		}

		if hasExif && orientation > 0 {
			result["exif_orientation"] = orientation
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func (s *Server) handleCompareImages() server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path1, err := request.RequireString("path1")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		path2, err := request.RequireString("path2")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		score, err := ssim.CompareFiles(path1, path2, s.registry)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("comparison failed: %v", err)), nil
		}

		result := map[string]any{
			"ssim_score":     score,
			"interpretation": ssim.Rating(score),
			"path1":          path1,
			"path2":          path2,
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}
