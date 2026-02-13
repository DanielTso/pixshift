package pipeline

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
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

// buildTestEXIF creates a minimal EXIF blob with the given orientation.
func buildTestEXIF(orientation int) []byte {
	var buf []byte
	buf = append(buf, 'E', 'x', 'i', 'f', 0, 0)
	buf = append(buf, 'I', 'I') // Little-endian
	bo := binary.LittleEndian
	magic := make([]byte, 2)
	bo.PutUint16(magic, 42)
	buf = append(buf, magic...)
	offset := make([]byte, 4)
	bo.PutUint32(offset, 8)
	buf = append(buf, offset...)
	numEntries := make([]byte, 2)
	bo.PutUint16(numEntries, 1)
	buf = append(buf, numEntries...)
	tag := make([]byte, 2)
	bo.PutUint16(tag, 0x0112)
	buf = append(buf, tag...)
	typ := make([]byte, 2)
	bo.PutUint16(typ, 3) // SHORT
	buf = append(buf, typ...)
	count := make([]byte, 4)
	bo.PutUint32(count, 1)
	buf = append(buf, count...)
	val := make([]byte, 4)
	bo.PutUint16(val, uint16(orientation))
	buf = append(buf, val...)
	return buf
}

// createTestJPEGWithExif creates a JPEG file with an embedded EXIF segment.
func createTestJPEGWithExif(t *testing.T, dir, name string, exifData []byte) string {
	t.Helper()

	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xD8}) // SOI

	// APP1 segment with EXIF data
	segLen := len(exifData) + 2
	buf.Write([]byte{0xFF, 0xE1, byte(segLen >> 8), byte(segLen & 0xFF)})
	buf.Write(exifData)

	// Encode a small image and append (skipping the SOI)
	img := image.NewRGBA(image.Rect(0, 0, 100, 80))
	for y := 0; y < 80; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var imgBuf bytes.Buffer
	if err := jpeg.Encode(&imgBuf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode JPEG: %v", err)
	}
	buf.Write(imgBuf.Bytes()[2:]) // Skip SOI

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("write test JPEG with EXIF: %v", err)
	}
	return path
}

// decodeOutputImage opens and decodes a JPEG or PNG file to image.Image.
func decodeOutputImage(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode output: %v", err)
	}
	return img
}

func diff(a, b uint32) uint32 {
	if a > b {
		return a - b
	}
	return b - a
}

func TestExecute_Backup(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.jpg")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:      inputPath,
		OutputPath:     outputPath,
		OutputFormat:   codec.JPEG,
		Quality:        85,
		BackupOriginal: true,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Verify .bak file was created
	bakPath := inputPath + ".bak"
	info, err := os.Stat(bakPath)
	if err != nil {
		t.Fatalf("backup file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("backup file is empty")
	}

	// Verify original and backup have same size
	origInfo, _ := os.Stat(inputPath)
	if origInfo.Size() != info.Size() {
		t.Errorf("backup size %d != original size %d", info.Size(), origInfo.Size())
	}
}

func TestExecute_Crop(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "cropped.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		CropWidth:    50,
		CropHeight:   40,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Decode output and verify dimensions
	img := decodeOutputImage(t, outputPath)
	bounds := img.Bounds()
	if bounds.Dx() != 50 || bounds.Dy() != 40 {
		t.Errorf("cropped dimensions = %dx%d, want 50x40", bounds.Dx(), bounds.Dy())
	}
}

func TestExecute_CropRatio(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir) // 100x80
	outputPath := filepath.Join(dir, "cropped_ratio.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:       inputPath,
		OutputPath:      outputPath,
		OutputFormat:    codec.PNG,
		Quality:         90,
		CropAspectRatio: "16:9",
		CropGravity:     "center",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Verify file exists and is not empty
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestExecute_Watermark(t *testing.T) {
	dir := t.TempDir()

	// Create the original (no watermark)
	inputPath := createTestJPEG(t, dir)
	plainPath := filepath.Join(dir, "plain.png")
	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	// Create the watermarked version
	wmPath := filepath.Join(dir, "watermarked.png")
	_, _, err = p.Execute(Job{
		InputPath:     inputPath,
		OutputPath:    wmPath,
		OutputFormat:  codec.PNG,
		Quality:       90,
		WatermarkText: "TEST WATERMARK",
		WatermarkPos:  "center",
	})
	if err != nil {
		t.Fatalf("Execute (watermark): %v", err)
	}

	// Both files should exist
	plainInfo, err := os.Stat(plainPath)
	if err != nil {
		t.Fatalf("plain output not found: %v", err)
	}
	wmInfo, err := os.Stat(wmPath)
	if err != nil {
		t.Fatalf("watermarked output not found: %v", err)
	}

	// Watermarked file should differ from plain (different size is a good indicator)
	if plainInfo.Size() == wmInfo.Size() {
		// Sizes being exactly equal would be suspicious; decode and compare pixels
		plainImg := decodeOutputImage(t, plainPath)
		wmImg := decodeOutputImage(t, wmPath)

		// Check some center pixels differ
		cx, cy := 50, 40
		pr, pg, pb, _ := plainImg.At(cx, cy).RGBA()
		wr, wg, wb, _ := wmImg.At(cx, cy).RGBA()
		if pr == wr && pg == wg && pb == wb {
			t.Error("watermarked image center pixel is identical to plain")
		}
	}
}

func TestExecute_Grayscale(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "gray.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Grayscale:    true,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Decode output and verify pixels have R==G==B
	img := decodeOutputImage(t, outputPath)
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// After grayscale conversion, R, G, B should be equal (within rounding)
			if diff(r, g) > 256 || diff(r, b) > 256 || diff(g, b) > 256 {
				t.Errorf("pixel (%d,%d) not grayscale: R=%d G=%d B=%d", x, y, r>>8, g>>8, b>>8)
				return
			}
		}
	}
}

