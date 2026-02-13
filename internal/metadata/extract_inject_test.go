package metadata

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"testing"

	"github.com/DanielTso/pixshift/internal/codec"
)

// buildTestJPEGWithExif creates a minimal valid JPEG with an APP1 EXIF segment.
func buildTestJPEGWithExif(exifData []byte) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{0xFF, 0xD8}) // SOI

	// APP1 segment
	segLen := len(exifData) + 2
	buf.Write([]byte{0xFF, 0xE1, byte(segLen >> 8), byte(segLen & 0xFF)})
	buf.Write(exifData)

	// Minimal JPEG: encode a 1x1 image
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	tmpBuf := bytes.Buffer{}
	jpeg.Encode(&tmpBuf, img, nil)
	// Skip SOI of the encoded JPEG (first 2 bytes)
	buf.Write(tmpBuf.Bytes()[2:])

	return buf.Bytes()
}

// buildTestJPEGNoExif creates a minimal valid JPEG without any APP1 segment.
func buildTestJPEGNoExif() []byte {
	var buf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 128, G: 64, B: 32, A: 255})
	jpeg.Encode(&buf, img, nil)
	return buf.Bytes()
}

// buildMinimalTIFF creates a minimal TIFF-like data with II header and 42 magic.
func buildMinimalTIFF() []byte {
	// Little-endian TIFF header: "II" + magic 42 + offset to IFD0
	data := []byte{
		'I', 'I',       // byte order: little-endian
		42, 0,           // magic number 42 (LE)
		8, 0, 0, 0,     // offset to IFD0 = 8
		0, 0,            // zero IFD entries (valid minimal TIFF)
	}
	return data
}

func TestExtractJPEG(t *testing.T) {
	exifData := buildTestEXIF(6, false) // orientation = 6, little-endian
	jpegData := buildTestJPEGWithExif(exifData)

	r := bytes.NewReader(jpegData)
	meta, err := Extract(r, codec.JPEG)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}
	if meta == nil {
		t.Fatal("Extract() returned nil metadata")
	}
	if !meta.HasEXIF() {
		t.Error("expected metadata to have EXIF data")
	}
	if len(meta.EXIFRaw) == 0 {
		t.Error("expected non-empty EXIFRaw")
	}

	// Verify orientation can be parsed from extracted EXIF
	orient := meta.Orientation()
	if orient != 6 {
		t.Errorf("Orientation() = %d, want 6", orient)
	}
}

func TestExtractJPEG_BigEndian(t *testing.T) {
	exifData := buildTestEXIF(3, true) // orientation = 3, big-endian
	jpegData := buildTestJPEGWithExif(exifData)

	r := bytes.NewReader(jpegData)
	meta, err := Extract(r, codec.JPEG)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}
	if meta == nil {
		t.Fatal("Extract() returned nil metadata")
	}

	orient := meta.Orientation()
	if orient != 3 {
		t.Errorf("Orientation() = %d, want 3", orient)
	}
}

func TestExtractJPEG_NoExif(t *testing.T) {
	jpegData := buildTestJPEGNoExif()

	r := bytes.NewReader(jpegData)
	_, err := Extract(r, codec.JPEG)
	if err == nil {
		t.Error("expected error for JPEG without EXIF, got nil")
	}
}

func TestExtractJPEG_SeekerResets(t *testing.T) {
	// Verify the reader position is reset after extraction
	exifData := buildTestEXIF(1, false)
	jpegData := buildTestJPEGWithExif(exifData)

	r := bytes.NewReader(jpegData)
	_, err := Extract(r, codec.JPEG)
	if err != nil {
		t.Fatalf("Extract() returned error: %v", err)
	}

	// Reader should be back at position 0
	pos, err := r.Seek(0, 1) // SeekCurrent
	if err != nil {
		t.Fatalf("Seek error: %v", err)
	}
	if pos != 0 {
		t.Errorf("reader position after Extract = %d, want 0", pos)
	}
}

