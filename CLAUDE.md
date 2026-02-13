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

CGO is required — the WebP, HEIC, AVIF, and JXL codecs use C libraries. On Linux, install `libwebp-dev` and `libjxl-dev` before building.

## Architecture

Pixshift is a CLI image converter written in Go. The entry point is `cmd/pixshift/main.go` which dispatches to mode-specific files. Arguments are parsed manually in `args.go` (no CLI framework).

### Core conversion flow

```
File → Detect (magic bytes) → Decode → Transform → Encode → Inject Metadata → Output
```

This is implemented in `internal/pipeline/pipeline.go:Execute()`. The `Job` struct (`pipeline/job.go`) carries all parameters for a single conversion. Animated GIFs take a multi-frame path: `DecodeAll` → per-frame `transformImage` → `EncodeAll`.

### Codec registry (`internal/codec/`)

All image formats implement `Decoder` and `Encoder` interfaces defined in `codec.go`. Optional interfaces: `AdvancedEncoder` (format-specific encoding options), `MultiFrameDecoder`/`MultiFrameEncoder` (animated images). Codecs register themselves in `DefaultRegistry()` in `registry.go`. Format detection uses magic bytes in `detect.go`, not file extensions.

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
| `server` | HTTP REST API with auth, rate limiting, CORS middleware |
| `watch` | fsnotify-based watch mode with configurable debounce, ignore patterns, retry |
| `dedup` | Duplicate detection via perceptual hashing (dHash + Hamming) |
| `ssim` | Structural similarity image comparison |
| `contact` | Contact sheet (JPEG thumbnail grid) generation |

### CLI structure (`cmd/pixshift/`)

| File | Contents |
|------|----------|
| `main.go` | Entry point, signal handling, preset application, mode dispatch |
| `args.go` | `options` struct, `parseArgs()`, `printUsage()` |
| `batch.go` | `runBatchMode()` — parallel batch conversion with progress bar |
| `rules_mode.go` | `runRulesMode()`, `runRulesWatch()` |
| `watch_mode.go` | `runWatchMode()` |
| `serve_mode.go` | `runServeMode()` |
| `analysis.go` | `runTreeMode()`, `runSSIMMode()`, `runDedupMode()`, `runContactSheetMode()` |
| `stdin.go` | `runStdinMode()` |
| `helpers.go` | `buildJob()`, `collectFiles()`, `buildOutputPath()`, formatting helpers |

## Build-time version injection

Version, commit, and date are injected via ldflags — see `Makefile` and `internal/version/`. Don't hardcode version strings.
