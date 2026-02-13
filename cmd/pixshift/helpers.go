package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/preset"
	"github.com/DanielTso/pixshift/internal/rules"
)

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

func fatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
