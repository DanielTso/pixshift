package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/completion"
	"github.com/DanielTso/pixshift/internal/contact"
	"github.com/DanielTso/pixshift/internal/dedup"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/preset"
	"github.com/DanielTso/pixshift/internal/rules"
	"github.com/DanielTso/pixshift/internal/server"
	"github.com/DanielTso/pixshift/internal/ssim"
	"github.com/DanielTso/pixshift/internal/tree"
	"github.com/DanielTso/pixshift/internal/version"
	"github.com/DanielTso/pixshift/internal/watch"
	"github.com/schollz/progressbar/v3"
)

const defaultQuality = 92

type options struct {
	format        string
	quality       int
	jobs          int
	outputDir     string
	recursive     bool
	metadata      bool
	stripMetadata bool
	watchMode     bool
	configFile    string
	overwrite     bool
	dryRun        bool
	verbose       bool
	width         int
	height        int
	maxDim        int
	template      string
	completionSh  string
	inputs        []string

	// v0.3.0 fields
	autoRotate       bool
	cropWidth        int
	cropHeight       int
	cropRatio        string
	cropGravity      string
	watermarkText    string
	watermarkPos     string
	watermarkOpacity float64
	presetName       string
	backup           bool
	jsonOutput       bool
	treeMode         bool
	dedupMode        bool
	dedupThreshold   int
	ssimFiles        []string
	contactSheet     bool
	contactCols      int
	contactSize      int
	serveAddr        string

	// v0.4.0 fields
	grayscale      bool
	sepia          float64
	brightness     float64
	contrast       float64
	sharpen        bool
	blur           float64
	invert         bool
	progressive    bool
	pngCompression int
	webpMethod     int
	lossless       bool
	watermarkSize  float64
	watermarkColor string
	watermarkBg    string
	interpolation  string
}

