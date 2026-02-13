package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/DanielTso/pixshift/internal/watch"
)

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
		Pipeline:       pipe,
		OutputDir:      opts.outputDir,
		OutputFormat:   outputFormat,
		Quality:        opts.quality,
		Recursive:      opts.recursive,
		Metadata:       opts.metadata,
		Verbose:        opts.verbose,
		DebounceDelay:  time.Duration(opts.watchDebounce) * time.Millisecond,
		IgnorePatterns: opts.watchIgnore,
		MaxRetries:     opts.watchRetry,
		JobTemplate:    buildJob(opts, "", "", outputFormat, ""),
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