func TestExecute_Sepia(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "sepia.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Sepia:        0.8,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Decode and verify pixels have changed from the original
	img := decodeOutputImage(t, outputPath)
	r, g, b, _ := img.At(50, 40).RGBA()
	// Original was R:100, G:150, B:200 - sepia should shift these values
	// At minimum, make sure it decoded and has some color
	if r == 0 && g == 0 && b == 0 {
		t.Error("sepia output is all black, expected color shift")
	}
}

func TestExecute_Brightness(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	plainPath := filepath.Join(dir, "plain.png")
	brightPath := filepath.Join(dir, "bright.png")

	p := NewPipeline(codec.DefaultRegistry())

	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	_, _, err = p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   brightPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Brightness:   50,
	})
	if err != nil {
		t.Fatalf("Execute (bright): %v", err)
	}

	plainImg := decodeOutputImage(t, plainPath)
	brightImg := decodeOutputImage(t, brightPath)

	pr, pg, pb, _ := plainImg.At(50, 40).RGBA()
	br, bg, bb, _ := brightImg.At(50, 40).RGBA()

	// Brightness +50 should increase values
	if br <= pr && bg <= pg && bb <= pb {
		t.Errorf("brightness +50 should increase pixel values: plain=(%d,%d,%d) bright=(%d,%d,%d)",
			pr>>8, pg>>8, pb>>8, br>>8, bg>>8, bb>>8)
	}
}

func TestExecute_Contrast(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	plainPath := filepath.Join(dir, "plain.png")
	contrastPath := filepath.Join(dir, "contrast.png")

	p := NewPipeline(codec.DefaultRegistry())

	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	_, _, err = p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   contrastPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Contrast:     50,
	})
	if err != nil {
		t.Fatalf("Execute (contrast): %v", err)
	}

	// Verify files differ
	plainData, _ := os.ReadFile(plainPath)
	contrastData, _ := os.ReadFile(contrastPath)
	if bytes.Equal(plainData, contrastData) {
		t.Error("contrast output should differ from plain output")
	}
}

func TestExecute_Sharpen(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	plainPath := filepath.Join(dir, "plain.png")
	sharpPath := filepath.Join(dir, "sharp.png")

	p := NewPipeline(codec.DefaultRegistry())

	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	_, _, err = p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   sharpPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Sharpen:      true,
	})
	if err != nil {
		t.Fatalf("Execute (sharpen): %v", err)
	}

	plainData, _ := os.ReadFile(plainPath)
	sharpData, _ := os.ReadFile(sharpPath)
	if bytes.Equal(plainData, sharpData) {
		t.Error("sharpened output should differ from plain output")
	}
}

