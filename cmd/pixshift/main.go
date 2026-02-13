package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/completion"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/preset"
)

const defaultQuality = 92

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

	// MCP mode
	if opts.mcpMode {
		runMCPMode(registry)
		return
	}

	pipe := pipeline.NewPipeline(registry)

	// Serve mode
	if opts.serveAddr != "" {
		runServeMode(ctx, registry, opts)
		return
	}

	// Scan mode
	if opts.scanMode {
		runScanMode(opts)
		return
	}

	// Palette mode
	if opts.paletteCount > 0 {
		runPaletteMode(registry, opts)
		return
	}

	// Tree mode
	if opts.treeMode {
		runTreeMode(opts)
		return
	}

	// SSIM comparison mode
	if len(opts.ssimFiles) == 2 {
		runSSIMMode(opts)
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
