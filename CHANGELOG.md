# Changelog

All notable changes to Pixshift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2026-02-12

### Changed
- **Main.go refactored** from 1582 lines into 9 focused files under `cmd/pixshift/`: `main.go`, `args.go`, `batch.go`, `helpers.go`, `rules_mode.go`, `serve_mode.go`, `stdin.go`, `watch_mode.go`, `analysis.go`

### Added
- **JPEG XL (JXL) format** — full decode and encode via CGO bindings to libjxl, with AdvancedEncoder support (quality mapping, lossless mode)
- **GIF animation support** — new `AnimatedImage` struct and `MultiFrameDecoder`/`MultiFrameEncoder` interfaces; animated GIFs now preserve all frames through the pipeline with per-frame transforms (resize, crop, filters, watermark)
- **Additional RAW formats**: ARW (Sony), RAF (Fujifilm), ORF (Olympus), RW2 (Panasonic) — decode-only via embedded JPEG preview extraction
- **Server hardening**:
  - Middleware chain: structured JSON logging, bearer token auth (`--api-key`), CORS (`--cors-origins`), IP-based sliding-window rate limiting (`--rate-limit`)
  - Structured JSON error responses with error codes (`UNAUTHORIZED`, `RATE_LIMITED`, `INVALID_FORMAT`, etc.)
  - Configurable request timeouts (`--request-timeout`) and upload size limits (`--max-upload`)
  - `/convert` endpoint now accepts all transform, filter, and encoding parameters as form fields
- **Rules transforms** — YAML rules now support all 30+ transform, filter, and encoding fields (crop, watermark, grayscale, sepia, brightness, contrast, sharpen, blur, invert, interpolation, PNG compression, WebP method, lossless, progressive, strip/preserve metadata)
- **Watch mode improvements**:
  - Configurable debounce delay (`--watch-debounce <ms>`)
  - Ignore patterns (`--watch-ignore <glob>`, repeatable)
  - Retry with exponential backoff on conversion errors (`--watch-retry <N>`)
  - Automatic watching of newly created subdirectories when `--recursive` is set
  - Full transform support via JobTemplate (watch mode now applies all filters, watermarks, crop, etc.)
- **CI/build improvements**:
  - Makefile targets: `help`, `install`, `bench`, `coverage`, `fmt`, `vet`
  - Code coverage upload to Codecov
  - macOS Intel (`macos-13`) added to CI test matrix
  - Release builds for `linux/arm64` (cross-compiled) and `darwin/amd64`
  - SHA-256 checksums file in release artifacts
- ~2000 lines of new tests across metadata extraction/injection, pipeline transforms, watch mode, shell completions, GIF codec, and rules engine (~340 total)

## [0.4.0] - 2026-02-12

### Fixed
- EXIF auto-rotate now works correctly — orientation is parsed from raw EXIF bytes (both little-endian and big-endian), and `--auto-rotate` no longer requires `--preserve-metadata`
- Rotation transforms rewritten with direct pixel buffer operations for significantly improved performance

### Added
- **Image filters**: `--grayscale`, `--sepia`, `--brightness`, `--contrast`, `--sharpen`, `--blur`, `--invert` — applied in a fixed pipeline order after watermark
- **Format-specific encoding options**:
  - `--png-compression <0-3>` — PNG compression level (default/none/fast/best)
  - `--webp-method <0-6>` — WebP encoding speed-quality tradeoff
  - `--lossless` — WebP lossless mode
  - `--progressive` — JPEG progressive encoding (reserved for future encoder)
- **Enhanced watermarks**: font scaling (`--watermark-size`), text color (`--watermark-color`), background color (`--watermark-bg`) with hex color support
- **Resize interpolation**: `--interpolation nearest|bilinear|catmullrom` for controlling resize quality
- **User-defined presets** in YAML config (`presets:` section) — custom presets can override built-in ones and include filter/transform options (`grayscale`, `sharpen`, `auto_rotate`)
- Shell completions updated with all v0.3.0 and v0.4.0 flags for bash, zsh, and fish, including value completions
- ~40 new tests covering EXIF orientation, filters, watermark options, interpolation, and custom presets (~140 total)

## [0.3.0] - 2026-02-12

