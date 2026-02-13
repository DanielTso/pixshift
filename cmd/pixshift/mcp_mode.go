package main

import (
	"fmt"
	"os"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/mcp"
)

func runMCPMode(registry *codec.Registry) {
	srv := mcp.NewServer(registry)
	if err := srv.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp: %v\n", err)
		os.Exit(1)
	}
}
