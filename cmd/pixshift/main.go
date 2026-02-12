package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/completion"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/rules"
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
	pipe := pipeline.NewPipeline(registry)

	// Auto-discover config if not specified
	if opts.configFile == "" {
		opts.configFile = discoverConfig(opts.verbose)
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

	// Watch mode
	if opts.watchMode {
		runWatchMode(ctx, pipe, outputFormat, opts)
		return
	}

	// Batch mode
	runBatchMode(ctx, pipe, registry, outputFormat, opts)
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

		jobs = append(jobs, pipeline.Job{
			InputPath:        f,
			OutputPath:       outPath,
			InputFormat:      inputFormat,
			OutputFormat:     outputFormat,
			Quality:          opts.quality,
			PreserveMetadata: opts.metadata,
			StripMetadata:    opts.stripMetadata,
			Width:            opts.width,
			Height:           opts.height,
			MaxDim:           opts.maxDim,
		})
	}

	if len(jobs) == 0 {
		fmt.Println("Nothing to convert.")
		return
	}

	if opts.dryRun {
		for _, j := range jobs {
			fmt.Printf("[dry-run] %s -> %s\n", j.InputPath, j.OutputPath)
		}
		fmt.Printf("\n%d file(s) would be converted.\n", len(jobs))
		return
	}

	pool := pipeline.NewPool(pipe, opts.jobs)

	var succeeded, failed int
	var totalInputSize, totalOutputSize int64

	if !opts.verbose && len(jobs) > 1 {
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
		// Verbose or single-file mode
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
	}

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

		// Apply resize and strip settings
		job.Width = opts.width
		job.Height = opts.height
		job.MaxDim = opts.maxDim
		job.StripMetadata = opts.stripMetadata

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
		quality: defaultQuality,
		jobs:    runtime.NumCPU(),
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

	if len(opts.inputs) == 0 && !opts.watchMode && opts.configFile == "" && opts.completionSh == "" {
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

Options:
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
      --width <N>           Resize: target width (preserves aspect ratio)
      --height <N>          Resize: target height (preserves aspect ratio)
      --max-dim <N>         Resize: max dimension (scale to fit)
      --template <pattern>  Output naming template (placeholders: {name}, {ext}, {format})
      --completion <shell>  Generate shell completion (bash, zsh, fish)
  -v, --verbose             Verbose output
  -V, --version             Show version
  -h, --help                Show this help

Supported formats:
  Decode: JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC/HEIF, AVIF, CR2, NEF, DNG
  Encode: JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC/HEIF, AVIF

Examples:
  pixshift photo.heic                            Convert HEIC to JPEG (default)
  pixshift -f webp -q 90 photo.heic             Convert to WebP at quality 90
  pixshift -f png -o converted/ photos/          Batch convert directory to PNG
  pixshift -j 8 -f webp -o output/ photos/      8 parallel workers
  pixshift -m -f jpg photo.heic                  Preserve EXIF metadata
  pixshift -s -f jpg photo.heic                  Strip all metadata
  pixshift photo.CR2                             Extract JPEG preview from RAW
  pixshift --max-dim 1920 -f webp photos/        Resize to fit 1920px
  pixshift --width 800 -f jpg -o thumbs/ photos/ Generate 800px-wide thumbnails
  pixshift -w -f webp ~/Pictures/                Watch mode: auto-convert new files
  pixshift -c pixshift.yaml photos/              Rules mode from config file
  pixshift --dry-run -f webp photos/             Preview what would happen
  pixshift -r -o output/ -f webp photos/         Preserve directory structure
  pixshift --template "{name}-web.{format}" -f webp photo.jpg
  pixshift --completion bash >> ~/.bashrc         Install bash completions

Shell completions:
  pixshift --completion bash  > /etc/bash_completion.d/pixshift
  pixshift --completion zsh   > "${fpath[1]}/_pixshift"
  pixshift --completion fish  > ~/.config/fish/completions/pixshift.fish
`)
}

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
