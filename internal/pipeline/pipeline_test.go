package pipeline

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/DanielTso/pixshift/internal/codec"
)

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

func TestPipeline_JPEGtoPNG(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.png")

	p := NewPipeline(codec.DefaultRegistry())
	inputSize, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if inputSize == 0 {
		t.Error("inputSize should be > 0")
	}
	if outputSize == 0 {
		t.Error("outputSize should be > 0")
	}

	// Verify output file exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestPipeline_WithResize(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "resized.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		MaxDim:       50,
	})

	if err != nil {
		t.Fatalf("Execute with resize: %v", err)
	}

	// Verify output exists
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestPipeline_JPEGtoJPEG(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.jpg")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.JPEG,
		Quality:      85,
	})

	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("outputSize should be > 0")
	}
}

func TestPipeline_InputNotFound(t *testing.T) {
	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    "/nonexistent/file.jpg",
		OutputPath:   "/tmp/output.png",
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err == nil {
		t.Error("expected error for nonexistent input, got nil")
	}
}
