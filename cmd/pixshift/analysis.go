package main

import (
	"encoding/json"
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/contact"
	"github.com/DanielTso/pixshift/internal/dedup"
	"github.com/DanielTso/pixshift/internal/ssim"
	"github.com/DanielTso/pixshift/internal/tree"
)

func runTreeMode(opts *options) {
	dir := "."
	if len(opts.inputs) > 0 {
		dir = opts.inputs[0]
	}
	if err := tree.Print(os.Stdout, dir, tree.Options{
		ShowSize:   true,
		ShowFormat: opts.verbose,
	}); err != nil {
		fatal("tree: %v", err)
	}
}

func runSSIMMode(opts *options) {
	score, err := ssim.CompareFiles(opts.ssimFiles[0], opts.ssimFiles[1], codec.DefaultRegistry())
	if err != nil {
		fatal("ssim: %v", err)
	}
	if opts.jsonOutput {
		out := map[string]interface{}{
			"file1":  opts.ssimFiles[0],
			"file2":  opts.ssimFiles[1],
			"ssim":   score,
			"rating": ssim.Rating(score),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	} else {
		fmt.Printf("SSIM: %.4f (%s)\n", score, ssim.Rating(score))
		fmt.Printf("  %s\n  %s\n", opts.ssimFiles[0], opts.ssimFiles[1])
	}
}

func runDedupMode(reg *codec.Registry, opts *options) {
	files := collectFiles(opts.inputs, opts.recursive)
	if len(files) == 0 {
		fatal("no supported image files found")
	}

	threshold := opts.dedupThreshold
	if threshold == 0 {
		threshold = 10
	}

	type fileHash struct {
		path string
		hash uint64
	}

	hashes := make([]fileHash, 0, len(files))
	for _, f := range files {
		h, err := dedup.HashFile(f, reg)
		if err != nil {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			}
			continue
		}
		hashes = append(hashes, fileHash{path: f, hash: h})
	}

	type dupGroup struct {
		Files    []string `json:"files"`
		Distance int      `json:"distance"`
	}

	seen := make(map[int]bool)
	var groups []dupGroup

	for i := 0; i < len(hashes); i++ {
		if seen[i] {
			continue
		}
		group := []string{hashes[i].path}
		minDist := 0
		for j := i + 1; j < len(hashes); j++ {
			if seen[j] {
				continue
			}
			dist := dedup.HammingDistance(hashes[i].hash, hashes[j].hash)
			if dist <= threshold {
				seen[j] = true
				group = append(group, hashes[j].path)
				if dist > minDist {
					minDist = dist
				}
			}
		}
		if len(group) > 1 {
			seen[i] = true
			groups = append(groups, dupGroup{Files: group, Distance: minDist})
		}
	}

	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{
			"scanned":    len(hashes),
			"threshold":  threshold,
			"groups":     groups,
			"duplicates": len(groups),
		})
		return
	}

	if len(groups) == 0 {
		fmt.Printf("No duplicates found among %d images (threshold: %d).\n", len(hashes), threshold)
		return
	}

	fmt.Printf("Found %d duplicate group(s) among %d images (threshold: %d):\n\n", len(groups), len(hashes), threshold)
	for i, g := range groups {
		fmt.Printf("Group %d (distance: %d):\n", i+1, g.Distance)
		for _, f := range g.Files {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println()
	}
}

func runContactSheetMode(reg *codec.Registry, opts *options) {
	files := collectFiles(opts.inputs, opts.recursive)
	if len(files) == 0 {
		fatal("no supported image files found")
	}

	var entries []contact.ImageEntry
	for _, f := range files {
		imgFile, err := os.Open(f)
		if err != nil {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			}
			continue
		}

		format, err := codec.DetectFormat(imgFile, f)
		if err != nil {
			imgFile.Close()
			continue
		}
		_, _ = imgFile.Seek(0, 0)

		dec, err := reg.Decoder(format)
		if err != nil {
			imgFile.Close()
			continue
		}

		img, err := dec.Decode(imgFile)
		imgFile.Close()
		if err != nil {
			continue
		}

		entries = append(entries, contact.ImageEntry{
			Path:  f,
			Image: img,
		})
	}

	if len(entries) == 0 {
		fatal("no images could be decoded")
	}

	csOpts := contact.DefaultOptions()
	if opts.contactCols > 0 {
		csOpts.Columns = opts.contactCols
	}
	if opts.contactSize > 0 {
		csOpts.ThumbSize = opts.contactSize
	}

	sheet := contact.Generate(entries, csOpts)

	outPath := "contact-sheet.jpg"
	if opts.outputDir != "" {
		_ = os.MkdirAll(opts.outputDir, 0755)
		outPath = filepath.Join(opts.outputDir, outPath)
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		fatal("create contact sheet: %v", err)
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, sheet, &jpeg.Options{Quality: opts.quality}); err != nil {
		fatal("encode contact sheet: %v", err)
	}

	fmt.Printf("Contact sheet: %s (%d images, %dx%d)\n",
		outPath, len(entries), sheet.Bounds().Dx(), sheet.Bounds().Dy())
}