func main() {
	opts := parseArgs(os.Args[1:])

	if opts == nil {
		os.Exit(0)
	}

	// Shell completion
	if opts.completionSh != "" {
		switch opts.completionSh {
		case "bash":
			fmt.Print(completion.GenerateBash())
		case "zsh":
			fmt.Print(completion.GenerateZsh())
		case "fish":
			fmt.Print(completion.GenerateFish())
		default:
			fatal("unknown shell: %s (use bash, zsh, or fish)", opts.completionSh)
		}
		os.Exit(0)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	registry := codec.DefaultRegistry()

	// Apply preset if specified
	if opts.presetName != "" {
		p, err := preset.Get(opts.presetName)
		if err != nil {
			fatal("%v", err)
		}
		if opts.format == "" {
			opts.format = p.Format
		}
		if opts.quality == defaultQuality {
			opts.quality = p.Quality
		}
		if opts.maxDim == 0 && p.MaxDim > 0 {
			opts.maxDim = p.MaxDim
		}
		if opts.width == 0 && p.Width > 0 {
			opts.width = p.Width
		}
		if opts.height == 0 && p.Height > 0 {
			opts.height = p.Height
		}
		if !opts.metadata && !opts.stripMetadata {
			opts.metadata = p.PreserveMetadata
			opts.stripMetadata = p.StripMetadata
		}
		if p.Grayscale {
			opts.grayscale = true
		}
		if p.Sharpen {
			opts.sharpen = true
		}
		if p.AutoRotate {
			opts.autoRotate = true
		}
	}

	pipe := pipeline.NewPipeline(registry)

	// Serve mode
	if opts.serveAddr != "" {
		fmt.Printf("Starting Pixshift HTTP server on %s...\n", opts.serveAddr)
		srv := server.New(registry, opts.serveAddr)
		if err := srv.Start(ctx); err != nil {
			fatal("server: %v", err)
		}
		return
	}

	// Tree mode
	if opts.treeMode {
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
		return
	}

	// SSIM comparison mode
	if len(opts.ssimFiles) == 2 {
		score, err := ssim.CompareFiles(opts.ssimFiles[0], opts.ssimFiles[1], registry)
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
		return
	}

	// Dedup mode
	if opts.dedupMode {
		runDedupMode(registry, opts)
		return
	}

	// Contact sheet mode
	if opts.contactSheet {
		runContactSheetMode(registry, opts)
		return
	}

	// Auto-discover config if not specified
	if opts.configFile == "" {
		opts.configFile = discoverConfig(opts.verbose)
	}

	// Load custom presets from config (if any) before rules mode
	if opts.configFile != "" {
		loadPresetsFromConfig(opts.configFile)
	}

	// Rules mode
	if opts.configFile != "" {
		runRulesMode(ctx, pipe, registry, opts)
		return
	}

	// Determine output format
	outputFormat := codec.JPEG
	if opts.format != "" {
		f, err := codec.ParseFormat(opts.format)
		if err != nil {
			fatal("invalid format: %v", err)
		}
		outputFormat = f
	}

	// Stdin/stdout mode
	if len(opts.inputs) == 1 && opts.inputs[0] == "-" {
		runStdinMode(pipe, registry, outputFormat, opts)
		return
	}

	// Watch mode
	if opts.watchMode {
		runWatchMode(ctx, pipe, outputFormat, opts)
		return
	}

	// Batch mode
	runBatchMode(ctx, pipe, registry, outputFormat, opts)
}

func runStdinMode(pipe *pipeline.Pipeline, reg *codec.Registry, outputFormat codec.Format, opts *options) {
	// Buffer stdin to a temp file for seeking
	tmpIn, err := os.CreateTemp("", "pixshift-stdin-*")
	if err != nil {
		fatal("create temp file: %v", err)
	}
	defer os.Remove(tmpIn.Name())
	defer tmpIn.Close()

	if _, err := io.Copy(tmpIn, os.Stdin); err != nil {
		fatal("read stdin: %v", err)
	}
	tmpIn.Close()

	// Create temp output
	tmpOut, err := os.CreateTemp("", "pixshift-out-*"+codec.DefaultExtension(outputFormat))
	if err != nil {
		fatal("create temp output: %v", err)
	}
	defer os.Remove(tmpOut.Name())
	tmpOut.Close()

	job := buildJob(opts, tmpIn.Name(), "", outputFormat, "")
	job.OutputPath = tmpOut.Name()

	if _, _, err := pipe.Execute(job); err != nil {
		fatal("convert: %v", err)
	}

	// Write output to stdout
	outFile, err := os.Open(tmpOut.Name())
	if err != nil {
		fatal("read output: %v", err)
	}
	defer outFile.Close()
	_, _ = io.Copy(os.Stdout, outFile)
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

func runBatchMode(ctx context.Context, pipe *pipeline.Pipeline, reg *codec.Registry, outputFormat codec.Format, opts *options) {
	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0755); err != nil {
			fatal("cannot create output directory: %v", err)
		}
	}

	files := collectFiles(opts.inputs, opts.recursive)
	if len(files) == 0 {
		fatal("no supported image files found")
	}

	// Determine base directories for relative path preservation
	baseDirs := resolveBaseDirs(opts.inputs)

	var jobs []pipeline.Job
	for _, f := range files {
		inputFormat, err := detectFileFormat(f)
		if err != nil {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			}
			continue
		}

		// Skip if input and output format are the same
		if inputFormat == outputFormat {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: already %s\n", f, outputFormat)
			}
			continue
		}

		outPath := buildOutputPath(f, opts.outputDir, outputFormat, opts.template, baseDirs, opts.recursive)

		if !opts.overwrite {
			if _, err := os.Stat(outPath); err == nil {
				if opts.verbose {
					fmt.Fprintf(os.Stderr, "skip %s: output exists (use --overwrite)\n", f)
				}
				continue
			}
		}

		job := buildJob(opts, f, outPath, outputFormat, inputFormat)
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		fmt.Println("Nothing to convert.")
		return
	}

	if opts.dryRun {
		if opts.jsonOutput {
			items := make([]map[string]string, len(jobs))
			for i, j := range jobs {
				items[i] = map[string]string{
					"input":  j.InputPath,
					"output": j.OutputPath,
					"format": string(j.OutputFormat),
				}
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(items)
		} else {
			for _, j := range jobs {
				fmt.Printf("[dry-run] %s -> %s\n", j.InputPath, j.OutputPath)
			}
			fmt.Printf("\n%d file(s) would be converted.\n", len(jobs))
		}
		return
	}

	pool := pipeline.NewPool(pipe, opts.jobs)

	var succeeded, failed int
	var totalInputSize, totalOutputSize int64
	var jsonResults []map[string]interface{}

	if opts.jsonOutput {
		jsonResults = make([]map[string]interface{}, 0, len(jobs))
	}

	if !opts.verbose && !opts.jsonOutput && len(jobs) > 1 {
		// Progress bar mode for non-verbose batch
		bar := progressbar.NewOptions(len(jobs),
			progressbar.OptionSetDescription("Converting"),
			progressbar.OptionShowCount(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetElapsedTime(true),
			progressbar.OptionClearOnFinish(),
		)

		pool.RunWithCallback(ctx, jobs, func(r pipeline.Result, completed, total int) {
			if r.Error != nil {
				failed++
			} else {
				succeeded++
				totalInputSize += r.InputSize
				totalOutputSize += r.OutputSize
			}
			_ = bar.Add(1)
		})
		_ = bar.Finish()
		fmt.Println()
	} else {
		// Verbose, JSON, or single-file mode
		pool.RunWithCallback(ctx, jobs, func(r pipeline.Result, completed, total int) {
			if r.Error != nil {
				failed++
				if opts.jsonOutput {
					jsonResults = append(jsonResults, map[string]interface{}{
						"input":  r.Job.InputPath,
						"error":  r.Error.Error(),
						"status": "failed",
					})
				} else {
					fmt.Fprintf(os.Stderr, "[%d/%d] FAIL %s: %v\n", completed, total, r.Job.InputPath, r.Error)
				}
			} else {
				succeeded++
				totalInputSize += r.InputSize
				totalOutputSize += r.OutputSize
				if opts.jsonOutput {
					jsonResults = append(jsonResults, map[string]interface{}{
						"input":       r.Job.InputPath,
						"output":      r.Job.OutputPath,
						"input_size":  r.InputSize,
						"output_size": r.OutputSize,
						"status":      "ok",
					})
				} else {
					fmt.Printf("[%d/%d] %s (%s) -> %s (%s) [%s]\n",
						completed, total,
						r.Job.InputPath, humanSize(r.InputSize),
						r.Job.OutputPath, humanSize(r.OutputSize),
						sizeRatio(r.InputSize, r.OutputSize))
				}
			}
		})
	}

	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]interface{}{
			"converted":    succeeded,
			"failed":       failed,
			"total_input":  totalInputSize,
			"total_output": totalOutputSize,
			"files":        jsonResults,
		})
	} else {
		fmt.Printf("\nDone. %d converted, %d failed.", succeeded, failed)
		if totalInputSize > 0 && succeeded > 0 {
			fmt.Printf(" Total: %s -> %s (%s)",
				humanSize(totalInputSize), humanSize(totalOutputSize),
				sizeRatio(totalInputSize, totalOutputSize))
		}
		fmt.Println()
	}
	if failed > 0 {
		os.Exit(1)
	}
}

