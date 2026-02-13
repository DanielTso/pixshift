package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
)

type formatStats struct {
	Format string `json:"format"`
	Count  int    `json:"count"`
	Size   int64  `json:"size"`
}

func runScanMode(opts *options) {
	dir := "."
	if len(opts.inputs) > 0 {
		dir = opts.inputs[0]
	}

	files := collectFiles([]string{dir}, opts.recursive)
	if len(files) == 0 {
		fmt.Printf("No supported images found in %s\n", dir)
		return
	}

	// Group by format
	stats := make(map[string]*formatStats)
	var totalSize int64

	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		size := info.Size()
		totalSize += size

		format, err := detectFileFormat(f)
		if err != nil {
			format = codec.Format("unknown")
		}
		fmtStr := string(format)
		if s, ok := stats[fmtStr]; ok {
			s.Count++
			s.Size += size
		} else {
			stats[fmtStr] = &formatStats{Format: fmtStr, Count: 1, Size: size}
		}
	}

	// Sort by count descending
	sorted := make([]formatStats, 0, len(stats))
	for _, s := range stats {
		sorted = append(sorted, *s)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})

	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{
			"directory":  dir,
			"total":      len(files),
			"total_size": totalSize,
			"formats":    sorted,
		})
		return
	}

	fmt.Printf("Scanning %s...\n\n", dir)
	fmt.Printf("Found %d images:\n", len(files))
	for _, s := range sorted {
		fmt.Printf("  %4d %-6s (%s)\n", s.Count, strings.ToUpper(s.Format), humanSize(s.Size))
	}
	fmt.Printf("  %s\n", strings.Repeat("\u2501", 28))
	fmt.Printf("  %4d Total  (%s)\n", len(files), humanSize(totalSize))
}
