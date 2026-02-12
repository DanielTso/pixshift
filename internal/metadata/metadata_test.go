package metadata

import (
	"encoding/binary"
	"testing"
)

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

// buildTestEXIF creates a minimal EXIF blob with the given orientation value.
// Format: "Exif\0\0" + TIFF header (II, 42, offset=8) + IFD0 with 1 entry (orientation tag).
func buildTestEXIF(orientation int, bigEndian bool) []byte {
	var buf []byte
	// Exif prefix
	buf = append(buf, 'E', 'x', 'i', 'f', 0, 0)

	// TIFF header
	var bo binary.ByteOrder
	if bigEndian {
		buf = append(buf, 'M', 'M')
		bo = binary.BigEndian
	} else {
		buf = append(buf, 'I', 'I')
		bo = binary.LittleEndian
	}

	// Magic 42
	magic := make([]byte, 2)
	bo.PutUint16(magic, 42)
	buf = append(buf, magic...)

	// Offset to IFD0 (8 bytes from TIFF start)
	offset := make([]byte, 4)
	bo.PutUint32(offset, 8)
	buf = append(buf, offset...)

	// IFD0: 1 entry
	numEntries := make([]byte, 2)
	bo.PutUint16(numEntries, 1)
	buf = append(buf, numEntries...)

	// IFD entry: tag=0x0112, type=SHORT(3), count=1, value=orientation
	tag := make([]byte, 2)
	bo.PutUint16(tag, 0x0112)
	buf = append(buf, tag...)

	typ := make([]byte, 2)
	bo.PutUint16(typ, 3) // SHORT
	buf = append(buf, typ...)

	count := make([]byte, 4)
	bo.PutUint32(count, 1)
	buf = append(buf, count...)

	val := make([]byte, 4) // value field (4 bytes, only first 2 used for SHORT)
	bo.PutUint16(val, uint16(orientation))
	buf = append(buf, val...)

	return buf
}

func TestOrientation_Nil(t *testing.T) {
	var m *Metadata
	if m.Orientation() != 0 {
		t.Error("Orientation() on nil Metadata should return 0")
	}
}

func TestOrientation_Empty(t *testing.T) {
	m := &Metadata{EXIFRaw: []byte{}}
	if m.Orientation() != 0 {
		t.Error("Orientation() on empty EXIFRaw should return 0")
	}
}

func TestOrientation_LittleEndian(t *testing.T) {
	for orient := 1; orient <= 8; orient++ {
		m := &Metadata{EXIFRaw: buildTestEXIF(orient, false)}
		got := m.Orientation()
		if got != orient {
			t.Errorf("Orientation() LE: got %d, want %d", got, orient)
		}
	}
}

func TestOrientation_BigEndian(t *testing.T) {
	for orient := 1; orient <= 8; orient++ {
		m := &Metadata{EXIFRaw: buildTestEXIF(orient, true)}
		got := m.Orientation()
		if got != orient {
			t.Errorf("Orientation() BE: got %d, want %d", got, orient)
		}
	}
}

func TestOrientation_InvalidOrientation(t *testing.T) {
	m := &Metadata{EXIFRaw: buildTestEXIF(0, false)}
	if m.Orientation() != 0 {
		t.Error("Orientation() with value 0 should return 0")
	}

	m = &Metadata{EXIFRaw: buildTestEXIF(9, false)}
	if m.Orientation() != 0 {
		t.Error("Orientation() with value 9 should return 0")
	}
}

func TestOrientation_GarbageData(t *testing.T) {
	m := &Metadata{EXIFRaw: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}}
	if m.Orientation() != 0 {
		t.Error("Orientation() with garbage data should return 0")
	}
}

func TestOrientation_NoOrientationTag(t *testing.T) {
	// Build valid TIFF header but with a different tag (not 0x0112)
	var buf []byte
	buf = append(buf, 'E', 'x', 'i', 'f', 0, 0)
	buf = append(buf, 'I', 'I') // Little endian
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
	// Different tag
	tag := make([]byte, 2)
	bo.PutUint16(tag, 0x0100) // ImageWidth, not Orientation
	buf = append(buf, tag...)
	typ := make([]byte, 2)
	bo.PutUint16(typ, 3)
	buf = append(buf, typ...)
	count := make([]byte, 4)
	bo.PutUint32(count, 1)
	buf = append(buf, count...)
	val := make([]byte, 4)
	bo.PutUint16(val, 800)
	buf = append(buf, val...)

	m := &Metadata{EXIFRaw: buf}
	if m.Orientation() != 0 {
		t.Error("Orientation() with no orientation tag should return 0")
	}
}
