package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/schollz/progressbar/v3"
)

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
