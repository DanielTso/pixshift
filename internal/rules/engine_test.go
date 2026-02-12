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