func TestExecute_Blur(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	plainPath := filepath.Join(dir, "plain.png")
	blurPath := filepath.Join(dir, "blur.png")

	p := NewPipeline(codec.DefaultRegistry())

	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	_, _, err = p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   blurPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Blur:         2.0,
	})
	if err != nil {
		t.Fatalf("Execute (blur): %v", err)
	}

	plainData, _ := os.ReadFile(plainPath)
	blurData, _ := os.ReadFile(blurPath)
	if bytes.Equal(plainData, blurData) {
		t.Error("blurred output should differ from plain output")
	}
}

func TestExecute_Invert(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)

	plainPath := filepath.Join(dir, "plain.png")
	invertPath := filepath.Join(dir, "invert.png")

	p := NewPipeline(codec.DefaultRegistry())

	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   plainPath,
		OutputFormat: codec.PNG,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute (plain): %v", err)
	}

	_, _, err = p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   invertPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Invert:       true,
	})
	if err != nil {
		t.Fatalf("Execute (invert): %v", err)
	}

	plainImg := decodeOutputImage(t, plainPath)
	invertImg := decodeOutputImage(t, invertPath)

	// Check that center pixel is approximately inverted
	pr, pg, pb, _ := plainImg.At(50, 40).RGBA()
	ir, ig, ib, _ := invertImg.At(50, 40).RGBA()

	// Invert: new = 255 - old. In uint32 16-bit space: ~65535 - old
	// Allow tolerance for JPEG compression artifacts
	tolerance := uint32(5000) // generous tolerance for JPEG rounding
	checkInv := func(name string, orig, inv uint32) {
		expected := uint32(65535) - orig
		if diff(inv, expected) > tolerance {
			t.Errorf("inverted %s: got %d, expected ~%d (original=%d)", name, inv, expected, orig)
		}
	}
	checkInv("R", pr, ir)
	checkInv("G", pg, ig)
	checkInv("B", pb, ib)
}

func TestExecute_AdvancedEncoder_PNG(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "advanced.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		EncodeOpts: codec.EncodeOptions{
			Compression: 3, // Best compression
		},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Verify output is a valid PNG
	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("png.Decode: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 80 {
		t.Errorf("output dimensions = %dx%d, want 100x80", bounds.Dx(), bounds.Dy())
	}
}

func TestExecute_AdvancedEncoder_NoCompression(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	noCompPath := filepath.Join(dir, "nocomp.png")
	bestCompPath := filepath.Join(dir, "bestcomp.png")

	p := NewPipeline(codec.DefaultRegistry())

	// No compression
	_, noCompSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   noCompPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		EncodeOpts: codec.EncodeOptions{
			Compression: 1, // No compression
		},
	})
	if err != nil {
		t.Fatalf("Execute (no compression): %v", err)
	}

	// Best compression
	_, bestCompSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   bestCompPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		EncodeOpts: codec.EncodeOptions{
			Compression: 3, // Best compression
		},
	})
	if err != nil {
		t.Fatalf("Execute (best compression): %v", err)
	}

	// No compression should produce a larger file than best compression
	if noCompSize <= bestCompSize {
		t.Errorf("no compression (%d) should be larger than best compression (%d)", noCompSize, bestCompSize)
	}
}

func TestExecute_MetadataPreserve(t *testing.T) {
	dir := t.TempDir()

	// Create a JPEG with EXIF orientation=6
	exifData := buildTestEXIF(6)
	inputPath := createTestJPEGWithExif(t, dir, "exif_input.jpg", exifData)
	outputPath := filepath.Join(dir, "exif_output.jpg")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:        inputPath,
		OutputPath:       outputPath,
		OutputFormat:     codec.JPEG,
		Quality:          90,
		PreserveMetadata: true,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Read the output file and check for EXIF data
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	// Look for Exif marker in the output
	if !bytes.Contains(data, []byte("Exif")) {
		t.Error("output JPEG should contain EXIF data when PreserveMetadata=true")
	}
}

