package main

import (
	"context"
	"fmt"
	"time"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/server"
)

func runServeMode(ctx context.Context, registry *codec.Registry, opts *options) {
	fmt.Printf("Starting Pixshift HTTP server on %s...\n", opts.serveAddr)
	srv := server.New(registry, opts.serveAddr)
	if opts.apiKey != "" {
		srv.APIKey = opts.apiKey
	}
	if opts.rateLimit > 0 {
		srv.RateLimit = opts.rateLimit
	}
	if opts.corsOrigins != "" {
		srv.AllowOrigins = opts.corsOrigins
	}
	if opts.requestTimeout > 0 {
		srv.Timeout = time.Duration(opts.requestTimeout) * time.Second
	}
	if opts.maxUpload > 0 {
		srv.MaxFileSize = opts.maxUpload
	}
	if err := srv.Start(ctx); err != nil {
		fatal("server: %v", err)
	}
}
