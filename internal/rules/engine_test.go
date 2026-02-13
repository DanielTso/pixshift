package rules

import (
	"testing"

	"github.com/DanielTso/pixshift/internal/codec"
)

func TestEngine_FormatMatch(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Format: "heic", Output: "jpeg", Quality: 90},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/image.heic", codec.HEIC)
	if job == nil {
		t.Fatal("expected match for HEIC input, got nil")
	}
	if job.OutputFormat != codec.JPEG {
		t.Errorf("OutputFormat = %q, want %q", job.OutputFormat, codec.JPEG)
	}
}

func TestEngine_GlobMatch(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Glob: "*.heic", Output: "jpeg"},
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/photo.heic", codec.HEIC)
	if job == nil {
		t.Fatal("expected glob match for *.heic, got nil")
	}
	if job.OutputFormat != codec.JPEG {
		t.Errorf("OutputFormat = %q, want %q", job.OutputFormat, codec.JPEG)
	}
}

func TestEngine_NoMatch(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Format: "heic", Output: "jpeg"},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/image.png", codec.PNG)
	if job != nil {
		t.Errorf("expected no match for PNG input, got %+v", job)
	}
}

func TestEngine_FirstMatchWins(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Name: "first", Format: "heic", Output: "webp"},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.WebP,
		},
		{
			Rule:         Rule{Name: "second", Format: "heic", Output: "jpeg"},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/image.heic", codec.HEIC)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if job.OutputFormat != codec.WebP {
		t.Errorf("first-match-wins: OutputFormat = %q, want %q (first rule)", job.OutputFormat, codec.WebP)
	}
}

func TestEngine_QualityFallback(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Format: "heic", Output: "jpeg"},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)
	// Default quality is 92 from the engine

	job := engine.Match("/photos/image.heic", codec.HEIC)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if job.Quality != 92 {
		t.Errorf("Quality = %d, want 92 (engine default)", job.Quality)
	}
}

func TestEngine_RuleQualityOverridesDefault(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Format: "heic", Output: "jpeg", Quality: 75},
			InputFormat:  codec.HEIC,
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/image.heic", codec.HEIC)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if job.Quality != 75 {
		t.Errorf("Quality = %d, want 75 (rule override)", job.Quality)
	}
}

func TestEngine_GlobNoMatch(t *testing.T) {
	rules := []ParsedRule{
		{
			Rule:         Rule{Glob: "*.heic", Output: "jpeg"},
			OutputFormat: codec.JPEG,
		},
	}
	engine := NewEngine(rules)

	job := engine.Match("/photos/photo.png", codec.PNG)
	if job != nil {
		t.Errorf("expected no glob match for .png file, got %+v", job)
	}
}

func TestMatch_WithTransforms(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{
				Format:      "jpeg",
				Output:      "webp",
				Quality:     85,
				Grayscale:   true,
				Sharpen:     true,
				MaxDim:      1920,
				CropRatio:   "16:9",
				CropGravity: "center",
			},
		},
	}
	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	engine := NewEngine(parsed)
	job := engine.Match("test.jpg", codec.JPEG)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if !job.Grayscale {
		t.Error("Grayscale should be true")
	}
	if !job.Sharpen {
		t.Error("Sharpen should be true")
	}
	if job.MaxDim != 1920 {
		t.Errorf("MaxDim = %d, want 1920", job.MaxDim)
	}
	if job.CropAspectRatio != "16:9" {
		t.Errorf("CropAspectRatio = %q, want 16:9", job.CropAspectRatio)
	}
	if job.CropGravity != "center" {
		t.Errorf("CropGravity = %q, want center", job.CropGravity)
	}
	if job.Quality != 85 {
		t.Errorf("Quality = %d, want 85", job.Quality)
	}
}

func TestMatch_WithFilters(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{
				Format:     "jpeg",
				Output:     "png",
				Invert:     true,
				Sepia:      0.5,
				Brightness: 10,
				Contrast:   -5,
				Blur:       2.5,
			},
		},
	}
	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	engine := NewEngine(parsed)
	job := engine.Match("photo.jpg", codec.JPEG)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if !job.Invert {
		t.Error("Invert should be true")
	}
	if job.Sepia != 0.5 {
		t.Errorf("Sepia = %f, want 0.5", job.Sepia)
	}
	if job.Brightness != 10 {
		t.Errorf("Brightness = %f, want 10", job.Brightness)
	}
	if job.Contrast != -5 {
		t.Errorf("Contrast = %f, want -5", job.Contrast)
	}
	if job.Blur != 2.5 {
		t.Errorf("Blur = %f, want 2.5", job.Blur)
	}
}

func TestMatch_WithEncodingOptions(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{
				Format:         "jpeg",
				Output:         "webp",
				Quality:        90,
				Lossless:       true,
				WebpMethod:     4,
				Interpolation:  "bilinear",
				StripMetadata:  true,
			},
		},
	}
	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	engine := NewEngine(parsed)
	job := engine.Match("img.jpg", codec.JPEG)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if !job.EncodeOpts.Lossless {
		t.Error("EncodeOpts.Lossless should be true")
	}
	if job.EncodeOpts.WebPMethod != 4 {
		t.Errorf("EncodeOpts.WebPMethod = %d, want 4", job.EncodeOpts.WebPMethod)
	}
	if job.Interpolation != "bilinear" {
		t.Errorf("Interpolation = %q, want bilinear", job.Interpolation)
	}
	if !job.StripMetadata {
		t.Error("StripMetadata should be true")
	}
}

func TestMatch_WithWatermark(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{
				Format:           "jpeg",
				Output:           "webp",
				WatermarkText:    "Copyright 2026",
				WatermarkPos:     "bottom-right",
				WatermarkOpacity: 0.7,
				WatermarkSize:    2.0,
				WatermarkColor:   "#FF0000",
				WatermarkBg:      "#000000",
			},
		},
	}
	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	engine := NewEngine(parsed)
	job := engine.Match("photo.jpg", codec.JPEG)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if job.WatermarkText != "Copyright 2026" {
		t.Errorf("WatermarkText = %q, want 'Copyright 2026'", job.WatermarkText)
	}
	if job.WatermarkPos != "bottom-right" {
		t.Errorf("WatermarkPos = %q, want bottom-right", job.WatermarkPos)
	}
	if job.WatermarkOpacity != 0.7 {
		t.Errorf("WatermarkOpacity = %f, want 0.7", job.WatermarkOpacity)
	}
	if job.WatermarkSize != 2.0 {
		t.Errorf("WatermarkSize = %f, want 2.0", job.WatermarkSize)
	}
	if job.WatermarkColor != "#FF0000" {
		t.Errorf("WatermarkColor = %q, want #FF0000", job.WatermarkColor)
	}
	if job.WatermarkBg != "#000000" {
		t.Errorf("WatermarkBg = %q, want #000000", job.WatermarkBg)
	}
}

func TestMatch_PreserveMetadataFromRule(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{
				Format:           "jpeg",
				Output:           "png",
				PreserveMetadata: true,
			},
		},
	}
	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	engine := NewEngine(parsed)
	// Engine.Metadata is false, but rule has PreserveMetadata=true
	engine.Metadata = false
	job := engine.Match("photo.jpg", codec.JPEG)
	if job == nil {
		t.Fatal("expected match, got nil")
	}
	if !job.PreserveMetadata {
		t.Error("PreserveMetadata should be true from rule")
	}
}
