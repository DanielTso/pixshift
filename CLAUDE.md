# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build binary (requires CGO_ENABLED=1)
make test           # Run all tests: go test ./...
make lint           # Run golangci-lint
go test ./internal/codec/...   # Run tests for a single package
go test -run TestDetect ./internal/codec/...  # Run a single test
```

CGO is required — the WebP, HEIC, and AVIF codecs use C libraries. On Linux, install `libwebp-dev` before building.

## Architecture

Pixshift is a CLI image converter written in Go. The entry point is `cmd/pixshift/main.go` which handles all argument parsing manually (no CLI framework).

### Core conversion flow

```
File → Detect (magic bytes) → Decode → Transform → Encode → Inject Metadata → Output
```

This is implemented in `internal/pipeline/pipeline.go:Execute()`. The `Job` struct (`pipeline/job.go`) carries all parameters for a single conversion.

### Codec registry (`internal/codec/`)

All image formats implement `Decoder` and `Encoder` interfaces defined in `codec.go`. Codecs register themselves in `DefaultRegistry()` in `registry.go`. Format detection uses magic bytes in `detect.go`, not file extensions.

**To add a new format:** create `internal/codec/yourformat.go` implementing both interfaces, add the format constant to `codec.go`, register it in `DefaultRegistry()`, add magic bytes to `detect.go`, and update `ParseFormat()`/`DefaultExtension()`/`IsSupportedExtension()`.

### Worker pool (`internal/pipeline/pool.go`)

Parallel processing via a channel-based worker pool. Configurable via `-j` flag, defaults to CPU count.

### Key packages

| Package | Purpose |
|---------|---------|
| `pipeline` | Core conversion: Job/Result structs, Execute flow, worker pool |
| `codec` | Format interfaces, registry, magic-byte detection, per-format codecs |
| `transform` | Crop (pixel/aspect/gravity), EXIF auto-rotate, text watermark |
| `resize` | CatmullRom interpolation resizing |
| `metadata` | EXIF extraction and injection |
| `rules` | YAML config engine (first-match-wins rule evaluation) |
| `preset` | Built-in presets: web, thumbnail, print, archive |
| `server` | HTTP REST API (`/convert`, `/formats`, `/health`) |
| `watch` | fsnotify-based watch mode (500ms debounce) |
| `dedup` | Duplicate detection via perceptual hashing (dHash + Hamming) |
| `ssim` | Structural similarity image comparison |
| `contact` | Contact sheet (JPEG thumbnail grid) generation |

### CLI modes in main.go

The main file (~1300 lines) orchestrates: batch conversion, watch mode, rules engine, stdin/stdout piping, HTTP server, and analysis tools (tree, dedup, SSIM, contact sheet). Arguments are parsed in `parseArgs()`.

## Build-time version injection

Version, commit, and date are injected via ldflags — see `Makefile` and `internal/version/`. Don't hardcode version strings.
