package rules

import (
	"testing"

	"github.com/DanielTso/pixshift/internal/codec"
)

func TestParseRules_Valid(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Name: "HEIC to JPEG", Format: "heic", Output: "jpeg", Quality: 90},
			{Name: "PNG to WebP", Format: "png", Output: "webp"},
		},
	}

	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	if len(parsed) != 2 {
		t.Fatalf("got %d rules, want 2", len(parsed))
	}

	if parsed[0].InputFormat != codec.HEIC {
		t.Errorf("rule 0 InputFormat = %q, want %q", parsed[0].InputFormat, codec.HEIC)
	}
	if parsed[0].OutputFormat != codec.JPEG {
		t.Errorf("rule 0 OutputFormat = %q, want %q", parsed[0].OutputFormat, codec.JPEG)
	}
	if parsed[0].Rule.Quality != 90 {
		t.Errorf("rule 0 Quality = %d, want 90", parsed[0].Rule.Quality)
	}

	if parsed[1].InputFormat != codec.PNG {
		t.Errorf("rule 1 InputFormat = %q, want %q", parsed[1].InputFormat, codec.PNG)
	}
	if parsed[1].OutputFormat != codec.WebP {
		t.Errorf("rule 1 OutputFormat = %q, want %q", parsed[1].OutputFormat, codec.WebP)
	}
}

func TestParseRules_InvalidOutputFormat(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Format: "heic", Output: "invalidformat"},
		},
	}

	_, err := ParseRules(cfg)
	if err == nil {
		t.Error("expected error for invalid output format, got nil")
	}
}

func TestParseRules_MissingOutput(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Format: "heic", Output: ""},
		},
	}

	_, err := ParseRules(cfg)
	if err == nil {
		t.Error("expected error for missing output, got nil")
	}
}

func TestParseRules_InvalidInputFormat(t *testing.T) {
	cfg := &Config{
		Rules: []Rule{
			{Format: "badformat", Output: "jpeg"},
		},
	}

	_, err := ParseRules(cfg)
	if err == nil {
		t.Error("expected error for invalid input format, got nil")
	}
}

func TestParseRules_NoFormat(t *testing.T) {
	// Rules without a Format field are valid (match any input format)
	cfg := &Config{
		Rules: []Rule{
			{Glob: "*.heic", Output: "jpeg"},
		},
	}

	parsed, err := ParseRules(cfg)
	if err != nil {
		t.Fatalf("ParseRules: %v", err)
	}
	if parsed[0].InputFormat != "" {
		t.Errorf("expected empty InputFormat, got %q", parsed[0].InputFormat)
	}
}
