package contact

import (
	"image"
	"image/color"
	"testing"
)

func testImg(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 37) % 256),
				G: uint8((y * 59) % 256),
				B: 100,
				A: 255,
			})
		}
	}
	return img
}

func TestGenerate_EmptyInput_Returns1x1(t *testing.T) {
	out := Generate(nil, DefaultOptions())
	b := out.Bounds()
	if b.Dx() != 1 || b.Dy() != 1 {
		t.Errorf("empty input: got %dx%d, want 1x1", b.Dx(), b.Dy())
	}
}

func TestGenerate_SingleImage_DefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	entries := []ImageEntry{
		{Path: "test.jpg", Image: testImg(400, 300)},
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	// 1 image, 4 columns -> 1 row, 1 col used but grid allocates for 4 cols
	// totalW = 4*(200+10) + 10 = 850
	// totalH = 1*(200+20+10) + 10 = 240
	expectedW := 4*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 1*(opts.ThumbSize+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("single image: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_MultipleImages_GridSize(t *testing.T) {
	opts := DefaultOptions()
	entries := make([]ImageEntry, 5)
	for i := range entries {
		entries[i] = ImageEntry{Path: "img.jpg", Image: testImg(300, 200)}
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	// 5 images, 4 columns -> 2 rows
	expectedW := 4*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 2*(opts.ThumbSize+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("5 images: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_CustomColumns(t *testing.T) {
	opts := DefaultOptions()
	opts.Columns = 2
	entries := make([]ImageEntry, 4)
	for i := range entries {
		entries[i] = ImageEntry{Path: "img.jpg", Image: testImg(300, 200)}
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	// 4 images, 2 columns -> 2 rows
	expectedW := 2*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 2*(opts.ThumbSize+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("2 cols, 4 images: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_CustomThumbSize(t *testing.T) {
	opts := DefaultOptions()
	opts.ThumbSize = 100
	opts.Columns = 3
	entries := make([]ImageEntry, 3)
	for i := range entries {
		entries[i] = ImageEntry{Path: "img.jpg", Image: testImg(300, 200)}
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	expectedW := 3*(100+opts.Padding) + opts.Padding
	expectedH := 1*(100+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("custom thumb 100: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_NoLabels(t *testing.T) {
	opts := DefaultOptions()
	opts.ShowLabels = false
	entries := []ImageEntry{
		{Path: "test.jpg", Image: testImg(400, 300)},
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	expectedW := 4*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 1*(opts.ThumbSize+0+opts.Padding) + opts.Padding // labelH = 0
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("no labels: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_ZeroColumns_DefaultsTo4(t *testing.T) {
	opts := DefaultOptions()
	opts.Columns = 0
	entries := []ImageEntry{
		{Path: "test.jpg", Image: testImg(400, 300)},
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	// Columns <= 0 defaults to 4
	expectedW := 4*(opts.ThumbSize+opts.Padding) + opts.Padding
	if b.Dx() != expectedW {
		t.Errorf("zero columns: width = %d, want %d (default 4 cols)", b.Dx(), expectedW)
	}
}

func TestGenerate_SmallSourceImage(t *testing.T) {
	// Image smaller than thumb size should not be upscaled
	opts := DefaultOptions()
	opts.ThumbSize = 200
	entries := []ImageEntry{
		{Path: "tiny.jpg", Image: testImg(50, 30)},
	}
	out := Generate(entries, opts)
	b := out.Bounds()
	// Should still produce correct grid dimensions
	expectedW := 4*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 1*(opts.ThumbSize+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("small source: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_ExactlyFillsGrid(t *testing.T) {
	opts := DefaultOptions()
	opts.Columns = 3
	entries := make([]ImageEntry, 6)
	for i := range entries {
		entries[i] = ImageEntry{Path: "img.jpg", Image: testImg(300, 200)}
	}
	out := Generate(entries, opts)
	b := out.Bounds()

	// 6 images, 3 columns -> 2 rows
	expectedW := 3*(opts.ThumbSize+opts.Padding) + opts.Padding
	expectedH := 2*(opts.ThumbSize+opts.LabelHeight+opts.Padding) + opts.Padding
	if b.Dx() != expectedW || b.Dy() != expectedH {
		t.Errorf("exact fill: got %dx%d, want %dx%d", b.Dx(), b.Dy(), expectedW, expectedH)
	}
}

func TestGenerate_WithLabels(t *testing.T) {
	opts := DefaultOptions()
	entries := []ImageEntry{
		{Path: "photo.jpg", Image: testImg(400, 300), Label: "Custom Label"},
	}
	// Should not panic
	out := Generate(entries, opts)
	b := out.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		t.Error("generate with labels produced zero-size image")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Columns != 4 {
		t.Errorf("Columns = %d, want 4", opts.Columns)
	}
	if opts.ThumbSize != 200 {
		t.Errorf("ThumbSize = %d, want 200", opts.ThumbSize)
	}
	if opts.Padding != 10 {
		t.Errorf("Padding = %d, want 10", opts.Padding)
	}
	if !opts.ShowLabels {
		t.Error("ShowLabels should be true by default")
	}
	if opts.LabelHeight != 20 {
		t.Errorf("LabelHeight = %d, want 20", opts.LabelHeight)
	}
}
