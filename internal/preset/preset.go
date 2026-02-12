package preset

import (
	"fmt"
	"sort"
)

// Preset defines a named conversion configuration.
type Preset struct {
	Name             string `yaml:"name,omitempty"`
	Format           string `yaml:"format,omitempty"`
	Quality          int    `yaml:"quality,omitempty"`
	MaxDim           int    `yaml:"max_dim,omitempty"`
	Width            int    `yaml:"width,omitempty"`
	Height           int    `yaml:"height,omitempty"`
	StripMetadata    bool   `yaml:"strip_metadata,omitempty"`
	PreserveMetadata bool   `yaml:"preserve_metadata,omitempty"`
	// v0.4.0
	Grayscale  bool `yaml:"grayscale,omitempty"`
	Sharpen    bool `yaml:"sharpen,omitempty"`
	AutoRotate bool `yaml:"auto_rotate,omitempty"`
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

// custom holds user-defined presets loaded from config.
var custom = map[string]*Preset{}

// Get returns a preset by name. Custom presets override built-ins.
func Get(name string) (*Preset, error) {
	if p, ok := custom[name]; ok {
		copy := *p
		return &copy, nil
	}
	p, ok := builtins[name]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %q (available: %v)", name, List())
	}
	// Return a copy so callers cannot modify the built-in.
	copy := *p
	return &copy, nil
}

// List returns the sorted names of all available presets (built-in + custom).
func List() []string {
	seen := make(map[string]bool)
	for name := range builtins {
		seen[name] = true
	}
	for name := range custom {
		seen[name] = true
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// LoadCustomPresets registers custom presets. Custom presets override built-ins.
func LoadCustomPresets(presets map[string]*Preset) {
	custom = make(map[string]*Preset, len(presets))
	for name, p := range presets {
		cp := *p
		cp.Name = name
		custom[name] = &cp
	}
}

// ResetCustomPresets clears all custom presets (useful for testing).
func ResetCustomPresets() {
	custom = map[string]*Preset{}
}