func runWatchMode(ctx context.Context, pipe *pipeline.Pipeline, outputFormat codec.Format, opts *options) {
	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0755); err != nil {
			fatal("cannot create output directory: %v", err)
		}
	}

	dirs := opts.inputs
	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	fmt.Printf("Watching for new images (output: %s)...\n", outputFormat)
	fmt.Println("Press Ctrl+C to stop.")

	w := &watch.Watcher{
		Pipeline:     pipe,
		OutputDir:    opts.outputDir,
		OutputFormat: outputFormat,
		Quality:      opts.quality,
		Recursive:    opts.recursive,
		Metadata:     opts.metadata,
		Verbose:      opts.verbose,
		OnConvert: func(r pipeline.Result) {
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", r.Job.InputPath, r.Error)
			} else {
				fmt.Printf("OK   %s (%s) -> %s (%s) [%s]\n",
					r.Job.InputPath, humanSize(r.InputSize),
					r.Job.OutputPath, humanSize(r.OutputSize),
					sizeRatio(r.InputSize, r.OutputSize))
			}
		},
	}

	if err := w.Watch(ctx, dirs); err != nil {
		fatal("watch: %v", err)
	}
}

func runRulesMode(ctx context.Context, pipe *pipeline.Pipeline, reg *codec.Registry, opts *options) {
	cfg, err := rules.LoadConfig(opts.configFile)
	if err != nil {
		fatal("load config: %v", err)
	}

	parsed, err := rules.ParseRules(cfg)
	if err != nil {
		fatal("parse rules: %v", err)
	}

	engine := rules.NewEngine(parsed)
	engine.OutputDir = opts.outputDir
	engine.Quality = opts.quality
	engine.Metadata = opts.metadata

	if opts.outputDir != "" {
		if err := os.MkdirAll(opts.outputDir, 0755); err != nil {
			fatal("cannot create output directory: %v", err)
		}
	}

	// Watch mode with rules
	if opts.watchMode {
		dirs := opts.inputs
		if len(dirs) == 0 {
			dirs = []string{"."}
		}

		fmt.Println("Watching with rules engine...")
		fmt.Println("Press Ctrl+C to stop.")

		runRulesWatch(ctx, pipe, engine, dirs, opts)
		return
	}

	// Batch mode with rules
	files := collectFiles(opts.inputs, opts.recursive)
	if len(files) == 0 {
		fatal("no supported image files found")
	}

	var jobs []pipeline.Job
	for _, f := range files {
		inputFormat, err := detectFileFormat(f)
		if err != nil {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: %v\n", f, err)
			}
			continue
		}

		job := engine.Match(f, inputFormat)
		if job == nil {
			if opts.verbose {
				fmt.Fprintf(os.Stderr, "skip %s: no matching rule\n", f)
			}
			continue
		}

		// Apply resize, transform, and strip settings from CLI
		applyOptsToJob(opts, job)

		if !opts.overwrite {
			if _, err := os.Stat(job.OutputPath); err == nil {
				if opts.verbose {
					fmt.Fprintf(os.Stderr, "skip %s: output exists\n", f)
				}
				continue
			}
		}

		jobs = append(jobs, *job)
	}

	if len(jobs) == 0 {
		fmt.Println("Nothing to convert.")
		return
	}

	if opts.dryRun {
		for _, j := range jobs {
			fmt.Printf("[dry-run] %s -> %s (%s)\n", j.InputPath, j.OutputPath, j.OutputFormat)
		}
		fmt.Printf("\n%d file(s) would be converted.\n", len(jobs))
		return
	}

	pool := pipeline.NewPool(pipe, opts.jobs)

	var succeeded, failed int
	var totalInputSize, totalOutputSize int64
	pool.RunWithCallback(ctx, jobs, func(r pipeline.Result, completed, total int) {
		if r.Error != nil {
			failed++
			fmt.Fprintf(os.Stderr, "[%d/%d] FAIL %s: %v\n", completed, total, r.Job.InputPath, r.Error)
		} else {
			succeeded++
			totalInputSize += r.InputSize
			totalOutputSize += r.OutputSize
			fmt.Printf("[%d/%d] %s (%s) -> %s (%s) [%s]\n",
				completed, total,
				r.Job.InputPath, humanSize(r.InputSize),
				r.Job.OutputPath, humanSize(r.OutputSize),
				sizeRatio(r.InputSize, r.OutputSize))
		}
	})

	fmt.Printf("\nDone. %d converted, %d failed.", succeeded, failed)
	if totalInputSize > 0 && succeeded > 0 {
		fmt.Printf(" Total: %s -> %s (%s)",
			humanSize(totalInputSize), humanSize(totalOutputSize),
			sizeRatio(totalInputSize, totalOutputSize))
	}
	fmt.Println()
	if failed > 0 {
		os.Exit(1)
	}
}

