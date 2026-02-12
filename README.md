# Pixshift

Universal image converter CLI. Convert between JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC/HEIF, AVIF, and RAW camera formats (CR2, NEF, DNG).

## Features

- **Any-to-any conversion** between 10+ image formats
- **Smart format detection** via magic bytes (not file extensions)
- **Image transforms** — auto-rotate from EXIF, crop (dimensions or aspect ratio), text watermarks
- **Image resizing** — scale by width, height, or max dimension with aspect ratio preservation
- **Named presets** — `web`, `thumbnail`, `print`, `archive` with predefined settings
- **Parallel processing** with configurable worker pool
- **File size reporting** — see input/output sizes, compression ratios, and total savings
- **Progress bar** — visual progress for batch conversions
- **Metadata preservation** — keep EXIF data across conversions
- **Strip metadata** — remove all EXIF/GPS data for privacy
- **Watch mode** — auto-convert new files dropped into a directory
- **Rules engine** — YAML config for batch conversion with per-format rules
- **Config auto-discovery** — automatically loads `pixshift.yaml` from current directory or `~/.config/pixshift/`
- **Output templates** — custom output filenames with `{name}`, `{ext}`, `{format}` placeholders
- **Directory structure preservation** — mirror input folder hierarchy with `-o` and `-r`
- **Duplicate detection** — find duplicate images using perceptual hashing (dHash)
- **SSIM comparison** — compare image quality with Structural Similarity Index
- **Contact sheets** — generate thumbnail grid images from a directory
- **Directory tree view** — display image files in a tree with sizes and formats
- **HTTP server** — REST API for image conversion (`pixshift serve`)
- **JSON output** — machine-readable results for scripting
- **Stdin/stdout** — pipe-based workflows (`cat img | pixshift -f webp - > out.webp`)
- **Backup originals** — create `.bak` files before converting
- **Shell completions** — bash, zsh, and fish
- **RAW support** — extract JPEG previews from CR2, NEF, DNG files
- **Cross-platform** — Linux, macOS, and Windows binaries

## Format Support

| Format | Decode | Encode | Notes |
|--------|--------|--------|-------|
| JPEG | Yes | Yes | stdlib |
| PNG | Yes | Yes | stdlib |
| GIF | Yes | Yes | stdlib |
| WebP | Yes | Yes | CGO for encode |
| TIFF | Yes | Yes | |
| BMP | Yes | Yes | |
| HEIC/HEIF | Yes | Yes | CGO |
| AVIF | Yes | Yes | CGO |
| CR2 | Yes | - | Extracts embedded preview |
| NEF | Yes | - | Extracts embedded preview |
| DNG | Yes | - | Extracts embedded preview |

## Install

### Binary download

