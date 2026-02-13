# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
make build          # Build Go binary (requires CGO_ENABLED=1)
make docker         # Build Docker image
make test           # Run all tests: go test ./...
make lint           # Run golangci-lint
make bench          # Run benchmarks
make coverage       # Generate HTML coverage report
make fmt            # Format code (gofmt)
make vet            # Run go vet
make package-deb    # Build .deb package (requires nfpm)
make package-rpm    # Build .rpm package (requires nfpm)
go test ./internal/codec/...   # Run tests for a single package
go test -run TestDetect ./internal/codec/...  # Run a single test
```

Requires Go 1.24+ and CGO. The WebP, HEIC, AVIF, and JXL codecs use C libraries. On Linux, install `libwebp-dev`, `libjxl-dev`, and `libheif-dev` before building. All CGO codecs compile unconditionally (no build tags) — you cannot selectively disable them.

## Architecture

Pixshift is a Go CLI tool that also serves as an HTTP server and MCP server from a single binary. The entry point is `cmd/pixshift/main.go` which dispatches to mode-specific files. Arguments are parsed manually in `args.go` (no CLI framework).

### Core conversion flow

```
File → Detect (magic bytes) → Decode → Transform → Encode → Inject Metadata → Output
```

This is implemented in `internal/pipeline/pipeline.go:Execute()`. The `Job` struct (`pipeline/job.go`) carries all parameters for a single conversion. Animated GIFs take a multi-frame path: `DecodeAll` → per-frame `transformImage` → `EncodeAll`.

### HTTP server (`internal/server/`)

Simple HTTP server with bearer token auth, basic rate limiting, and conversion endpoints.

**Files:**
- `server.go` — Server struct, routing, middleware (CORS, logging, auth, security headers), handlers for `/health`, `/formats`, `/convert`, `/palette`, `/analyze`
- `ratelimit.go` — Sliding window rate limiter with `AllowN(key, limit)`

**Endpoints:**

| Path | Method | Description |
|------|--------|-------------|
| `/convert` | POST | Convert an image (multipart form with all transforms) |
| `/palette` | POST | Extract dominant color palette |
| `/analyze` | POST | Get image dimensions, format, size |
| `/formats` | GET | List supported decode/encode formats |
| `/health` | GET | Health check |

### Codec registry (`internal/codec/`)

All image formats implement `Decoder` and `Encoder` interfaces defined in `codec.go`. Optional interfaces: `AdvancedEncoder` (format-specific encoding options), `MultiFrameDecoder`/`MultiFrameEncoder` (animated images). Codecs register themselves in `DefaultRegistry()` in `registry.go`. Format detection uses magic bytes in `detect.go`, not file extensions.

**To add a new format:** create `internal/codec/yourformat.go` implementing both interfaces, add the format constant to `codec.go`, register it in `DefaultRegistry()`, add magic bytes to `detect.go` (both `detectByMagic` and `detectByExtension`), update `ParseFormat()`/`DefaultExtension()`/`IsSupportedExtension()`, and update shell completions in `internal/completion/completion.go` (all 3 shells).

### Worker pool (`internal/pipeline/pool.go`)

Parallel processing via a channel-based worker pool. Configurable via `-j` flag, defaults to CPU count.

### Key packages

| Package | Purpose |
|---------|---------|
| `pipeline` | Core conversion: Job/Result structs, Execute flow, worker pool |
| `codec` | Format interfaces, registry, magic-byte detection, per-format codecs |
| `transform` | Crop (pixel/aspect/gravity), EXIF auto-rotate, text watermark, entropy-based smart crop |
| `resize` | CatmullRom interpolation resizing |
| `metadata` | EXIF extraction and injection |
| `rules` | YAML config engine (first-match-wins rule evaluation) |
| `preset` | Built-in presets: web, thumbnail, print, archive |
| `server` | HTTP server with rate limiting, CORS, auth, conversion/palette/analyze endpoints |
| `mcp` | MCP server (stdio) with convert, formats, analyze, compare tools |
| `watch` | fsnotify-based watch mode with configurable debounce, ignore patterns, retry |
| `dedup` | Duplicate detection via perceptual hashing (dHash + Hamming) |
| `ssim` | Structural similarity image comparison |
| `color` | Dominant color extraction via K-means clustering |
| `contact` | Contact sheet (JPEG thumbnail grid) generation |
| `sdk` | Public Go SDK: Convert, ConvertBytes, Analyze, Palette, Compare with functional options |

### CLI structure (`cmd/pixshift/`)

| File | Contents |
|------|----------|
| `main.go` | Entry point, signal handling, preset application, mode dispatch |
| `args.go` | `options` struct, `parseArgs()`, `printUsage()` |
| `batch.go` | `runBatchMode()` — parallel batch conversion with progress bar |
| `rules_mode.go` | `runRulesMode()`, `runRulesWatch()` |
| `watch_mode.go` | `runWatchMode()` |
| `serve_mode.go` | `runServeMode()` — starts simple HTTP server |
| `mcp_mode.go` | `runMCPMode()` — starts MCP server on stdio |
| `analysis.go` | `runTreeMode()`, `runSSIMMode()`, `runDedupMode()`, `runContactSheetMode()`, `runPaletteMode()` |
| `scan_mode.go` | `runScanMode()` — directory scanning with format stats |
| `stdin.go` | `runStdinMode()` |
| `helpers.go` | `buildJob()`, `collectFiles()`, `buildOutputPath()`, formatting helpers |

## Common extension patterns

### Adding a new CLI flag

1. Add field to `options` struct in `cmd/pixshift/args.go`
2. Add `case` in `parseArgs()` switch
3. Add to `printUsage()` help text
4. Wire through in the relevant mode file (batch.go, watch_mode.go, serve_mode.go, etc.)
5. If it maps to a Job field, add to `buildJob()` and `applyOptsToJob()` in `helpers.go`
6. Update shell completions in `internal/completion/completion.go`

### Adding a new pipeline transform

1. Add function to `internal/transform/`
2. Add field to `pipeline.Job` struct in `job.go`
3. Add to `transformImage()` helper in `pipeline.go` (respects fixed order)
4. Add to `buildJob()` and `applyOptsToJob()` in `helpers.go`
5. Add to `rules.Rule` struct and `Engine.Match()` for YAML rules support
6. Add form field parsing in `server.go:handleConvert()` for HTTP API

## Testing

Tests generate all images in-memory (no testdata directory). Each test package has its own helper functions like `solidImage()`, `gradientImage()`, etc. Run uncached tests with `go test -count=1 ./...`.

## Known gotchas

### libjpeg ABI conflict

heif-go and avif-go transitively link against libjpeg-turbo at a different ABI version than the system headers declare (headers say v80, runtime has v62). This means you **cannot** use go-libjpeg or any direct libjpeg-turbo CGO binding. The JPEG codec uses stdlib `image/jpeg` only, which produces baseline (not progressive) JPEG. The `--progressive` flag is accepted but silently falls back to baseline. See `internal/codec/jpeg.go` for details.

### CGO transitive dependencies

Adding new CGO codecs may conflict with the C libraries that heif-go and avif-go pull in at link time. Always test `make build` early when introducing new CGO dependencies.

## Linux packaging

`.deb` and `.rpm` packages are built with [nfpm](https://nfpm.goreleaser.com/). Config is in `nfpm.yaml`. The CI release workflow builds packages for amd64 and arm64 as part of the `package` job. Packages declare runtime dependencies on the CGO libraries (libwebp, libjxl, libjpeg-turbo, libheif). For local testing: `make package-deb` / `make package-rpm` (requires nfpm installed).

## Build-time version injection

Version, commit, and date are injected via ldflags — see `Makefile` and `internal/version/`. Don't hardcode version strings.

## Development history

See `CHATLOG.md` for a detailed log of every development session, what was built, issues resolved, and agent team structure used for each version.