func TestExecute_MetadataStrip(t *testing.T) {
	dir := t.TempDir()

	// Create a JPEG with EXIF
	exifData := buildTestEXIF(3)
	inputPath := createTestJPEGWithExif(t, dir, "strip_input.jpg", exifData)
	outputPath := filepath.Join(dir, "strip_output.jpg")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:     inputPath,
		OutputPath:    outputPath,
		OutputFormat:  codec.JPEG,
		Quality:       90,
		StripMetadata: true,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Read the output JPEG and verify no EXIF APP1 segment
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}

	// The standard jpeg encoder does not inject EXIF, so StripMetadata
	// means we don't inject it. Check no Exif header in the output.
	// Search for APP1 EXIF marker pattern after SOI
	hasExif := false
	if len(data) > 4 {
		for i := 2; i < len(data)-5; i++ {
			if data[i] == 0xFF && data[i+1] == 0xE1 {
				segLen := int(data[i+2])<<8 | int(data[i+3])
				if i+4+6 <= len(data) && segLen >= 6 {
					if string(data[i+4:i+10]) == "Exif\x00\x00" {
						hasExif = true
						break
					}
				}
			}
		}
	}
	if hasExif {
		t.Error("output JPEG should not contain EXIF data when StripMetadata=true")
	}
}

func TestExecute_MultipleFilters(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "multi_filter.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Brightness:   20,
		Contrast:     10,
		Sharpen:      true,
		Grayscale:    true,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Verify it's a valid image and is grayscale
	img := decodeOutputImage(t, outputPath)
	r, g, b, _ := img.At(50, 40).RGBA()
	if diff(r, g) > 256 || diff(r, b) > 256 {
		t.Errorf("multi-filter output should be grayscale: R=%d G=%d B=%d", r>>8, g>>8, b>>8)
	}
}

func TestExecute_ResizeWidth(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "resized_width.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Width:        50,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	img := decodeOutputImage(t, outputPath)
	bounds := img.Bounds()
	if bounds.Dx() != 50 {
		t.Errorf("width = %d, want 50", bounds.Dx())
	}
}

func TestExecute_ResizeHeight(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "resized_height.png")

	p := NewPipeline(codec.DefaultRegistry())
	_, _, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.PNG,
		Quality:      90,
		Height:       40,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	img := decodeOutputImage(t, outputPath)
	bounds := img.Bounds()
	if bounds.Dy() != 40 {
		t.Errorf("height = %d, want 40", bounds.Dy())
	}
}

func TestExecute_CropGravityVariants(t *testing.T) {
	gravities := []string{"center", "north", "south", "east", "west"}

	for _, grav := range gravities {
		t.Run(grav, func(t *testing.T) {
			dir := t.TempDir()
			inputPath := createTestJPEG(t, dir)
			outputPath := filepath.Join(dir, "crop_"+grav+".png")

			p := NewPipeline(codec.DefaultRegistry())
			_, _, err := p.Execute(Job{
				InputPath:    inputPath,
				OutputPath:   outputPath,
				OutputFormat: codec.PNG,
				Quality:      90,
				CropWidth:    60,
				CropHeight:   50,
				CropGravity:  grav,
			})
			if err != nil {
				t.Fatalf("Execute with gravity %s: %v", grav, err)
			}

			img := decodeOutputImage(t, outputPath)
			bounds := img.Bounds()
			if bounds.Dx() != 60 || bounds.Dy() != 50 {
				t.Errorf("cropped dimensions = %dx%d, want 60x50", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestExecute_JPEGtoGIF(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "output.gif")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.GIF,
		Quality:      90,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}
}

func TestExecute_ProgressiveJPEG(t *testing.T) {
	dir := t.TempDir()
	inputPath := createTestJPEG(t, dir)
	outputPath := filepath.Join(dir, "progressive.jpg")

	p := NewPipeline(codec.DefaultRegistry())
	_, outputSize, err := p.Execute(Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: codec.JPEG,
		Quality:      85,
		EncodeOpts: codec.EncodeOptions{
			Progressive: true,
		},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if outputSize == 0 {
		t.Error("output size should be > 0")
	}

	// Verify it's still a valid JPEG
	f, err := os.Open(outputPath)
	if err != nil {
		t.Fatalf("open output: %v", err)
	}
	defer f.Close()

	_, err = jpeg.Decode(f)
	if err != nil {
		t.Fatalf("jpeg.Decode: %v", err)
	}
}