Download a binary from [Releases](https://github.com/DanielTso/pixshift/releases) for Linux, macOS, or Windows.

### Homebrew (macOS / Linux)

```bash
brew install DanielTso/tap/pixshift
```

### Build from source

```bash
git clone https://github.com/DanielTso/pixshift.git
cd pixshift
make build
```

## Usage

```bash
# Basic conversion (auto-detects input format)
pixshift photo.heic                            # -> photo.jpg

# Specify output format and quality
pixshift -f webp -q 90 photo.heic             # -> photo.webp at quality 90

# Batch convert a directory
pixshift -f png -o converted/ photos/

# Use a named preset
pixshift --preset web -o output/ photos/       # WebP, q85, max 1920px, strip metadata
pixshift --preset thumbnail -o thumbs/ photos/ # JPEG, q80, max 300px

# Parallel workers
pixshift -j 8 -f webp -o output/ photos/

# Resize images
pixshift --max-dim 1920 -f webp photos/        # Scale to fit 1920px
pixshift --width 800 -f jpg -o thumbs/ photos/ # 800px-wide thumbnails

# Auto-rotate from EXIF orientation
pixshift --auto-rotate -f jpg photo.heic

# Crop images
pixshift --crop 800x600 -f jpg photo.heic      # Exact dimensions
pixshift --crop-ratio 16:9 -f webp photo.jpg   # Aspect ratio
pixshift --crop-ratio 1:1 --crop-gravity north -f jpg photo.heic  # Square, top-anchored

# Add watermark
pixshift --watermark "© 2026" -f jpg photos/
pixshift --watermark "DRAFT" --watermark-pos center --watermark-opacity 0.3 -f png photo.jpg

# Preserve or strip EXIF metadata
pixshift -m -f jpg photo.heic                  # Preserve EXIF
pixshift -s -f jpg photo.heic                  # Strip all metadata

# Backup originals before converting
pixshift --backup -f webp photos/

# Extract JPEG preview from RAW
pixshift photo.CR2

# Watch mode: auto-convert new files
pixshift -w -f webp ~/Pictures/

# Rules mode from config file
pixshift -c pixshift.yaml photos/

# Preview what would happen
pixshift --dry-run -f webp photos/

# Preserve directory structure
pixshift -r -o output/ -f webp photos/

# Custom output naming
pixshift --template "{name}-web.{format}" -f webp photo.jpg

# JSON output for scripting
pixshift --json -f webp photos/

# Stdin/stdout pipeline
cat photo.heic | pixshift -f webp - > photo.webp
```

### Output example

```
[1/3] photo1.heic (4.2 MB) -> photo1.webp (890 KB) [79% smaller]
[2/3] photo2.heic (3.8 MB) -> photo2.webp (720 KB) [81% smaller]
[3/3] photo3.heic (5.1 MB) -> photo3.webp (1.1 MB) [78% smaller]

Done. 3 converted, 0 failed. Total: 13.1 MB -> 2.7 MB (79% smaller)
```

## Analysis Tools

```bash
# Show directory tree of supported images
pixshift --tree ~/Pictures
pixshift --tree -v ~/Pictures    # Include format labels

# Find duplicate images
pixshift --dedup ~/Pictures
pixshift --dedup --dedup-threshold 5 photos/   # Stricter matching

# Compare image quality (SSIM)
pixshift --ssim original.jpg compressed.jpg
# Output: SSIM: 0.9542 (Good)

# Generate contact sheet (thumbnail grid)
pixshift --contact-sheet photos/
pixshift --contact-sheet --contact-cols 6 --contact-size 150 -o output/ photos/
```

## HTTP Server

```bash
# Start conversion server (default :8080)
pixshift serve
pixshift serve :9090

# Convert an image via API
curl -F "file=@photo.heic" -F "format=webp" -F "quality=90" \
  http://localhost:8080/convert -o photo.webp

# List supported formats
curl http://localhost:8080/formats

# Health check
curl http://localhost:8080/health
```

## Flags

### Conversion

| Flag | Description |
|------|-------------|
| `-f, --format` | Output format (jpg, png, gif, webp, tiff, bmp, heic, avif) |
| `-q, --quality` | Encoding quality 1-100 (default: 92) |
| `-j, --jobs` | Number of parallel workers (default: CPU count) |
| `-o, --output` | Output directory |
| `-r, --recursive` | Process directories recursively |
| `-m, --preserve-metadata` | Preserve EXIF metadata |
| `-s, --strip-metadata` | Strip all EXIF/GPS metadata |
| `-w, --watch` | Watch mode |
| `-c, --config` | Rules config file |
| `--preset` | Named preset: `web`, `thumbnail`, `print`, `archive` |
| `--template` | Output naming template (`{name}`, `{ext}`, `{format}`) |
| `--overwrite` | Overwrite existing files |
| `--dry-run` | Preview without converting |
| `--backup` | Create `.bak` backup of originals |
| `--json` | Output results as JSON |

### Image Transforms

| Flag | Description |
|------|-------------|
| `--auto-rotate` | Auto-rotate based on EXIF orientation |
| `--crop WxH` | Crop to exact pixel dimensions (e.g. `800x600`) |
| `--crop-ratio W:H` | Crop to aspect ratio (e.g. `16:9`) |
| `--crop-gravity` | Crop anchor: `center`, `north`, `south`, `east`, `west` |
| `--watermark` | Add text watermark |
| `--watermark-pos` | Position: `bottom-right`, `bottom-left`, `top-right`, `top-left`, `center` |
| `--watermark-opacity` | Opacity 0.0-1.0 (default: 0.5) |
| `--width` | Resize: target width (preserves aspect ratio) |
| `--height` | Resize: target height (preserves aspect ratio) |
| `--max-dim` | Resize: max dimension (scale to fit) |

### Analysis

| Flag | Description |
|------|-------------|
| `--tree` | Show directory tree of supported images |
| `--dedup` | Find duplicate images using perceptual hashing |
| `--dedup-threshold` | Hamming distance threshold (default: 10) |
| `--ssim file1 file2` | Compare two images using SSIM |
| `--contact-sheet` | Generate a contact sheet (thumbnail grid) |
| `--contact-cols` | Contact sheet columns (default: 4) |
| `--contact-size` | Thumbnail size in pixels (default: 200) |

### Other

| Flag | Description |
|------|-------------|
| `--completion` | Generate shell completion (bash, zsh, fish) |
| `-v, --verbose` | Verbose output |
| `-V, --version` | Show version |
| `-h, --help` | Show help |

## Presets

| Preset | Format | Quality | Max Dim | Metadata |
|--------|--------|---------|---------|----------|
| `web` | WebP | 85 | 1920px | Strip |
| `thumbnail` | JPEG | 80 | 300px | Strip |
| `print` | TIFF | 100 | - | Preserve |
| `archive` | PNG | 100 | - | Preserve |

## Rules Config

Create a `pixshift.yaml` to define per-format conversion rules. Pixshift auto-discovers config files in the current directory or `~/.config/pixshift/`.

```yaml
rules:
  - name: heic-to-webp
    format: heic
    output: webp
    quality: 90

  - name: raw-to-jpeg
    format: cr2
    output: jpg
    quality: 95

  - name: default
    output: jpg
    quality: 92
```

Rules are evaluated in order. First match wins. See [pixshift.yaml.example](pixshift.yaml.example) for more examples.

## Shell Completions

```bash
# Bash
pixshift --completion bash > /etc/bash_completion.d/pixshift

# Zsh
pixshift --completion zsh > "${fpath[1]}/_pixshift"

# Fish
pixshift --completion fish > ~/.config/fish/completions/pixshift.fish
```

## Building

Requires Go 1.24+ and CGO (for HEIC, AVIF, and WebP encoding).

```bash
make build          # Build for current platform
make build-static   # Build with static linking (Linux)
make test           # Run tests
make lint           # Run linter
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
