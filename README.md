# Pixshift

Universal image converter CLI. Convert between JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC/HEIF, AVIF, and RAW camera formats (CR2, NEF, DNG).

## Features

- **Any-to-any conversion** between 10+ image formats
- **Smart format detection** via magic bytes (not file extensions)
- **Image resizing** — scale by width, height, or max dimension with aspect ratio preservation
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

# Parallel workers
pixshift -j 8 -f webp -o output/ photos/

# Resize images
pixshift --max-dim 1920 -f webp photos/        # Scale to fit 1920px
pixshift --width 800 -f jpg -o thumbs/ photos/  # 800px-wide thumbnails

# Preserve or strip EXIF metadata
pixshift -m -f jpg photo.heic                  # Preserve EXIF
pixshift -s -f jpg photo.heic                  # Strip all metadata

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
```

### Output example

```
[1/3] photo1.heic (4.2 MB) -> photo1.webp (890 KB) [79% smaller]
[2/3] photo2.heic (3.8 MB) -> photo2.webp (720 KB) [81% smaller]
[3/3] photo3.heic (5.1 MB) -> photo3.webp (1.1 MB) [78% smaller]

Done. 3 converted, 0 failed. Total: 13.1 MB -> 2.7 MB (79% smaller)
```

## Flags

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
| `--width` | Resize: target width (preserves aspect ratio) |
| `--height` | Resize: target height (preserves aspect ratio) |
| `--max-dim` | Resize: max dimension (scale to fit) |
| `--template` | Output naming template (`{name}`, `{ext}`, `{format}`) |
| `--overwrite` | Overwrite existing files |
| `--dry-run` | Preview without converting |
| `--completion` | Generate shell completion (bash, zsh, fish) |
| `-v, --verbose` | Verbose output |
| `-V, --version` | Show version |

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