### Added
- Auto-rotate images based on EXIF orientation (`--auto-rotate`, supports orientations 1-8)
- Image cropping with exact dimensions (`--crop WxH`), aspect ratio (`--crop-ratio W:H`), and gravity anchor (`--crop-gravity center|north|south|east|west`)
- Text watermarking (`--watermark`, `--watermark-pos`, `--watermark-opacity`) with configurable position and opacity
- Directory tree view (`--tree`) showing supported image files with box-drawing characters, file sizes, and format labels
- Duplicate image detection (`--dedup`, `--dedup-threshold`) using perceptual hashing (dHash + Hamming distance)
- SSIM image quality comparison (`--ssim file1 file2`) with quality ratings (Excellent/Good/Acceptable/Poor/Bad)
- Contact sheet generation (`--contact-sheet`, `--contact-cols`, `--contact-size`) for thumbnail grid images
- Named presets (`--preset web|thumbnail|print|archive`) with predefined conversion settings
- Backup originals (`--backup`) creates `.bak` files before converting
- JSON output mode (`--json`) for machine-readable results
- Stdin/stdout piping (`cat img | pixshift -f webp - > out.webp`) for pipe-based workflows
- HTTP conversion server (`pixshift serve [addr]`) with `/convert`, `/formats`, and `/health` endpoints
- Unit tests: 60 new tests across transform, tree, preset, dedup, ssim, contact, and server packages (103 total)

## [0.2.0] - 2026-02-12

### Added
- Image resizing with `--width`, `--height`, and `--max-dim` flags (CatmullRom interpolation, aspect ratio preservation, no upscaling)
- File size reporting: human-readable input/output sizes and compression ratio per file, with total savings summary
- Progress bar for batch conversions in non-verbose mode (`schollz/progressbar`)
- Strip metadata mode (`-s` / `--strip-metadata`) to remove all EXIF/GPS data from output, mutually exclusive with `-m`
- Output naming templates (`--template`) with `{name}`, `{ext}`, `{format}` placeholders
- Shell completions for bash, zsh, and fish (`--completion bash|zsh|fish`)
- Config auto-discovery: automatically loads `pixshift.yaml` from current directory or `~/.config/pixshift/config.yaml`
- Directory structure preservation when using `-o` with `-r` (mirrors input folder hierarchy)
- Windows build target (`pixshift-windows-amd64.exe`) in release workflow via MSYS2 + MinGW
- Homebrew formula (`Formula/pixshift.rb`) for macOS and Linux binary installs
- Unit tests: 43 tests across codec, resize, pipeline, rules, and metadata packages

## [0.1.0] - 2026-02-12

### Added
- Universal image converter CLI with any-to-any format conversion
- Codec registry pattern with Decoder/Encoder interfaces for extensibility
- Smart format detection via magic bytes (not file extensions)
- **Format support:**
  - JPEG, PNG, GIF — stdlib (decode + encode)
  - WebP — decode via `x/image/webp`, encode via `kolesa-team/go-webp` (CGO)
  - TIFF, BMP — decode + encode via `x/image`
  - HEIC/HEIF — decode + encode via `vegidio/heif-go` (CGO)
  - AVIF — decode + encode via `vegidio/avif-go` (CGO)
  - CR2, NEF, DNG — decode only (extracts embedded JPEG preview from RAW)
- Parallel worker pool with configurable job count (`-j`)
- EXIF metadata preservation (`-m` flag) with extract/inject for JPEG, HEIC, TIFF sources
- Watch mode (`-w`) with fsnotify and 500ms debounce for auto-converting new files
- YAML rules engine (`-c`) with format matching, glob patterns, and first-match-wins evaluation
- Dry-run mode (`--dry-run`) to preview conversions without writing files
- Overwrite protection (skip existing outputs unless `--overwrite` is set)
- Recursive directory scanning (`-r`)
- Verbose output (`-v`) and version flag (`-V`)
- Build-time version injection via ldflags
- Makefile with `build`, `build-static`, `build-windows`, `test`, `lint` targets
- GitHub Actions CI (test + lint on PRs) and release workflow (cross-platform binaries on tags)
- Apache 2.0 license
- `pixshift.yaml.example` with sample rules configuration
- `CONTRIBUTING.md` with guide for adding new codecs

[Unreleased]: https://github.com/DanielTso/pixshift/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/DanielTso/pixshift/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/DanielTso/pixshift/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/DanielTso/pixshift/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/DanielTso/pixshift/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/DanielTso/pixshift/releases/tag/v0.1.0