func TestInjectJPEG_Roundtrip(t *testing.T) {
	// Create a valid JPEG file
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}

	// Write JPEG to temp file
	tmpFile := t.TempDir() + "/roundtrip.jpg"
	if err := os.WriteFile(tmpFile, jpegBuf.Bytes(), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Create metadata with known EXIF data
	exifData := buildTestEXIF(5, false) // orientation = 5, little-endian
	meta := &Metadata{EXIFRaw: exifData}

	// Inject metadata
	if err := Inject(tmpFile, codec.JPEG, meta); err != nil {
		t.Fatalf("Inject() error: %v", err)
	}

	// Re-extract metadata
	f, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("open temp file: %v", err)
	}
	defer f.Close()

	extracted, err := Extract(f, codec.JPEG)
	if err != nil {
		t.Fatalf("Extract() after inject error: %v", err)
	}
	if !extracted.HasEXIF() {
		t.Error("re-extracted metadata has no EXIF")
	}

	// Verify orientation matches what was injected
	orient := extracted.Orientation()
	if orient != 5 {
		t.Errorf("re-extracted Orientation() = %d, want 5", orient)
	}
}

func TestInjectJPEG_PreservesImage(t *testing.T) {
	// Create a valid JPEG image
	origImg := image.NewRGBA(image.Rect(0, 0, 20, 15))
	for y := 0; y < 15; y++ {
		for x := 0; x < 20; x++ {
			origImg.Set(x, y, color.RGBA{R: 50, G: 100, B: 200, A: 255})
		}
	}

	var jpegBuf bytes.Buffer
	if err := jpeg.Encode(&jpegBuf, origImg, &jpeg.Options{Quality: 95}); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}

	tmpFile := t.TempDir() + "/preserve.jpg"
	if err := os.WriteFile(tmpFile, jpegBuf.Bytes(), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Inject EXIF metadata
	exifData := buildTestEXIF(1, false)
	meta := &Metadata{EXIFRaw: exifData}
	if err := Inject(tmpFile, codec.JPEG, meta); err != nil {
		t.Fatalf("Inject() error: %v", err)
	}

	// Verify the output is still a valid JPEG that can be decoded
	f, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("open temp file: %v", err)
	}
	defer f.Close()

	decoded, err := jpeg.Decode(f)
	if err != nil {
		t.Fatalf("jpeg.Decode after inject: %v", err)
	}

	bounds := decoded.Bounds()
	if bounds.Dx() != 20 || bounds.Dy() != 15 {
		t.Errorf("decoded dimensions = %dx%d, want 20x15", bounds.Dx(), bounds.Dy())
	}
}

func TestInjectJPEG_ReplacesExistingExif(t *testing.T) {
	// Create a JPEG that already has EXIF (orientation=3)
	oldExif := buildTestEXIF(3, false)
	jpegData := buildTestJPEGWithExif(oldExif)

	tmpFile := t.TempDir() + "/replace_exif.jpg"
	if err := os.WriteFile(tmpFile, jpegData, 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	// Inject new EXIF (orientation=7)
	newExif := buildTestEXIF(7, false)
	meta := &Metadata{EXIFRaw: newExif}
	if err := Inject(tmpFile, codec.JPEG, meta); err != nil {
		t.Fatalf("Inject() error: %v", err)
	}

	// Re-extract and verify new orientation
	f, err := os.Open(tmpFile)
	if err != nil {
		t.Fatalf("open temp file: %v", err)
	}
	defer f.Close()

	extracted, err := Extract(f, codec.JPEG)
	if err != nil {
		t.Fatalf("Extract() after inject: %v", err)
	}

	orient := extracted.Orientation()
	if orient != 7 {
		t.Errorf("Orientation() = %d, want 7 (new value)", orient)
	}
}

func TestInjectJPEG_NilMetadata(t *testing.T) {
	// Inject with no EXIF should be a no-op
	tmpFile := t.TempDir() + "/noexif.jpg"

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, nil)
	os.WriteFile(tmpFile, buf.Bytes(), 0644)

	meta := &Metadata{EXIFRaw: nil}
	err := Inject(tmpFile, codec.JPEG, meta)
	if err != nil {
		t.Errorf("Inject() with nil EXIF should not error, got: %v", err)
	}
}

