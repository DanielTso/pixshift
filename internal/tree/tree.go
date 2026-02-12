package tree

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
)

// Options controls the output of Print.
type Options struct {
	ShowSize   bool
	ShowFormat bool
	MaxDepth   int // 0 = unlimited
}

// Print writes a directory tree of supported image files to w.
func Print(w io.Writer, root string, opts Options) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", root)
	}
	fmt.Fprintf(w, "%s/\n", filepath.Base(root))
	return printDir(w, root, "", opts, 1)
}

// printDir recursively prints the contents of a directory.
func printDir(w io.Writer, dir, prefix string, opts Options, depth int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Filter to only supported images and directories that contain images.
	type item struct {
		name  string
		isDir bool
		size  int64
	}
	var items []item
	for _, e := range entries {
		if e.IsDir() {
			sub := filepath.Join(dir, e.Name())
			if hasImages(sub, opts.MaxDepth, depth+1) {
				items = append(items, item{name: e.Name(), isDir: true})
			}
			continue
		}
		ext := filepath.Ext(e.Name())
		if codec.IsSupportedExtension(ext) {
			var size int64
			if info, err := e.Info(); err == nil {
				size = info.Size()
			}
			items = append(items, item{name: e.Name(), size: size})
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	for i, it := range items {
		last := i == len(items)-1
		connector := "├── "
		if last {
			connector = "└── "
		}

		if it.isDir {
			fmt.Fprintf(w, "%s%s%s/\n", prefix, connector, it.name)
			if opts.MaxDepth == 0 || depth < opts.MaxDepth {
				childPrefix := prefix + "│   "
				if last {
					childPrefix = prefix + "    "
				}
				if err := printDir(w, filepath.Join(dir, it.name), childPrefix, opts, depth+1); err != nil {
					return err
				}
			}
			continue
		}

		line := prefix + connector + it.name
		if opts.ShowSize {
			line += " (" + humanSize(it.size) + ")"
		}
		if opts.ShowFormat {
			line += " [" + formatLabel(filepath.Ext(it.name)) + "]"
		}
		fmt.Fprintln(w, line)
	}
	return nil
}

// hasImages reports whether dir (or its subdirectories) contains supported images.
func hasImages(dir string, maxDepth, currentDepth int) bool {
	if maxDepth > 0 && currentDepth > maxDepth {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			if hasImages(filepath.Join(dir, e.Name()), maxDepth, currentDepth+1) {
				return true
			}
			continue
		}
		if codec.IsSupportedExtension(filepath.Ext(e.Name())) {
			return true
		}
	}
	return false
}

// humanSize formats a byte count as a human-readable string.
func humanSize(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// formatLabel returns the format label for a file extension.
func formatLabel(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "JPEG"
	case ".png":
		return "PNG"
	case ".gif":
		return "GIF"
	case ".webp":
		return "WebP"
	case ".tiff", ".tif":
		return "TIFF"
	case ".bmp":
		return "BMP"
	case ".heic", ".heif":
		return "HEIC"
	case ".avif":
		return "AVIF"
	case ".cr2":
		return "CR2"
	case ".nef":
		return "NEF"
	case ".dng":
		return "DNG"
	default:
		return strings.TrimPrefix(strings.ToUpper(ext), ".")
	}
}
