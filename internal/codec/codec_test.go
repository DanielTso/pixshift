package codec

import "testing"

func TestDefaultExtension(t *testing.T) {
	tests := []struct {
		format Format
		want   string
	}{
		{JPEG, ".jpg"},
		{PNG, ".png"},
		{GIF, ".gif"},
		{WebP, ".webp"},
		{TIFF, ".tiff"},
		{BMP, ".bmp"},
		{HEIC, ".heic"},
		{AVIF, ".avif"},
		{CR2, ".cr2"},
		{NEF, ".nef"},
		{DNG, ".dng"},
	}
	for _, tt := range tests {
		t.Run(string(tt.format), func(t *testing.T) {
			got := DefaultExtension(tt.format)
			if got != tt.want {
				t.Errorf("DefaultExtension(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestDefaultExtension_Unknown(t *testing.T) {
	got := DefaultExtension(Format("xyz"))
	if got != ".xyz" {
		t.Errorf("DefaultExtension(\"xyz\") = %q, want %q", got, ".xyz")
	}
}

func TestIsSupportedExtension(t *testing.T) {
	valid := []string{
		".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".tif",
		".bmp", ".heic", ".heif", ".avif", ".cr2", ".nef", ".dng",
	}
	for _, ext := range valid {
		if !IsSupportedExtension(ext) {
			t.Errorf("IsSupportedExtension(%q) = false, want true", ext)
		}
	}

	// Case insensitive
	if !IsSupportedExtension(".JPG") {
		t.Error("IsSupportedExtension(\".JPG\") = false, want true (case-insensitive)")
	}
	if !IsSupportedExtension(".Png") {
		t.Error("IsSupportedExtension(\".Png\") = false, want true (case-insensitive)")
	}
}

func TestIsSupportedExtension_Invalid(t *testing.T) {
	invalid := []string{".txt", ".pdf", ".doc", ".mp4", ".exe", ""}
	for _, ext := range invalid {
		if IsSupportedExtension(ext) {
			t.Errorf("IsSupportedExtension(%q) = true, want false", ext)
		}
	}
}

func TestIsRAW(t *testing.T) {
	rawFormats := []Format{CR2, NEF, DNG}
	for _, f := range rawFormats {
		if !IsRAW(f) {
			t.Errorf("IsRAW(%q) = false, want true", f)
		}
	}

	nonRAW := []Format{JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC, AVIF}
	for _, f := range nonRAW {
		if IsRAW(f) {
			t.Errorf("IsRAW(%q) = true, want false", f)
		}
	}
}
