# Changelog

All notable changes to Pixshift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/DanielTso/pixshift/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/DanielTso/pixshift/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/DanielTso/pixshift/releases/tag/v0.1.0
