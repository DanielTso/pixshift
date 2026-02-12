package preset

import (
	"sort"
	"testing"
)

func TestGet_Web(t *testing.T) {
	p, err := Get("web")
	if err != nil {
		t.Fatalf("Get(web): %v", err)
	}
	if p.Name != "web" {
		t.Errorf("Name = %q, want %q", p.Name, "web")
	}
	if p.Format != "webp" {
		t.Errorf("Format = %q, want %q", p.Format, "webp")
	}
	if p.Quality != 85 {
		t.Errorf("Quality = %d, want 85", p.Quality)
	}
	if p.MaxDim != 1920 {
		t.Errorf("MaxDim = %d, want 1920", p.MaxDim)
	}
	if !p.StripMetadata {
		t.Error("StripMetadata should be true for web preset")
	}
}

func TestGet_Thumbnail(t *testing.T) {
	p, err := Get("thumbnail")
	if err != nil {
		t.Fatalf("Get(thumbnail): %v", err)
	}
	if p.Name != "thumbnail" {
		t.Errorf("Name = %q, want %q", p.Name, "thumbnail")
	}
	if p.Format != "jpg" {
		t.Errorf("Format = %q, want %q", p.Format, "jpg")
	}
	if p.Quality != 80 {
		t.Errorf("Quality = %d, want 80", p.Quality)
	}
	if p.MaxDim != 300 {
		t.Errorf("MaxDim = %d, want 300", p.MaxDim)
	}
	if !p.StripMetadata {
		t.Error("StripMetadata should be true for thumbnail preset")
	}
}

func TestGet_Print(t *testing.T) {
	p, err := Get("print")
	if err != nil {
		t.Fatalf("Get(print): %v", err)
	}
	if p.Name != "print" {
		t.Errorf("Name = %q, want %q", p.Name, "print")
	}
	if p.Format != "tiff" {
		t.Errorf("Format = %q, want %q", p.Format, "tiff")
	}
	if p.Quality != 100 {
		t.Errorf("Quality = %d, want 100", p.Quality)
	}
	if !p.PreserveMetadata {
		t.Error("PreserveMetadata should be true for print preset")
	}
}

func TestGet_Archive(t *testing.T) {
	p, err := Get("archive")
	if err != nil {
		t.Fatalf("Get(archive): %v", err)
	}
	if p.Name != "archive" {
		t.Errorf("Name = %q, want %q", p.Name, "archive")
	}
	if p.Format != "png" {
		t.Errorf("Format = %q, want %q", p.Format, "png")
	}
	if p.Quality != 100 {
		t.Errorf("Quality = %d, want 100", p.Quality)
	}
	if !p.PreserveMetadata {
		t.Error("PreserveMetadata should be true for archive preset")
	}
}

func TestGet_UnknownPreset_Error(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Error("Get(nonexistent) should return an error")
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	p1, err := Get("web")
	if err != nil {
		t.Fatalf("Get(web): %v", err)
	}

	// Modify the returned copy
	p1.Quality = 1
	p1.Format = "modified"

	// Get again and verify built-in is unchanged
	p2, err := Get("web")
	if err != nil {
		t.Fatalf("Get(web) again: %v", err)
	}
	if p2.Quality != 85 {
		t.Errorf("built-in quality modified: got %d, want 85", p2.Quality)
	}
	if p2.Format != "webp" {
		t.Errorf("built-in format modified: got %q, want %q", p2.Format, "webp")
	}
}

func TestList_ReturnsSortedNames(t *testing.T) {
	names := List()
	if len(names) != 4 {
		t.Fatalf("List() returned %d names, want 4", len(names))
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("List() should return sorted names, got %v", names)
	}
}

func TestList_ContainsAllPresets(t *testing.T) {
	names := List()
	expected := map[string]bool{"web": false, "thumbnail": false, "print": false, "archive": false}
	for _, n := range names {
		if _, ok := expected[n]; ok {
			expected[n] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("List() missing preset %q", name)
		}
	}
}

func TestLoadCustomPresets(t *testing.T) {
	defer ResetCustomPresets()

	LoadCustomPresets(map[string]*Preset{
		"social": {
			Format:        "jpg",
			Quality:       90,
			MaxDim:        1080,
			StripMetadata: true,
			Grayscale:     true,
		},
	})

	p, err := Get("social")
	if err != nil {
		t.Fatalf("Get(social): %v", err)
	}
	if p.Format != "jpg" {
		t.Errorf("Format = %q, want %q", p.Format, "jpg")
	}
	if p.Quality != 90 {
		t.Errorf("Quality = %d, want 90", p.Quality)
	}
	if !p.Grayscale {
		t.Error("Grayscale should be true")
	}
}

func TestCustomPresetOverridesBuiltin(t *testing.T) {
	defer ResetCustomPresets()

	LoadCustomPresets(map[string]*Preset{
		"web": {
			Format:  "jpg",
			Quality: 70,
		},
	})

	p, err := Get("web")
	if err != nil {
		t.Fatalf("Get(web): %v", err)
	}
	if p.Format != "jpg" {
		t.Errorf("custom web Format = %q, want %q", p.Format, "jpg")
	}
	if p.Quality != 70 {
		t.Errorf("custom web Quality = %d, want 70", p.Quality)
	}
}

func TestListIncludesCustomPresets(t *testing.T) {
	defer ResetCustomPresets()

	LoadCustomPresets(map[string]*Preset{
		"social": {Format: "jpg", Quality: 90},
	})

	names := List()
	found := false
	for _, n := range names {
		if n == "social" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("List() should include custom preset 'social', got %v", names)
	}
}

func TestResetCustomPresets(t *testing.T) {
	LoadCustomPresets(map[string]*Preset{
		"social": {Format: "jpg"},
	})
	ResetCustomPresets()

	_, err := Get("social")
	if err == nil {
		t.Error("after reset, custom preset should not be found")
	}
}

func TestGet_AllListedPresetsAreValid(t *testing.T) {
	for _, name := range List() {
		p, err := Get(name)
		if err != nil {
			t.Errorf("Get(%q) failed: %v", name, err)
			continue
		}
		if p.Name != name {
			t.Errorf("Get(%q).Name = %q", name, p.Name)
		}
	}
}
