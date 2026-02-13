package completion

import (
	"strings"
	"testing"
)

func TestGenerateBash_ContainsFlags(t *testing.T) {
	output := GenerateBash()

	flags := []string{
		"--format",
		"--quality",
		"--grayscale",
		"--progressive",
		"--lossless",
		"--webp-method",
		"--png-compression",
		"--sepia",
		"--brightness",
		"--contrast",
		"--sharpen",
		"--blur",
		"--invert",
		"--watermark",
		"--watermark-pos",
		"--watermark-size",
		"--watermark-color",
		"--watermark-bg",
		"--interpolation",
		"--crop",
		"--crop-ratio",
		"--crop-gravity",
		"--preset",
		"--auto-rotate",
		"--backup",
		"--json",
		"--tree",
		"--dedup",
		"--ssim",
		"--contact-sheet",
	}

	for _, flag := range flags {
		if !strings.Contains(output, flag) {
			t.Errorf("bash completion missing flag %q", flag)
		}
	}
}

func TestGenerateBash_ContainsFormats(t *testing.T) {
	output := GenerateBash()

	formats := []string{"jpg", "png", "gif", "webp", "heic", "avif"}
	for _, fmt := range formats {
		if !strings.Contains(output, fmt) {
			t.Errorf("bash completion missing format %q", fmt)
		}
	}
}

func TestGenerateBash_ContainsPresets(t *testing.T) {
	output := GenerateBash()

	presets := []string{"web", "thumbnail", "print", "archive"}
	for _, p := range presets {
		if !strings.Contains(output, p) {
			t.Errorf("bash completion missing preset %q", p)
		}
	}
}

func TestGenerateBash_ContainsShortFlags(t *testing.T) {
	output := GenerateBash()

	shortFlags := []string{"-f", "-q", "-j", "-o", "-r", "-m", "-w", "-c", "-v", "-V", "-h", "-s"}
	for _, flag := range shortFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("bash completion missing short flag %q", flag)
		}
	}
}

func TestGenerateZsh_ContainsFlags(t *testing.T) {
	output := GenerateZsh()

	flags := []string{
		"--format",
		"--quality",
		"--grayscale",
		"--progressive",
		"--lossless",
		"--webp-method",
		"--png-compression",
		"--sepia",
		"--brightness",
		"--contrast",
		"--sharpen",
		"--blur",
		"--invert",
		"--watermark",
		"--watermark-pos",
		"--watermark-size",
		"--watermark-color",
		"--watermark-bg",
		"--interpolation",
	}

	for _, flag := range flags {
		if !strings.Contains(output, flag) {
			t.Errorf("zsh completion missing flag %q", flag)
		}
	}
}

func TestGenerateZsh_ContainsFormats(t *testing.T) {
	output := GenerateZsh()

	formats := []string{"jpg", "png", "gif", "webp", "heic", "avif"}
	for _, fmt := range formats {
		if !strings.Contains(output, fmt) {
			t.Errorf("zsh completion missing format %q", fmt)
		}
	}
}

func TestGenerateZsh_ContainsDescriptions(t *testing.T) {
	output := GenerateZsh()

	descriptions := []string{
		"output format",
		"quality level",
		"grayscale",
		"progressive",
		"lossless",
	}
	for _, desc := range descriptions {
		if !strings.Contains(strings.ToLower(output), strings.ToLower(desc)) {
			t.Errorf("zsh completion missing description containing %q", desc)
		}
	}
}

func TestGenerateFish_ContainsFlags(t *testing.T) {
	output := GenerateFish()

	flags := []string{
		"format",
		"quality",
		"grayscale",
		"progressive",
		"lossless",
		"webp-method",
		"png-compression",
		"sepia",
		"brightness",
		"contrast",
		"sharpen",
		"blur",
		"invert",
		"watermark",
		"watermark-pos",
		"watermark-size",
		"watermark-color",
		"watermark-bg",
		"interpolation",
	}

	for _, flag := range flags {
		if !strings.Contains(output, flag) {
			t.Errorf("fish completion missing flag %q", flag)
		}
	}
}

func TestGenerateFish_ContainsFormats(t *testing.T) {
	output := GenerateFish()

	formats := []string{"jpg", "png", "gif", "webp", "heic", "avif"}
	for _, fmt := range formats {
		if !strings.Contains(output, fmt) {
			t.Errorf("fish completion missing format %q", fmt)
		}
	}
}

func TestGenerateFish_ContainsDescriptions(t *testing.T) {
	output := GenerateFish()

	descriptions := []string{
		"Output format",
		"Quality level",
		"grayscale",
		"progressive",
		"lossless",
	}
	for _, desc := range descriptions {
		found := false
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(strings.ToLower(line), strings.ToLower(desc)) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("fish completion missing description containing %q", desc)
		}
	}
}

func TestGenerateBash_IsValidScript(t *testing.T) {
	output := GenerateBash()

	// Should contain the function definition and complete command
	if !strings.Contains(output, "_pixshift()") {
		t.Error("bash completion missing _pixshift() function")
	}
	if !strings.Contains(output, "complete -F _pixshift pixshift") {
		t.Error("bash completion missing complete command")
	}
}

func TestGenerateZsh_IsValidScript(t *testing.T) {
	output := GenerateZsh()

	if !strings.Contains(output, "#compdef pixshift") {
		t.Error("zsh completion missing #compdef header")
	}
	if !strings.Contains(output, "_arguments") {
		t.Error("zsh completion missing _arguments call")
	}
}

func TestGenerateFish_IsValidScript(t *testing.T) {
	output := GenerateFish()

	if !strings.Contains(output, "complete -c pixshift") {
		t.Error("fish completion missing complete command")
	}
}
