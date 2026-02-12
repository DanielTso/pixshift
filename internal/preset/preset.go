package preset

import (
	"fmt"
	"sort"
)

// Preset defines a named conversion configuration.
type Preset struct {
	Name             string
	Format           string
	Quality          int
	MaxDim           int
	Width            int
	Height           int
	StripMetadata    bool
	PreserveMetadata bool
}

var builtins = map[string]*Preset{
	"web": {
		Name:          "web",
		Format:        "webp",
		Quality:       85,
		MaxDim:        1920,
		StripMetadata: true,
	},
	"thumbnail": {
		Name:          "thumbnail",
		Format:        "jpg",
		Quality:       80,
		MaxDim:        300,
		StripMetadata: true,
	},
	"print": {
		Name:             "print",
		Format:           "tiff",
		Quality:          100,
		PreserveMetadata: true,
	},
	"archive": {
		Name:             "archive",
		Format:           "png",
		Quality:          100,
		PreserveMetadata: true,
	},
}

// Get returns a built-in preset by name.
func Get(name string) (*Preset, error) {
	p, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %q (available: %v)", name, List())
	}
	// Return a copy so callers cannot modify the built-in.
	copy := *p
	return &copy, nil
}

// List returns the sorted names of all built-in presets.
func List() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