func runRulesWatch(ctx context.Context, pipe *pipeline.Pipeline, engine *rules.Engine, dirs []string, opts *options) {
	w := &watch.Watcher{
		Pipeline:     pipe,
		OutputDir:    opts.outputDir,
		OutputFormat: codec.JPEG, // Default, but overridden by engine per file
		Quality:      opts.quality,
		Recursive:    opts.recursive,
		Metadata:     opts.metadata,
		Verbose:      opts.verbose,
		OnConvert: func(r pipeline.Result) {
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", r.Job.InputPath, r.Error)
			} else {
				fmt.Printf("OK   %s (%s) -> %s (%s) [%s]\n",
					r.Job.InputPath, humanSize(r.InputSize),
					r.Job.OutputPath, humanSize(r.OutputSize),
					sizeRatio(r.InputSize, r.OutputSize))
			}
		},
	}

	if err := w.Watch(ctx, dirs); err != nil {
		fatal("watch: %v", err)
	}
}

func parseArgs(args []string) *options {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	opts := &options{
		quality:          defaultQuality,
		jobs:             runtime.NumCPU(),
		watermarkOpacity: 0.5,
		dedupThreshold:   10,
		contactCols:      4,
		contactSize:      200,
	}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "-f", "--format":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.format = args[i+1]
			i += 2
		case "-q", "--quality":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			q, err := strconv.Atoi(args[i+1])
			if err != nil || q < 1 || q > 100 {
				fatal("quality must be an integer between 1 and 100")
			}
			opts.quality = q
			i += 2
		case "-j", "--jobs":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			j, err := strconv.Atoi(args[i+1])
			if err != nil || j < 1 {
				fatal("jobs must be a positive integer")
			}
			opts.jobs = j
			i += 2
		case "-o", "--output":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.outputDir = args[i+1]
			i += 2
		case "-c", "--config":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.configFile = args[i+1]
			i += 2
		case "-r", "--recursive":
			opts.recursive = true
			i++
		case "-m", "--preserve-metadata":
			opts.metadata = true
			i++
		case "-s", "--strip-metadata":
			opts.stripMetadata = true
			i++
		case "-w", "--watch":
			opts.watchMode = true
			i++
		case "--overwrite":
			opts.overwrite = true
			i++
		case "--dry-run":
			opts.dryRun = true
			i++
		case "-v", "--verbose":
			opts.verbose = true
			i++
		case "--width":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			w, err := strconv.Atoi(args[i+1])
			if err != nil || w < 1 {
				fatal("width must be a positive integer")
			}
			opts.width = w
			i += 2
		case "--height":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			h, err := strconv.Atoi(args[i+1])
			if err != nil || h < 1 {
				fatal("height must be a positive integer")
			}
			opts.height = h
			i += 2
		case "--max-dim":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			d, err := strconv.Atoi(args[i+1])
			if err != nil || d < 1 {
				fatal("max-dim must be a positive integer")
			}
			opts.maxDim = d
			i += 2
		case "--template":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.template = args[i+1]
			i += 2
		case "--completion":
			if i+1 >= len(args) {
				fatal("missing value for %s (bash, zsh, or fish)", args[i])
			}
			opts.completionSh = args[i+1]
			i += 2
		case "--auto-rotate":
			opts.autoRotate = true
			i++
		case "--crop":
			if i+1 >= len(args) {
				fatal("missing value for %s (WxH, e.g. 800x600)", args[i])
			}
			parts := strings.SplitN(args[i+1], "x", 2)
			if len(parts) != 2 {
				fatal("crop must be WxH (e.g. 800x600)")
			}
			cw, err1 := strconv.Atoi(parts[0])
			ch, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || cw < 1 || ch < 1 {
				fatal("crop dimensions must be positive integers")
			}
			opts.cropWidth = cw
			opts.cropHeight = ch
			i += 2
		case "--crop-ratio":
			if i+1 >= len(args) {
				fatal("missing value for %s (e.g. 16:9)", args[i])
			}
			opts.cropRatio = args[i+1]
			i += 2
		case "--crop-gravity":
			if i+1 >= len(args) {
				fatal("missing value for %s (center, north, south, east, west)", args[i])
			}
			opts.cropGravity = args[i+1]
			i += 2
		case "--watermark":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkText = args[i+1]
			i += 2
		case "--watermark-pos":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkPos = args[i+1]
			i += 2
		case "--watermark-opacity":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			o, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || o < 0 || o > 1 {
				fatal("watermark-opacity must be a number between 0 and 1")
			}
			opts.watermarkOpacity = o
			i += 2
		case "--preset":
			if i+1 >= len(args) {
				fatal("missing value for %s (available: %v)", args[i], preset.List())
			}
			opts.presetName = args[i+1]
			i += 2
		case "--backup":
			opts.backup = true
			i++
		case "--json":
			opts.jsonOutput = true
			i++
		case "--tree":
			opts.treeMode = true
			i++
		case "--dedup":
			opts.dedupMode = true
			i++
		case "--dedup-threshold":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			t, err := strconv.Atoi(args[i+1])
			if err != nil || t < 0 {
				fatal("dedup-threshold must be a non-negative integer")
			}
			opts.dedupThreshold = t
			i += 2
		case "--ssim":
			if i+2 >= len(args) {
				fatal("--ssim requires two file arguments")
			}
			opts.ssimFiles = []string{args[i+1], args[i+2]}
			i += 3
		case "--contact-sheet":
			opts.contactSheet = true
			i++
		case "--contact-cols":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			c, err := strconv.Atoi(args[i+1])
			if err != nil || c < 1 {
				fatal("contact-cols must be a positive integer")
			}
			opts.contactCols = c
			i += 2
		case "--contact-size":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			s, err := strconv.Atoi(args[i+1])
			if err != nil || s < 1 {
				fatal("contact-size must be a positive integer")
			}
			opts.contactSize = s
			i += 2
		case "--grayscale":
			opts.grayscale = true
			i++
		case "--sepia":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				s, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || s < 0 || s > 1 {
					fatal("sepia intensity must be a number between 0 and 1")
				}
				opts.sepia = s
				i += 2
			} else {
				opts.sepia = 0.8
				i++
			}
		case "--brightness":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			b, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || b < -100 || b > 100 {
				fatal("brightness must be a number between -100 and 100")
			}
			opts.brightness = b
			i += 2
		case "--contrast":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			c, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || c < -100 || c > 100 {
				fatal("contrast must be a number between -100 and 100")
			}
			opts.contrast = c
			i += 2
		case "--sharpen":
			opts.sharpen = true
			i++
		case "--blur":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			bl, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || bl < 0 {
				fatal("blur radius must be a non-negative number")
			}
			opts.blur = bl
			i += 2
		case "--invert":
			opts.invert = true
			i++
		case "--progressive":
			opts.progressive = true
			i++
		case "--png-compression":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			pc, err := strconv.Atoi(args[i+1])
			if err != nil || pc < 0 || pc > 3 {
				fatal("png-compression must be 0-3")
			}
			opts.pngCompression = pc
			i += 2
		case "--webp-method":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			wm, err := strconv.Atoi(args[i+1])
			if err != nil || wm < 0 || wm > 6 {
				fatal("webp-method must be 0-6")
			}
			opts.webpMethod = wm
			i += 2
		case "--lossless":
			opts.lossless = true
			i++
		case "--watermark-size":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			ws, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || ws <= 0 {
				fatal("watermark-size must be a positive number")
			}
			opts.watermarkSize = ws
			i += 2
		case "--watermark-color":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkColor = args[i+1]
			i += 2
		case "--watermark-bg":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkBg = args[i+1]
			i += 2
		case "--interpolation":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			interp := strings.ToLower(args[i+1])
			if interp != "nearest" && interp != "bilinear" && interp != "catmullrom" {
				fatal("interpolation must be nearest, bilinear, or catmullrom")
			}
			opts.interpolation = interp
			i += 2
		case "serve":
			addr := ":8080"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				addr = args[i+1]
				i += 2
			} else {
				i++
			}
			opts.serveAddr = addr
		case "-V", "--version":
			fmt.Println(version.String())
			return nil
		case "-h", "--help":
			printUsage()
			return nil
		default:
			if strings.HasPrefix(args[i], "-") {
				fatal("unknown flag: %s", args[i])
			}
			opts.inputs = append(opts.inputs, args[i])
			i++
		}
	}

	if opts.metadata && opts.stripMetadata {
		fatal("--preserve-metadata and --strip-metadata are mutually exclusive")
	}

	needsInput := !opts.watchMode && opts.configFile == "" && opts.completionSh == "" &&
		opts.serveAddr == "" && len(opts.ssimFiles) == 0
	if len(opts.inputs) == 0 && needsInput {
		fatal("no input files or directories specified")
	}

	return opts
}

