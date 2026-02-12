# Changelog

All notable changes to Pixshift will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
