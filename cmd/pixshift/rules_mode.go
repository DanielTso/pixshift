package main

import (
	"context"
	"fmt"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/rules"
	"github.com/DanielTso/pixshift/internal/watch"
)

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