func TestExtractTIFF(t *testing.T) {
	tiffData := buildMinimalTIFF()

	r := bytes.NewReader(tiffData)
	meta, err := Extract(r, codec.TIFF)
	if err != nil {
		t.Fatalf("Extract(TIFF) returned error: %v", err)
	}
	if meta == nil {
		t.Fatal("Extract(TIFF) returned nil metadata")
	}
	if !meta.HasEXIF() {
		t.Error("expected TIFF metadata to have EXIF data")
	}

	// The TIFF extractor wraps the entire file with "Exif\0\0" prefix
	if len(meta.EXIFRaw) < 6 {
		t.Fatalf("EXIFRaw too short: %d bytes", len(meta.EXIFRaw))
	}
	prefix := string(meta.EXIFRaw[:4])
	if prefix != "Exif" {
		t.Errorf("EXIFRaw prefix = %q, want %q", prefix, "Exif")
	}
}

func TestExtractTIFF_TooSmall(t *testing.T) {
	// TIFF file must be at least 8 bytes
	tinyData := []byte{'I', 'I', 42, 0}

	r := bytes.NewReader(tinyData)
	_, err := Extract(r, codec.TIFF)
	if err == nil {
		t.Error("expected error for TIFF data smaller than 8 bytes, got nil")
	}
}

func TestExtractTIFF_BigEndian(t *testing.T) {
	data := []byte{
		'M', 'M',       // byte order: big-endian
		0, 42,           // magic number 42 (BE)
		0, 0, 0, 8,     // offset to IFD0 = 8
		0, 0,            // zero IFD entries
	}

	r := bytes.NewReader(data)
	meta, err := Extract(r, codec.TIFF)
	if err != nil {
		t.Fatalf("Extract(TIFF BE) returned error: %v", err)
	}
	if !meta.HasEXIF() {
		t.Error("expected TIFF BE metadata to have EXIF data")
	}
}

func TestExtract_UnsupportedFormat(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes

	r := bytes.NewReader(data)
	_, err := Extract(r, codec.PNG)
	if err == nil {
		t.Error("expected error for unsupported format (PNG), got nil")
	}
}

func TestExtract_UnsupportedFormat_GIF(t *testing.T) {
	data := []byte("GIF89a")

	r := bytes.NewReader(data)
	_, err := Extract(r, codec.GIF)
	if err == nil {
		t.Error("expected error for unsupported format (GIF), got nil")
	}
}

func TestExtract_UnsupportedFormat_WebP(t *testing.T) {
	data := []byte("RIFF....WEBP")

	r := bytes.NewReader(data)
	_, err := Extract(r, codec.WebP)
	if err == nil {
		t.Error("expected error for unsupported format (WebP), got nil")
	}
}

func TestExtract_UnsupportedFormat_BMP(t *testing.T) {
	data := []byte("BM")

	r := bytes.NewReader(data)
	_, err := Extract(r, codec.BMP)
	if err == nil {
		t.Error("expected error for unsupported format (BMP), got nil")
	}
}

func TestExtract_CR2(t *testing.T) {
	// CR2 uses the TIFF code path
	data := []byte{
		'I', 'I',
		42, 0,
		8, 0, 0, 0,
		0, 0,
	}

	r := bytes.NewReader(data)
	meta, err := Extract(r, codec.CR2)
	if err != nil {
		t.Fatalf("Extract(CR2) returned error: %v", err)
	}
	if !meta.HasEXIF() {
		t.Error("expected CR2 metadata to have EXIF data")
	}
}

