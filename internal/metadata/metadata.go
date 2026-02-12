package metadata

import "encoding/binary"

// Metadata holds raw EXIF data extracted from a source image.
type Metadata struct {
	// EXIFRaw contains the raw EXIF bytes (APP1 payload after "Exif\0\0" prefix).
	EXIFRaw []byte
}

// HasEXIF returns true if EXIF data is present.
func (m *Metadata) HasEXIF() bool {
	return m != nil && len(m.EXIFRaw) > 0
}

// Orientation parses the EXIF orientation tag (0x0112) from raw EXIF bytes.
// Returns 0 if the orientation cannot be determined.
func (m *Metadata) Orientation() int {
	if !m.HasEXIF() {
		return 0
	}

	raw := m.EXIFRaw

	// Skip "Exif\0\0" prefix if present.
	if len(raw) > 6 && raw[0] == 'E' && raw[1] == 'x' && raw[2] == 'i' &&
		raw[3] == 'f' && raw[4] == 0 && raw[5] == 0 {
		raw = raw[6:]
	}

	if len(raw) < 8 {
		return 0
	}

	// Determine byte order from TIFF header.
	var bo binary.ByteOrder
	switch {
	case raw[0] == 'I' && raw[1] == 'I':
		bo = binary.LittleEndian
	case raw[0] == 'M' && raw[1] == 'M':
		bo = binary.BigEndian
	default:
		return 0
	}

	// Verify TIFF magic number (42).
	if bo.Uint16(raw[2:4]) != 42 {
		return 0
	}

	// Read offset to first IFD.
	ifdOffset := int(bo.Uint32(raw[4:8]))
	if ifdOffset < 8 || ifdOffset+2 > len(raw) {
		return 0
	}

	// Read number of IFD entries.
	numEntries := int(bo.Uint16(raw[ifdOffset : ifdOffset+2]))
	entryStart := ifdOffset + 2

	// Each IFD entry is 12 bytes: tag(2) + type(2) + count(4) + value(4).
	for i := 0; i < numEntries; i++ {
		offset := entryStart + i*12
		if offset+12 > len(raw) {
			break
		}

		tag := bo.Uint16(raw[offset : offset+2])
		if tag == 0x0112 { // Orientation tag
			// Type should be SHORT (3), count should be 1.
			valOffset := offset + 8
			orientation := int(bo.Uint16(raw[valOffset : valOffset+2]))
			if orientation >= 1 && orientation <= 8 {
				return orientation
			}
			return 0
		}
	}

	return 0
}
