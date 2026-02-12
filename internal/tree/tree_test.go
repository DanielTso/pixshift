package tree

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create some image files
	if err := os.WriteFile(filepath.Join(dir, "photo.jpg"), []byte("fake jpg data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "icon.png"), []byte("fake png data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a subdirectory with an image
	sub := filepath.Join(dir, "subdir")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "nested.jpg"), []byte("fake nested"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a non-image file (should be excluded)
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not an image"), 0644); err != nil {
		t.Fatal(err)
	}

	return dir
}

func TestPrint_BoxDrawingChars(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	// Should contain box-drawing characters
	if !strings.Contains(output, "\u251c\u2500\u2500") && !strings.Contains(output, "\u2514\u2500\u2500") {
		t.Errorf("output should contain box-drawing chars, got:\n%s", output)
	}
}

func TestPrint_ExcludesNonImageFiles(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "readme.txt") {
		t.Errorf("output should not contain non-image file, got:\n%s", output)
	}
}

func TestPrint_IncludesImageFiles(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "photo.jpg") {
		t.Errorf("output should contain photo.jpg, got:\n%s", output)
	}
	if !strings.Contains(output, "icon.png") {
		t.Errorf("output should contain icon.png, got:\n%s", output)
	}
}

func TestPrint_IncludesSubdirectory(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "subdir/") {
		t.Errorf("output should contain subdirectory, got:\n%s", output)
	}
	if !strings.Contains(output, "nested.jpg") {
		t.Errorf("output should contain nested.jpg, got:\n%s", output)
	}
}

func TestPrint_ShowSize(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{ShowSize: true}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	// Files have sizes like "13 B" or "14 B"
	if !strings.Contains(output, " B)") && !strings.Contains(output, " KB)") {
		t.Errorf("ShowSize should display file sizes, got:\n%s", output)
	}
}

func TestPrint_ShowFormat(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{ShowFormat: true}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[JPEG]") {
		t.Errorf("ShowFormat should display JPEG label, got:\n%s", output)
	}
	if !strings.Contains(output, "[PNG]") {
		t.Errorf("ShowFormat should display PNG label, got:\n%s", output)
	}
}

func TestPrint_MaxDepth(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{MaxDepth: 1}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	// MaxDepth 1 means only immediate children, but subdirs with images
	// should still appear as entries (just not recurse into them)
	if !strings.Contains(output, "photo.jpg") {
		t.Errorf("MaxDepth 1 should still show top-level files, got:\n%s", output)
	}
}

func TestPrint_ErrorOnFile(t *testing.T) {
	// Passing a file instead of a directory should return an error
	dir := t.TempDir()
	f := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(f, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := Print(&buf, f, Options{})
	if err == nil {
		t.Error("Print on a file should return an error")
	}
}

func TestPrint_EmptyDirWithNoImages(t *testing.T) {
	dir := t.TempDir()
	// Create a non-image file
	if err := os.WriteFile(filepath.Join(dir, "doc.txt"), []byte("text"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	// Should only contain the root directory line
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("empty dir should have only root line, got %d lines:\n%s", len(lines), output)
	}
}

func TestPrint_SubdirWithOnlyNonImages_Hidden(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "docs")
	if err := os.Mkdir(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "readme.md"), []byte("text"), 0644); err != nil {
		t.Fatal(err)
	}
	// Add one image at top level so output is not totally empty
	if err := os.WriteFile(filepath.Join(dir, "img.png"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "docs/") {
		t.Errorf("subdirectory with no images should be hidden, got:\n%s", output)
	}
}

func TestPrint_RootName(t *testing.T) {
	dir := setupTestDir(t)

	var buf bytes.Buffer
	if err := Print(&buf, dir, Options{}); err != nil {
		t.Fatalf("Print: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")
	if len(lines) == 0 {
		t.Fatal("empty output")
	}
	// First line should end with /
	if !strings.HasSuffix(lines[0], "/") {
		t.Errorf("first line should end with '/', got: %s", lines[0])
	}
}

func TestPrint_NonexistentDir(t *testing.T) {
	var buf bytes.Buffer
	err := Print(&buf, "/nonexistent/path/abc123", Options{})
	if err == nil {
		t.Error("Print on nonexistent path should return an error")
	}
}
