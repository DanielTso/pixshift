# Pixshift Development Chat Log

## Session: 2026-02-12

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
