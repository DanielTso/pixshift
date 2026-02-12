package metadata

import "testing"

func TestHasEXIF_Nil(t *testing.T) {
	var m *Metadata
	if m.HasEXIF() {
		t.Error("HasEXIF() on nil Metadata should return false")
	}
}

func TestHasEXIF_EmptyEXIFRaw(t *testing.T) {
	m := &Metadata{EXIFRaw: []byte{}}
	if m.HasEXIF() {
		t.Error("HasEXIF() with empty EXIFRaw should return false")
	}
}

func TestHasEXIF_WithData(t *testing.T) {
	m := &Metadata{EXIFRaw: []byte{0x45, 0x78, 0x69, 0x66}}
	if !m.HasEXIF() {
		t.Error("HasEXIF() with non-empty EXIFRaw should return true")
	}
}

func TestHasEXIF_NilEXIFRaw(t *testing.T) {
	m := &Metadata{EXIFRaw: nil}
	if m.HasEXIF() {
		t.Error("HasEXIF() with nil EXIFRaw should return false")
	}
}
