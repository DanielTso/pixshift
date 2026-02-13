package watch

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
)

// createTestJPEGFile creates a small valid JPEG file at the given path.
func createTestJPEGFile(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test JPEG: %v", err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode test JPEG: %v", err)
	}
}

func TestWatcher_BasicConversion(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()

	reg := codec.DefaultRegistry()
	pipe := pipeline.NewPipeline(reg)

	var converted atomic.Bool
	var convertResult pipeline.Result

	w := &Watcher{
		Pipeline:     pipe,
		OutputDir:    outDir,
		OutputFormat: codec.PNG,
		Quality:      90,
		OnConvert: func(r pipeline.Result) {
			convertResult = r
			converted.Store(true)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, []string{dir})
	time.Sleep(200 * time.Millisecond) // Let watcher start

	// Write a JPEG into the watched directory
	createTestJPEGFile(t, filepath.Join(dir, "test.jpg"))

	// Wait for conversion with timeout
	deadline := time.After(5 * time.Second)
	for !converted.Load() {
		select {
		case <-deadline:
			t.Fatal("timeout waiting for conversion callback")
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	// Verify conversion succeeded
	if convertResult.Error != nil {
		t.Errorf("conversion error: %v", convertResult.Error)
	}

	// Verify output file exists
	outputPath := filepath.Join(outDir, "test.png")
	if _, err := os.Stat(outputPath); err != nil {
		t.Errorf("output file not found: %v", err)
	}
}

func TestWatcher_Debounce(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()

	reg := codec.DefaultRegistry()
	pipe := pipeline.NewPipeline(reg)

	var convertCount atomic.Int32

	w := &Watcher{
		Pipeline:     pipe,
		OutputDir:    outDir,
		OutputFormat: codec.PNG,
		Quality:      90,
		OnConvert: func(r pipeline.Result) {
			convertCount.Add(1)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, []string{dir})
	time.Sleep(200 * time.Millisecond) // Let watcher start

	// Write a file and then immediately overwrite it
	testPath := filepath.Join(dir, "debounce_test.jpg")
	createTestJPEGFile(t, testPath)
	time.Sleep(100 * time.Millisecond) // Small delay, but less than debounce (500ms)
	createTestJPEGFile(t, testPath)     // Overwrite

	// Wait for the debounced conversion plus some margin
	time.Sleep(2 * time.Second)

	count := convertCount.Load()
	if count > 1 {
		t.Errorf("expected at most 1 conversion (debounced), got %d", count)
	}
}

func TestWatcher_UnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()

	reg := codec.DefaultRegistry()
	pipe := pipeline.NewPipeline(reg)

	var converted atomic.Bool

	w := &Watcher{
		Pipeline:     pipe,
		OutputDir:    outDir,
		OutputFormat: codec.PNG,
		Quality:      90,
		OnConvert: func(r pipeline.Result) {
			converted.Store(true)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, []string{dir})
	time.Sleep(200 * time.Millisecond) // Let watcher start

	// Write a .txt file (unsupported extension)
	txtPath := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(txtPath, []byte("hello"), 0644); err != nil {
		t.Fatalf("write test.txt: %v", err)
	}

	// Wait a bit to make sure no conversion fires
	time.Sleep(1500 * time.Millisecond)

	if converted.Load() {
		t.Error("conversion should not fire for unsupported .txt extension")
	}
}

func TestWatcher_ContextCancel(t *testing.T) {
	dir := t.TempDir()

	reg := codec.DefaultRegistry()
	pipe := pipeline.NewPipeline(reg)

	w := &Watcher{
		Pipeline:     pipe,
		OutputDir:    t.TempDir(),
		OutputFormat: codec.PNG,
		Quality:      90,
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- w.Watch(ctx, []string{dir})
	}()

	time.Sleep(200 * time.Millisecond) // Let watcher start
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Watch() returned error on cancel: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for Watch() to return after context cancel")
	}
}

func TestWatcher_SameFormatSkipped(t *testing.T) {
	dir := t.TempDir()
	outDir := t.TempDir()

	reg := codec.DefaultRegistry()
	pipe := pipeline.NewPipeline(reg)

	var converted atomic.Bool

	// Output format is JPEG, same as input
	w := &Watcher{
		Pipeline:     pipe,
		OutputDir:    outDir,
		OutputFormat: codec.JPEG,
		Quality:      90,
		OnConvert: func(r pipeline.Result) {
			converted.Store(true)
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go w.Watch(ctx, []string{dir})
	time.Sleep(200 * time.Millisecond)

	// Write a JPEG file - should be skipped since output format is also JPEG
	createTestJPEGFile(t, filepath.Join(dir, "same_format.jpg"))

	time.Sleep(1500 * time.Millisecond)

	if converted.Load() {
		t.Error("conversion should not fire when input format matches output format")
	}
}

func TestBuildOutputPath(t *testing.T) {
	tests := []struct {
		name      string
		inputPath string
		outputDir string
		format    codec.Format
		want      string
	}{
		{
			name:      "with output dir",
			inputPath: "/home/user/photos/vacation.jpg",
			outputDir: "/tmp/output",
			format:    codec.PNG,
			want:      "/tmp/output/vacation.png",
		},
		{
			name:      "without output dir",
			inputPath: "/home/user/photos/vacation.jpg",
			outputDir: "",
			format:    codec.WebP,
			want:      "/home/user/photos/vacation.webp",
		},
		{
			name:      "gif output",
			inputPath: "/data/image.png",
			outputDir: "/data/converted",
			format:    codec.GIF,
			want:      "/data/converted/image.gif",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildOutputPath(tt.inputPath, tt.outputDir, tt.format)
			if got != tt.want {
				t.Errorf("buildOutputPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