func collectFiles(inputs []string, recursive bool) []string {
	var files []string
	for _, input := range inputs {
		matches, err := filepath.Glob(input)
		if err != nil || len(matches) == 0 {
			matches = []string{input}
		}
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: cannot access %s: %v\n", match, err)
				continue
			}
			if info.IsDir() {
				dirFiles := collectDir(match, recursive)
				files = append(files, dirFiles...)
			} else {
				if codec.IsSupportedExtension(filepath.Ext(match)) {
					files = append(files, match)
				}
			}
		}
	}
	return files
}

func collectDir(dir string, recursive bool) []string {
	var files []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && !recursive && path != dir {
			return filepath.SkipDir
		}
		if !info.IsDir() && codec.IsSupportedExtension(filepath.Ext(path)) {
			files = append(files, path)
		}
		return nil
	}
	_ = filepath.Walk(dir, walkFn)
	return files
}

func detectFileFormat(path string) (codec.Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return codec.DetectFormat(f, path)
}

// resolveBaseDirs returns the base directories from the input paths,
// used for preserving directory structure with -o.
func resolveBaseDirs(inputs []string) map[string]string {
	bases := make(map[string]string)
	for _, input := range inputs {
		abs, err := filepath.Abs(input)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil {
			continue
		}
		if info.IsDir() {
			bases[abs] = abs
		} else {
			bases[abs] = filepath.Dir(abs)
		}
	}
	return bases
}

