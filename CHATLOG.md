# Pixshift Development Chat Log

## Session 2: 2026-02-12 (v0.2.0)

### Summary

Added 11 new features to Pixshift using an agent team (4 parallel agents) for implementation. All features tested, 43 unit tests added, v0.2.0 released with Linux, macOS, and Windows binaries.

### What Was Built

**Wave 1 — Parallel (3 agents):**

1. **Image Resize Package** (`internal/resize/`) — `Resize()` function using `golang.org/x/image/draw` with CatmullRom interpolation. Supports width-only, height-only, max-dimension, and exact sizing. Preserves aspect ratio, no upscaling.

2. **Shell Completions** (`internal/completion/`) — `GenerateBash()`, `GenerateZsh()`, `GenerateFish()` generating full completion scripts for all flags, format values, and file arguments.

3. **Windows CI + Homebrew** — Added `windows-amd64` target to release workflow using MSYS2/MinGW for CGO. Created `Formula/pixshift.rb` Homebrew formula for binary installs.

**Wave 2 — Integration (sequential, touches shared files):**

4. **Pipeline Changes** — `Execute()` now returns `(inputSize, outputSize, error)`. Added resize step between decode and encode. Added `StripMetadata`, `Width`, `Height`, `MaxDim` fields to `Job`. Added `InputSize`, `OutputSize` to `Result`.

5. **File Size Reporting** — Human-readable sizes per file and total savings summary: `photo.heic (4.2 MB) -> photo.webp (890 KB) [79% smaller]`.

6. **Progress Bar** — Added `schollz/progressbar/v3` dependency. Non-verbose batch mode shows a progress bar instead of per-file output.

7. **Strip Metadata** (`-s`/`--strip-metadata`) — Mutually exclusive with `-m`. Skips EXIF extraction/injection entirely.

8. **Output Naming Templates** (`--template`) — Placeholders `{name}`, `{ext}`, `{format}` for custom output filenames.

9. **Config Auto-Discovery** — Checks `./pixshift.yaml`, `./pixshift.yml`, `~/.config/pixshift/config.yaml`, `~/.config/pixshift/config.yml` when no `-c` flag is set.

10. **Directory Structure Preservation** — When using `-o` with `-r`, mirrors input folder hierarchy in the output directory.

11. **Shell Completion Flag** (`--completion bash|zsh|fish`) — Prints completion script and exits.

**Wave 3 — Testing (1 agent):**

12. **43 Unit Tests** across 8 test files:
    - `codec/codec_test.go` — Format constants, extensions, IsRAW
    - `codec/detect_test.go` — Magic byte detection for JPEG, PNG, GIF, BMP, WebP, TIFF, extension fallback
    - `codec/roundtrip_test.go` — Encode/decode round-trips for JPEG, PNG, GIF, BMP, TIFF; quality size test
    - `resize/resize_test.go` — Width, height, max-dim, both, no-op, no-upscale, portrait
    - `pipeline/pipeline_test.go` — JPEG→PNG conversion, resize integration, error handling
    - `rules/config_test.go` — Valid parsing, invalid formats, missing output
    - `rules/engine_test.go` — Format match, glob match, no-match, first-match-wins, quality fallback
    - `metadata/metadata_test.go` — HasEXIF edge cases (nil, empty, valid)

### CI/Release Issues Resolved

- **Windows CGO build**: Git Bash on Windows can't see MSYS2 MinGW GCC. Fixed by splitting into `Build (Unix)` and `Build (Windows)` steps with `shell: msys2 {0}` and `path-type: inherit`.
- **Release continues on failure**: Added `if: ${{ !cancelled() }}` to release job so Linux/macOS binaries ship even if Windows fails.

### Release

- Tagged and released **v0.2.0** with 3 platform binaries: `pixshift-linux-amd64`, `pixshift-darwin-arm64`, `pixshift-windows-amd64.exe`.
- Updated CHANGELOG.md, README.md with all new features.
- Binary installed to `~/.local/bin/pixshift`.

### File Count

30 Go source files (8 test files), 1,541 new lines added.

---

## Session 1: 2026-02-12 (v0.1.0)

### Summary

Built **Pixshift**, a universal image converter CLI in Go, evolving from the existing `photo-converter` HEIC-to-JPEG tool into a full any-to-any converter.

### What Was Built

Implemented all 6 planned phases in a single session:

1. **Core Foundation** — Go module, codec interfaces (`Decoder`/`Encoder`), format registry pattern, magic-byte detection, and worker pool with parallel job processing.

2. **All Format Codecs** — JPEG, PNG, GIF (stdlib), WebP (x/image + go-webp CGO), TIFF, BMP (x/image), HEIC/HEIF (heif-go CGO), AVIF (avif-go CGO), and RAW pseudo-codec (CR2/NEF/DNG embedded JPEG preview extraction).

3. **Metadata Preservation** — EXIF extraction from JPEG (APP1 segment parsing), HEIC (ISO BMFF Exif box), and TIFF/RAW sources. EXIF injection into output JPEG files.

4. **Watch Mode** — fsnotify-based filesystem watcher with 500ms debounce to handle partial file writes. Graceful shutdown via context cancellation.

5. **Rules Engine** — YAML config parser with ordered rules supporting format match, glob patterns, output format, quality, and directory overrides. First-match-wins evaluation.

6. **CLI + Build System** — Full flag parsing (`-f`, `-q`, `-j`, `-o`, `-r`, `-m`, `-w`, `-c`, `--overwrite`, `--dry-run`, `-v`, `-V`), Makefile with ldflags version injection, GitHub Actions CI and cross-platform release workflows.

### Build Issues Resolved

- **heif-go / avif-go API**: `Encode()` takes `*Options` struct, not plain `int` for quality. AVIF uses `ColorQuality`/`AlphaQuality` fields instead of a single `Quality`.
- **libwebp-dev headers**: System had runtime `libwebp7` but not development headers. Solved by extracting `libwebp-dev` .deb to `/tmp` and passing custom `CGO_CFLAGS`/`CGO_LDFLAGS`.
- **Static vs dynamic linking**: Removed static `libwebp.a` (missing `libsharpyuv` dependency) and linked against system `libwebp.so.7` instead.
- **`.gitignore` conflict**: Pattern `pixshift` matched `cmd/pixshift/` directory. Fixed with `/pixshift` (root-only match).

### Testing

- Verified all 8 encode formats and 10 decode formats with synthetic test images.
- Tested HEIC round-trip (HEIC → JPEG → back).
- Tested dry-run, overwrite protection, directory batch mode, and rules engine.
- Successfully converted 25 real HEIC photos from `~/GoogleDrive/SLN6A/` to JPEG (0 failures).
- Successfully converted 50 images (HEIC + JPEG) to WebP at quality 90 (0 failures).
- Binary installed to `~/.local/bin/pixshift` for system-wide use.

### Repo Setup

- Created GitHub repo `DanielTso/pixshift` and pushed initial commit.
- Reverted pixshift code from `photo-converter` repo to keep projects separate.
- Fixed git email from work address to personal (`dnltso@hotmail.com`).

### File Count

22 Go source files, 33 total files committed (2,720 lines).