func TestExtract_NEF(t *testing.T) {
	data := []byte{
		'M', 'M',
		0, 42,
		0, 0, 0, 8,
		0, 0,
	}

	r := bytes.NewReader(data)
	meta, err := Extract(r, codec.NEF)
	if err != nil {
		t.Fatalf("Extract(NEF) returned error: %v", err)
	}
	if !meta.HasEXIF() {
		t.Error("expected NEF metadata to have EXIF data")
	}
}

func TestExtract_DNG(t *testing.T) {
	data := []byte{
		'I', 'I',
		42, 0,
		8, 0, 0, 0,
		0, 0,
	}

	r := bytes.NewReader(data)
	meta, err := Extract(r, codec.DNG)
	if err != nil {
		t.Fatalf("Extract(DNG) returned error: %v", err)
	}
	if !meta.HasEXIF() {
		t.Error("expected DNG metadata to have EXIF data")
	}
}

func TestInject_UnsupportedFormat(t *testing.T) {
	tmpFile := t.TempDir() + "/test.webp"
	os.WriteFile(tmpFile, []byte("fake"), 0644)

	meta := &Metadata{EXIFRaw: buildTestEXIF(1, false)}
	err := Inject(tmpFile, codec.WebP, meta)
	if err == nil {
		t.Error("expected error for unsupported injection format (WebP), got nil")
	}
}

func TestInject_TIFFNoOp(t *testing.T) {
	// TIFF injection is a no-op (returns nil)
	tmpFile := t.TempDir() + "/test.tiff"
	os.WriteFile(tmpFile, []byte("fake"), 0644)

	meta := &Metadata{EXIFRaw: buildTestEXIF(1, false)}
	err := Inject(tmpFile, codec.TIFF, meta)
	if err != nil {
		t.Errorf("Inject(TIFF) should be a no-op, got error: %v", err)
	}
}

func TestStripExistingExif(t *testing.T) {
	// Build JPEG with EXIF, strip it, verify no EXIF remains
	exifData := buildTestEXIF(4, false)
	jpegData := buildTestJPEGWithExif(exifData)

	stripped := stripExistingExif(jpegData)

	// Verify SOI is still present
	if len(stripped) < 2 || stripped[0] != 0xFF || stripped[1] != 0xD8 {
		t.Error("stripped JPEG should start with SOI marker")
	}

	// Verify no APP1 EXIF segment exists
	found := findJPEGExifSegment(stripped)
	if found != nil {
		t.Error("stripped JPEG should not contain EXIF segment")
	}
}

func TestBuildAPP1Segment(t *testing.T) {
	payload := []byte("Exif\x00\x00test-data")
	segment := buildAPP1Segment(payload)

	// Should start with FF E1
	if segment[0] != 0xFF || segment[1] != 0xE1 {
		t.Errorf("APP1 marker = %02X %02X, want FF E1", segment[0], segment[1])
	}

	// Length field should be len(payload) + 2
	expectedLen := len(payload) + 2
	gotLen := int(segment[2])<<8 | int(segment[3])
	if gotLen != expectedLen {
		t.Errorf("APP1 length = %d, want %d", gotLen, expectedLen)
	}

	// Payload should follow
	gotPayload := segment[4:]
	if !bytes.Equal(gotPayload, payload) {
		t.Error("APP1 payload does not match input")
	}
}

func TestFindJPEGExifSegment_NotJPEG(t *testing.T) {
	data := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic
	result := findJPEGExifSegment(data)
	if result != nil {
		t.Error("expected nil for non-JPEG data")
	}
}

func TestFindJPEGExifSegment_TooShort(t *testing.T) {
	data := []byte{0xFF}
	result := findJPEGExifSegment(data)
	if result != nil {
		t.Error("expected nil for data shorter than 2 bytes")
	}
}