func buildOutputPath(inputPath, outputDir string, format codec.Format, tmpl string, baseDirs map[string]string, recursive bool) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	// Apply template if set
	var outName string
	if tmpl != "" {
		outName = tmpl
		outName = strings.ReplaceAll(outName, "{name}", name)
		outName = strings.ReplaceAll(outName, "{ext}", strings.TrimPrefix(ext, "."))
		outName = strings.ReplaceAll(outName, "{format}", string(format))
		// Ensure it has the right extension
		if filepath.Ext(outName) == "" {
			outName += codec.DefaultExtension(format)
		}
	} else {
		outName = name + codec.DefaultExtension(format)
	}

	if outputDir != "" {
		// Preserve directory structure when using -o with -r
		if recursive {
			absInput, err := filepath.Abs(inputPath)
			if err == nil {
				for _, baseDir := range baseDirs {
					rel, err := filepath.Rel(baseDir, filepath.Dir(absInput))
					if err == nil && !strings.HasPrefix(rel, "..") {
						subDir := filepath.Join(outputDir, rel)
						_ = os.MkdirAll(subDir, 0755)
						return filepath.Join(subDir, outName)
					}
				}
			}
		}
		return filepath.Join(outputDir, outName)
	}
	return filepath.Join(filepath.Dir(inputPath), outName)
}

// discoverConfig looks for a config file in standard locations.
func discoverConfig(verbose bool) string {
	candidates := []string{
		"pixshift.yaml",
		"pixshift.yml",
	}

	// Add home config paths
	home, err := os.UserHomeDir()
	if err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".config", "pixshift", "config.yaml"),
			filepath.Join(home, ".config", "pixshift", "config.yml"),
		)
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "auto-discovered config: %s\n", c)
			}
			return c
		}
	}
	return ""
}

// humanSize formats bytes into human-readable size.
func humanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// sizeRatio returns a human-readable comparison string.
func sizeRatio(input, output int64) string {
	if input == 0 {
		return "n/a"
	}
	pct := float64(input-output) / float64(input) * 100
	if pct > 0 {
		return fmt.Sprintf("%.0f%% smaller", pct)
	} else if pct < 0 {
		return fmt.Sprintf("%.0f%% larger", -pct)
	}
	return "same size"
}

