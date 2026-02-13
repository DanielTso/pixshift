package main

import (
	"io"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
)

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
