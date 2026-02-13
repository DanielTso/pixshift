package sdk

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

// createTestJPEG creates a small JPEG file for testing and returns its path.
func createTestJPEG(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "test.jpg")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test file: %v", err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode test jpeg: %v", err)
	}
	return path
}

func TestConvert(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.png")

	if err := Convert(inputPath, outputPath); err != nil {
		t.Fatalf("Convert: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}

	// Verify it's a valid PNG by checking magic bytes.
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) < 4 || data[0] != 0x89 || data[1] != 'P' || data[2] != 'N' || data[3] != 'G' {
		t.Error("output is not a valid PNG file")
	}
}

func TestConvertBytes(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	inputData, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read input: %v", err)
	}

	output, err := ConvertBytes(inputData, PNG)
	if err != nil {
		t.Fatalf("ConvertBytes: %v", err)
	}
	if len(output) == 0 {
		t.Error("output bytes are empty")
	}
	// Verify PNG magic bytes.
	if len(output) < 4 || output[0] != 0x89 || output[1] != 'P' || output[2] != 'N' || output[3] != 'G' {
		t.Error("output is not a valid PNG")
	}
}

func TestConvertWithOptions(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.jpg")

	if err := Convert(inputPath, outputPath, WithQuality(50)); err != nil {
		t.Fatalf("Convert with quality: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestAnalyze(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	info, err := Analyze(inputPath)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if info.Width != 100 {
		t.Errorf("width = %d, want 100", info.Width)
	}
	if info.Height != 80 {
		t.Errorf("height = %d, want 80", info.Height)
	}
	if info.Format != JPEG {
		t.Errorf("format = %q, want %q", info.Format, JPEG)
	}
	if info.Size == 0 {
		t.Error("size should be > 0")
	}
}

func TestPalette(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	colors, err := Palette(inputPath, 3)
	if err != nil {
		t.Fatalf("Palette: %v", err)
	}
	if len(colors) == 0 {
		t.Fatal("expected at least 1 color")
	}

	for _, c := range colors {
		if len(c.Hex) != 7 || c.Hex[0] != '#' {
			t.Errorf("invalid hex format: %q", c.Hex)
		}
	}
}

func TestCompare(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	score, err := Compare(inputPath, inputPath)
	if err != nil {
		t.Fatalf("Compare: %v", err)
	}
	if score < 0.99 {
		t.Errorf("SSIM of identical image = %f, want ~1.0", score)
	}
}