func printUsage() {
	fmt.Print(`pixshift - Universal image converter

Usage:
  pixshift [options] <files or directories...>
  pixshift serve [addr]
  pixshift --tree [dir]
  pixshift --dedup [dir]
  pixshift --ssim <file1> <file2>
  pixshift --contact-sheet [dir]

Conversion options:
  -f, --format <fmt>        Output format: jpg, png, gif, webp, tiff, bmp, heic, avif
  -q, --quality <1-100>     Encoding quality (default: 92)
  -j, --jobs <N>            Parallel workers (default: number of CPUs)
  -o, --output <dir>        Output directory (default: same as input)
  -r, --recursive           Process directories recursively
  -m, --preserve-metadata   Preserve EXIF metadata
  -s, --strip-metadata      Strip all EXIF/GPS metadata from output
  -w, --watch               Watch mode: auto-convert new files
  -c, --config <file>       Rules mode: use YAML config file
      --overwrite           Overwrite existing output files
      --dry-run             Preview what would happen
      --template <pattern>  Output naming template (placeholders: {name}, {ext}, {format})
      --preset <name>       Named preset: web, thumbnail, print, archive

Image transforms:
      --auto-rotate          Auto-rotate based on EXIF orientation
      --crop <WxH>           Crop to exact pixel dimensions (e.g. 800x600)
      --crop-ratio <W:H>     Crop to aspect ratio (e.g. 16:9)
      --crop-gravity <pos>   Crop anchor: center, north, south, east, west
      --watermark <text>     Add text watermark
      --watermark-pos <pos>  Watermark position: bottom-right, bottom-left, top-right, top-left, center
      --watermark-opacity <N> Watermark opacity 0.0-1.0 (default: 0.5)
      --watermark-size <N>   Watermark font scale (default: 1.0, e.g. 3.0 = 3x)
      --watermark-color <hex> Watermark text color (default: #FFFFFF)
      --watermark-bg <hex>   Watermark background color (default: #000000)

Filters:
      --grayscale            Convert to grayscale
      --sepia [intensity]    Apply sepia tone (default: 0.8, range: 0.0-1.0)
      --brightness <N>       Adjust brightness (-100 to +100)
      --contrast <N>         Adjust contrast (-100 to +100)
      --sharpen              Apply sharpen filter
      --blur <radius>        Apply blur filter (radius in pixels)
      --invert               Invert colors

Resize:
      --width <N>           Target width (preserves aspect ratio)
      --height <N>          Target height (preserves aspect ratio)
      --max-dim <N>         Max dimension (scale to fit)
      --interpolation <m>   Resize method: nearest, bilinear, catmullrom (default)

Encoding options:
      --progressive          JPEG progressive encoding (reserved for future encoder)
      --png-compression <N>  PNG compression: 0=default, 1=none, 2=fast, 3=best
      --webp-method <N>     WebP encoding method: 0-6 (0=fast, 6=best)
      --lossless             WebP lossless mode

Analysis tools:
      --tree                Show directory tree of supported images
      --dedup               Find duplicate images using perceptual hashing
      --dedup-threshold <N> Hamming distance threshold (default: 10)
      --ssim <f1> <f2>      Compare two images using Structural Similarity Index
      --contact-sheet       Generate a contact sheet (thumbnail grid)
      --contact-cols <N>    Contact sheet columns (default: 4)
      --contact-size <N>    Contact sheet thumbnail size in px (default: 200)

Other:
      --backup              Create .bak backup of originals before converting
      --json                Output results as JSON
      --completion <shell>  Generate shell completion (bash, zsh, fish)
  -v, --verbose             Verbose output
  -V, --version             Show version
  -h, --help                Show this help

Serve mode:
  pixshift serve [addr]     Start HTTP conversion server (default: :8080)
    POST /convert           Upload image with multipart form (file, format, quality)
    GET  /formats           List supported formats
    GET  /health            Health check

Stdin/stdout:
  cat photo.heic | pixshift -f webp - > photo.webp

Presets:
  web         WebP, q85, max 1920px, strip metadata
  thumbnail   JPEG, q80, max 300px, strip metadata
  print       TIFF, q100, preserve metadata
  archive     PNG, q100, preserve metadata
  (custom presets can be defined in config YAML under "presets:" section)

Examples:
  pixshift photo.heic                            Convert HEIC to JPEG (default)
  pixshift -f webp -q 90 photo.heic             Convert to WebP at quality 90
  pixshift -f png -o converted/ photos/          Batch convert directory to PNG
  pixshift --preset web -o output/ photos/       Use web preset
  pixshift --auto-rotate -f jpg photo.heic       Auto-rotate from EXIF
  pixshift --crop-ratio 16:9 -f webp photo.jpg   Crop to 16:9
  pixshift --watermark "© 2026" -f jpg photos/   Add watermark
  pixshift --grayscale --sharpen -f webp photo.jpg Apply filters
  pixshift --watermark "Test" --watermark-size 3 --watermark-color "#FF0000" -f jpg photo.jpg
  pixshift --png-compression 3 -f png photo.jpg  Best PNG compression
  pixshift --lossless -f webp photo.jpg           Lossless WebP
  pixshift --tree ~/Pictures                     Show image directory tree
  pixshift --dedup ~/Pictures                    Find duplicate images
  pixshift --ssim original.jpg compressed.jpg    Compare image quality
  pixshift --contact-sheet -o output/ photos/    Generate contact sheet
  pixshift --json -f webp photos/                JSON output for scripting
  pixshift --backup -f webp photos/              Backup originals first
  pixshift serve :9090                           Start HTTP server on port 9090
  cat photo.heic | pixshift -f webp - > out.webp Stdin/stdout pipeline
`)
}

// loadPresetsFromConfig reads custom presets from a config file.
func loadPresetsFromConfig(path string) {
	cfg, err := rules.LoadConfig(path)
	if err != nil {
		return // Non-fatal: config may not be readable for presets only
	}
	if len(cfg.Presets) > 0 {
		presetMap := make(map[string]*preset.Preset, len(cfg.Presets))
		for name, pc := range cfg.Presets {
			presetMap[name] = &preset.Preset{
				Name:             name,
				Format:           pc.Format,
				Quality:          pc.Quality,
				MaxDim:           pc.MaxDim,
				Width:            pc.Width,
				Height:           pc.Height,
				StripMetadata:    pc.StripMetadata,
				PreserveMetadata: pc.PreserveMetadata,
				Grayscale:        pc.Grayscale,
				Sharpen:          pc.Sharpen,
				AutoRotate:       pc.AutoRotate,
			}
		}
		preset.LoadCustomPresets(presetMap)
	}
}

// buildJob creates a pipeline.Job from CLI options.
func buildJob(opts *options, inputPath, outputPath string, outputFormat codec.Format, inputFormat codec.Format) pipeline.Job {
	return pipeline.Job{
		InputPath:        inputPath,
		OutputPath:       outputPath,
		InputFormat:      inputFormat,
		OutputFormat:     outputFormat,
		Quality:          opts.quality,
		PreserveMetadata: opts.metadata,
		StripMetadata:    opts.stripMetadata,
		Width:            opts.width,
		Height:           opts.height,
		MaxDim:           opts.maxDim,
		AutoRotate:       opts.autoRotate,
		CropWidth:        opts.cropWidth,
		CropHeight:       opts.cropHeight,
		CropAspectRatio:  opts.cropRatio,
		CropGravity:      opts.cropGravity,
		WatermarkText:    opts.watermarkText,
		WatermarkPos:     opts.watermarkPos,
		WatermarkOpacity: opts.watermarkOpacity,
		WatermarkSize:    opts.watermarkSize,
		WatermarkColor:   opts.watermarkColor,
		WatermarkBg:      opts.watermarkBg,
		BackupOriginal:   opts.backup,
		Grayscale:        opts.grayscale,
		Sepia:            opts.sepia,
		Brightness:       opts.brightness,
		Contrast:         opts.contrast,
		Sharpen:          opts.sharpen,
		Blur:             opts.blur,
		Invert:           opts.invert,
		Interpolation:    opts.interpolation,
		EncodeOpts: codec.EncodeOptions{
			Quality:     opts.quality,
			Progressive: opts.progressive,
			Compression: opts.pngCompression,
			WebPMethod:  opts.webpMethod,
			Lossless:    opts.lossless,
		},
	}
}

// applyOptsToJob applies CLI options to a job (used in rules mode where job is pre-built).
func applyOptsToJob(opts *options, job *pipeline.Job) {
	job.Width = opts.width
	job.Height = opts.height
	job.MaxDim = opts.maxDim
	job.StripMetadata = opts.stripMetadata
	job.AutoRotate = opts.autoRotate
	job.CropWidth = opts.cropWidth
	job.CropHeight = opts.cropHeight
	job.CropAspectRatio = opts.cropRatio
	job.CropGravity = opts.cropGravity
	job.WatermarkText = opts.watermarkText
	job.WatermarkPos = opts.watermarkPos
	job.WatermarkOpacity = opts.watermarkOpacity
	job.WatermarkSize = opts.watermarkSize
	job.WatermarkColor = opts.watermarkColor
	job.WatermarkBg = opts.watermarkBg
	job.BackupOriginal = opts.backup
	job.Grayscale = opts.grayscale
	job.Sepia = opts.sepia
	job.Brightness = opts.brightness
	job.Contrast = opts.contrast
	job.Sharpen = opts.sharpen
	job.Blur = opts.blur
	job.Invert = opts.invert
	job.Interpolation = opts.interpolation
	job.EncodeOpts = codec.EncodeOptions{
		Quality:     opts.quality,
		Progressive: opts.progressive,
		Compression: opts.pngCompression,
		WebPMethod:  opts.webpMethod,
		Lossless:    opts.lossless,
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
