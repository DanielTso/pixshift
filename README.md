# Pixshift

Universal image converter — CLI, Web App, Hosted API, and MCP Server. Convert between JPEG, PNG, GIF, WebP, TIFF, BMP, HEIC/HEIF, AVIF, JPEG XL, and RAW camera formats (CR2, NEF, DNG, ARW, RAF, ORF, RW2).

## 4 Ways to Use Pixshift

| Surface | Description |
|---------|-------------|
| **CLI** | `pixshift -f webp photo.heic` — batch convert, watch mode, rules engine |
| **Web App** | Drag-and-drop converter at [pixshift.dev](https://pixshift.dev) |
| **Hosted API** | `POST /api/v1/convert` with API key auth and usage tracking |
| **MCP Server** | `pixshift mcp` — use from Claude Desktop or any MCP client |

One binary serves everything. Without `DATABASE_URL`, it runs as a simple CLI/server. With a database, it unlocks auth, billing, and the web app.

## Features

- **Any-to-any conversion** between 15+ image formats
- **Smart format detection** via magic bytes (not file extensions)
- **GIF animation** — preserve all frames through the pipeline with per-frame transforms
- **Image transforms** — auto-rotate from EXIF, crop (dimensions or aspect ratio), smart crop (entropy-based), text watermarks with font scaling and custom colors
- **Image filters** — grayscale, sepia, brightness, contrast, sharpen, blur, invert
- **Color palette extraction** — extract dominant colors using K-means clustering
- **Image resizing** — scale by width, height, or max dimension with selectable interpolation (nearest, bilinear, catmull-rom)
- **Format-specific encoding** — PNG compression level, WebP method/lossless, JXL quality/lossless, JPEG progressive (reserved)
- **Named presets** — built-in `web`, `thumbnail`, `print`, `archive` plus user-defined presets in YAML config
- **Parallel processing** with configurable worker pool
- **File size reporting** — see input/output sizes, compression ratios, and total savings
- **Progress bar** — visual progress for batch conversions
- **Metadata preservation** — keep EXIF data across conversions
- **Strip metadata** — remove all EXIF/GPS data for privacy
- **Watch mode** — auto-convert new files with configurable debounce, ignore patterns, and retry
- **Rules engine** — YAML config with per-format rules supporting all transforms, filters, and encoding options
- **Config auto-discovery** — automatically loads `pixshift.yaml` from current directory or `~/.config/pixshift/`
- **Output templates** — custom output filenames with `{name}`, `{ext}`, `{format}` placeholders
- **Directory structure preservation** — mirror input folder hierarchy with `-o` and `-r`
- **Duplicate detection** — find duplicate images using perceptual hashing (dHash)
- **SSIM comparison** — compare image quality with Structural Similarity Index
- **Contact sheets** — generate thumbnail grid images from a directory
- **Directory tree view** — display image files in a tree with sizes and formats
- **Web application** — React SPA with batch upload, drag-and-drop converter, live preview, and before/after comparison
- **Hosted API** — REST API with API key auth, tier-based rate limiting, and usage tracking
- **MCP server** — Model Context Protocol server for AI assistant integration
- **Go SDK** — embed Pixshift in your own Go applications with a functional options API
- **Stripe billing** — Starter/Pro/Business tiers with subscription management and annual billing
- **HTTP server** — REST API with auth, rate limiting, CORS, and full transform support (`pixshift serve`)
- **JSON output** — machine-readable results for scripting
- **Stdin/stdout** — pipe-based workflows (`cat img | pixshift -f webp - > out.webp`)
- **Backup originals** — create `.bak` files before converting
- **Shell completions** — bash, zsh, and fish
- **RAW support** — extract JPEG previews from CR2, NEF, DNG, ARW, RAF, ORF, RW2 files
- **Cross-platform** — Linux (amd64/arm64), macOS (Intel/Apple Silicon), and Windows binaries

## Format Support

| Format | Decode | Encode | Notes |
|--------|--------|--------|-------|
| JPEG | Yes | Yes | stdlib |
| PNG | Yes | Yes | stdlib |
| GIF | Yes | Yes | stdlib, animated GIF support |
| WebP | Yes | Yes | CGO for encode |
| TIFF | Yes | Yes | |
| BMP | Yes | Yes | |
| HEIC/HEIF | Yes | Yes | CGO |
| AVIF | Yes | Yes | CGO |
| JPEG XL | Yes | Yes | CGO (libjxl) |
| CR2 | Yes | - | Extracts embedded preview |
| NEF | Yes | - | Extracts embedded preview |
| DNG | Yes | - | Extracts embedded preview |
| ARW | Yes | - | Sony, extracts embedded preview |
| RAF | Yes | - | Fujifilm, extracts embedded preview |
| ORF | Yes | - | Olympus, extracts embedded preview |
| RW2 | Yes | - | Panasonic, extracts embedded preview |

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

## CLI Usage

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
pixshift --max-dim 1920 --interpolation nearest -f webp photos/  # Nearest-neighbor resize

# Auto-rotate from EXIF orientation
pixshift --auto-rotate -f jpg photo.heic

# Crop images
pixshift --crop 800x600 -f jpg photo.heic      # Exact dimensions
pixshift --crop-ratio 16:9 -f webp photo.jpg   # Aspect ratio
pixshift --crop-ratio 1:1 --crop-gravity north -f jpg photo.heic  # Square, top-anchored

# Image filters
pixshift --grayscale -f jpg photo.heic                 # Convert to grayscale
pixshift --brightness 20 --contrast 30 -f webp photo.jpg  # Adjust brightness/contrast
pixshift --sharpen -f png photo.jpg                    # Sharpen image
pixshift --blur 2 -f jpg photo.heic                    # Gaussian blur (radius 2)
pixshift --sepia 0.8 -f jpg photo.heic                 # Warm sepia tone
pixshift --invert -f png photo.jpg                     # Invert colors

# Add watermark
pixshift --watermark "© 2026" -f jpg photos/
pixshift --watermark "DRAFT" --watermark-pos center --watermark-opacity 0.3 -f png photo.jpg
pixshift --watermark "PROOF" --watermark-size 3 --watermark-color "#FF0000" -f jpg photo.jpg

# Preserve or strip EXIF metadata
pixshift -m -f jpg photo.heic                  # Preserve EXIF
pixshift -s -f jpg photo.heic                  # Strip all metadata

# Backup originals before converting
pixshift --backup -f webp photos/

# Convert to JPEG XL
pixshift -f jxl -q 90 photo.jpg
pixshift -f jxl --lossless photo.png                    # Lossless JXL

# Extract JPEG preview from RAW
pixshift photo.CR2
pixshift photo.arw                                       # Sony ARW
pixshift photo.raf                                       # Fujifilm RAF

# Watch mode: auto-convert new files
pixshift -w -f webp ~/Pictures/
pixshift -w --watch-debounce 200 --watch-ignore "*.tmp" -f webp ~/Pictures/
pixshift -w --watch-retry 3 -f webp ~/Pictures/          # Retry failed conversions

# Format-specific encoding options
pixshift --png-compression 3 -f png photo.jpg           # Best PNG compression
pixshift --webp-method 6 -f webp photo.jpg              # Best WebP quality (slower)
pixshift --lossless -f webp photo.jpg                    # Lossless WebP

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

# Smart crop (entropy-based, finds most interesting region)
pixshift --smart-crop 400x300 -f webp photo.jpg

# Extract color palette
pixshift --palette photo.jpg                      # 5 dominant colors
pixshift --palette 8 photo.jpg                    # 8 colors
pixshift --palette --json photo.jpg               # JSON output

# Scan directory for images
pixshift --scan ~/Pictures

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

## Web Application

The web app is a React SPA embedded in the Go binary and served at the root path.

**Features:**
- Batch upload — drop multiple images, thumbnail strip with status indicators, sequential conversion with progress
- Download All completed files or retry failed ones
- Drag-and-drop image upload with instant preview
- Format picker with 8+ output formats
- Transform panel: quality, resize, filters, watermark
- Before/after comparison slider
- Dashboard with usage charts, conversion history, API key management
- Three tiers: Starter (free, 20/day), Pro ($19/mo, 500/day), Business ($59/mo, unlimited)

**Access:** Visit [pixshift.dev](https://pixshift.dev) or self-host with `DATABASE_URL` set.

## Hosted API

Authenticate with an API key (`pxs_` prefix) via the `X-API-Key` header.

```bash
# Convert an image
curl -X POST https://pixshift.dev/api/v1/convert \
  -H "X-API-Key: pxs_your_key_here" \
  -F "file=@photo.heic" \
  -F "format=webp" \
  -F "quality=90" \
  -o photo.webp

# With transforms
curl -X POST https://pixshift.dev/api/v1/convert \
  -H "X-API-Key: pxs_your_key_here" \
  -F "file=@photo.jpg" \
  -F "format=webp" \
  -F "grayscale=true" \
  -F "max_dim=1920" \
  -F "watermark_text=PROOF" \
  -o out.webp

# List supported formats
curl https://pixshift.dev/api/v1/formats

# Extract color palette
curl -X POST https://pixshift.dev/api/v1/palette \
  -H "X-API-Key: pxs_your_key_here" \
  -F "file=@photo.jpg" \
  -F "count=8"

# Analyze image metadata
curl -X POST https://pixshift.dev/api/v1/analyze \
  -H "X-API-Key: pxs_your_key_here" \
  -F "file=@photo.jpg"

# Compare two images (SSIM)
curl -X POST https://pixshift.dev/api/v1/compare \
  -H "X-API-Key: pxs_your_key_here" \
  -F "file1=@original.jpg" \
  -F "file2=@compressed.jpg"
```

### API Tiers

| | Starter | Pro ($19/mo) | Business ($59/mo) |
|---|---|---|---|
| Web conversions/day | 20 | 500 | Unlimited |
| API requests/month | 100 | 5,000 | 50,000 |
| Max file size | 10 MB | 100 MB | 500 MB |
| Batch uploads | 1 file | 20 files | 100 files |
| Rate limit | 10 req/min | 30 req/min | 120 req/min |
| API keys | 1 | 5 | 20 |
| MCP integration | — | Included | Included |
| Annual billing | — | $190/yr (2 mo free) | $590/yr (2 mo free) |

All formats and transforms available on every tier. Manage your subscription and API keys from the [Dashboard](https://pixshift.dev/dashboard).

## MCP Server

Pixshift integrates with Claude Desktop and other MCP-compatible AI assistants.

```bash
# Start MCP server on stdio
pixshift mcp
```

**Claude Desktop config** (`~/.config/claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "pixshift": {
      "command": "pixshift",
      "args": ["mcp"]
    }
  }
}
```

### MCP Tools

| Tool | Description |
|------|-------------|
| `convert_image` | Convert image with optional transforms (resize, filters, watermark) |
| `get_formats` | List all supported decode/encode formats |
| `analyze_image` | Get format, dimensions, file size, and EXIF metadata |
| `compare_images` | SSIM comparison between two images |

## Go SDK

Embed Pixshift in your own Go applications:

```go
import "github.com/DanielTso/pixshift/sdk"

// Convert an image
err := sdk.Convert("photo.heic", "photo.webp",
    sdk.WithFormat(sdk.WebP),
    sdk.WithQuality(90),
    sdk.WithMaxDim(1920),
)

// Smart crop to find most interesting region
err := sdk.Convert("photo.jpg", "thumb.webp",
    sdk.WithFormat(sdk.WebP),
    sdk.WithSmartCrop(400, 300),
)

// Extract color palette
colors, err := sdk.Palette("photo.jpg", 5)
for _, c := range colors {
    fmt.Printf("%s (%.1f%%)\n", c.Hex, c.Percentage)
}

// Analyze image
info, err := sdk.Analyze("photo.jpg")
fmt.Printf("%dx%d %s\n", info.Width, info.Height, info.Format)

// Compare images (SSIM)
result, err := sdk.Compare("original.jpg", "compressed.jpg")
fmt.Printf("SSIM: %.4f (%s)\n", result.Score, result.Rating)
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

# Extract color palette
pixshift --palette photo.jpg                     # 5 dominant colors
pixshift --palette 8 photo.jpg                   # 8 colors
pixshift --palette --json photo.jpg              # JSON output

# Scan directory for image stats
pixshift --scan ~/Pictures
```

## HTTP Server (Simple Mode)

For simple self-hosted use without a database:

```bash
# Start conversion server (default :8080)
pixshift serve
pixshift serve :9090

# With authentication and rate limiting
pixshift serve --api-key mysecretkey --rate-limit 30 --cors-origins "*"

# Configure timeouts and upload limits
pixshift serve --request-timeout 120 --max-upload 100

# Convert an image via API
curl -F "file=@photo.heic" -F "format=webp" -F "quality=90" \
  http://localhost:8080/convert -o photo.webp

# With authentication
curl -H "Authorization: Bearer mysecretkey" \
  -F "file=@photo.heic" -F "format=webp" \
  http://localhost:8080/convert -o photo.webp

# List supported formats
curl http://localhost:8080/formats

# Health check
curl http://localhost:8080/health
```

## Full Mode (Self-Hosted with Database)

Set `DATABASE_URL` to enable auth, billing, web app, and hosted API:

```bash
export DATABASE_URL="postgres://user:pass@localhost/pixshift?sslmode=disable"
export SESSION_SECRET="your-random-secret"
export STRIPE_SECRET_KEY="sk_..."                    # optional
export STRIPE_PRO_MONTHLY_PRICE_ID="price_..."       # optional
export STRIPE_PRO_ANNUAL_PRICE_ID="price_..."        # optional
export STRIPE_BUSINESS_MONTHLY_PRICE_ID="price_..."  # optional
export STRIPE_BUSINESS_ANNUAL_PRICE_ID="price_..."   # optional
export STRIPE_WEBHOOK_SECRET="whsec_..."             # optional
export GOOGLE_CLIENT_ID="..."              # optional
export GOOGLE_CLIENT_SECRET="..."          # optional
export BASE_URL="https://pixshift.dev"     # optional

pixshift serve :8080
```

Routes in full mode:

| Path | Auth | Description |
|------|------|-------------|
| `POST /api/v1/convert` | API key | Hosted API conversion |
| `POST /api/v1/palette` | API key | Extract color palette |
| `POST /api/v1/analyze` | API key | Image metadata analysis |
| `POST /api/v1/compare` | API key | SSIM image comparison |
| `GET /api/v1/formats` | None | List formats |
| `POST /api/webhooks/stripe` | Signature | Stripe webhooks |
| `POST /internal/auth/signup` | None | Create account |
| `POST /internal/auth/login` | None | Login |
| `POST /internal/auth/logout` | None | Logout |
| `GET /internal/auth/google` | None | Google OAuth |
| `POST /internal/convert` | Session | Web UI conversion |
| `GET /internal/user` | Session | User profile |
| `GET/POST /internal/keys` | Session | API key management |
| `GET /internal/usage` | Session | Usage stats |
| `POST /internal/billing/checkout` | Session | Stripe checkout |
| `POST /internal/billing/portal` | Session | Billing portal |
| `GET /health` | None | Health check |
| `/*` | None | React SPA |

## Flags

### Conversion

| Flag | Description |
|------|-------------|
| `-f, --format` | Output format (jpg, png, gif, webp, tiff, bmp, heic, avif, jxl) |
| `-q, --quality` | Encoding quality 1-100 (default: 92) |
| `-j, --jobs` | Number of parallel workers (default: CPU count) |
| `-o, --output` | Output directory |
| `-r, --recursive` | Process directories recursively |
| `-m, --preserve-metadata` | Preserve EXIF metadata |
| `-s, --strip-metadata` | Strip all EXIF/GPS metadata |
| `-w, --watch` | Watch mode |
| `-c, --config` | Rules config file |
| `--preset` | Named preset: `web`, `thumbnail`, `print`, `archive` (or custom) |
| `--template` | Output naming template (`{name}`, `{ext}`, `{format}`) |
| `--overwrite` | Overwrite existing files |
| `--dry-run` | Preview without converting |
| `--backup` | Create `.bak` backup of originals |
| `--json` | Output results as JSON |

### Image Transforms

| Flag | Description |
|------|-------------|
| `--auto-rotate` | Auto-rotate based on EXIF orientation |
| `--smart-crop WxH` | Smart crop to most interesting region (entropy-based) |
| `--crop WxH` | Crop to exact pixel dimensions (e.g. `800x600`) |
| `--crop-ratio W:H` | Crop to aspect ratio (e.g. `16:9`) |
| `--crop-gravity` | Crop anchor: `center`, `north`, `south`, `east`, `west` |
| `--watermark` | Add text watermark |
| `--watermark-pos` | Position: `bottom-right`, `bottom-left`, `top-right`, `top-left`, `center` |
| `--watermark-opacity` | Opacity 0.0-1.0 (default: 0.5) |
| `--watermark-size` | Font scale factor (default: 1.0) |
| `--watermark-color` | Text color as hex (default: `#FFFFFF`) |
| `--watermark-bg` | Background color as hex (default: `#000000`) |
| `--width` | Resize: target width (preserves aspect ratio) |
| `--height` | Resize: target height (preserves aspect ratio) |
| `--max-dim` | Resize: max dimension (scale to fit) |
| `--interpolation` | Resize method: `nearest`, `bilinear`, `catmullrom` (default) |

### Image Filters

| Flag | Description |
|------|-------------|
| `--grayscale` | Convert to grayscale |
| `--sepia` | Apply sepia tone (default intensity: 0.8) |
| `--brightness` | Adjust brightness (-100 to +100) |
| `--contrast` | Adjust contrast (-100 to +100) |
| `--sharpen` | Apply sharpening filter |
| `--blur` | Apply blur with radius in pixels |
| `--invert` | Invert colors |

### Encoding Options

| Flag | Description |
|------|-------------|
| `--png-compression` | PNG compression: `0` default, `1` none, `2` fast, `3` best |
| `--webp-method` | WebP method 0-6: speed vs quality tradeoff |
| `--lossless` | WebP lossless encoding |
| `--progressive` | JPEG progressive encoding (reserved for future encoder) |

### Server

| Flag | Description |
|------|-------------|
| `--api-key` | Require bearer token authentication (simple mode) |
| `--rate-limit` | Max requests per minute per IP (0 = off) |
| `--cors-origins` | CORS allowed origins (default: `*`) |
| `--request-timeout` | Request timeout in seconds (default: 60) |
| `--max-upload` | Max upload size in MB (default: 50) |

### Watch Mode

| Flag | Description |
|------|-------------|
| `--watch-debounce` | Debounce delay in milliseconds (default: 500) |
| `--watch-ignore` | Glob pattern to ignore (repeatable, e.g. `*.tmp`) |
| `--watch-retry` | Retry failed conversions N times with exponential backoff |

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
| `--palette [N]` | Extract N dominant colors (default: 5, max: 20) |
| `--scan` | Scan directory for image files with format stats |

### Other

| Flag | Description |
|------|-------------|
| `--completion` | Generate shell completion (bash, zsh, fish) |
| `-v, --verbose` | Verbose output |
| `-V, --version` | Show version |
| `-h, --help` | Show help |

## Presets

### Built-in

| Preset | Format | Quality | Max Dim | Metadata |
|--------|--------|---------|---------|----------|
| `web` | WebP | 85 | 1920px | Strip |
| `thumbnail` | JPEG | 80 | 300px | Strip |
| `print` | TIFF | 100 | - | Preserve |
| `archive` | PNG | 100 | - | Preserve |

### Custom presets

Define your own presets in `pixshift.yaml`. Custom presets can override built-ins.

```yaml
presets:
  social:
    format: jpg
    quality: 90
    max_dim: 1080
    strip_metadata: true
    grayscale: false

  bw-archive:
    format: png
    quality: 100
    grayscale: true
    sharpen: true
    auto_rotate: true
```

```bash
pixshift --preset social -o output/ photos/
pixshift --preset bw-archive -o output/ photos/
```

## Rules Config

Create a `pixshift.yaml` to define per-format conversion rules and custom presets. Pixshift auto-discovers config files in the current directory or `~/.config/pixshift/`.

```yaml
presets:
  social:
    format: jpg
    quality: 90
    max_dim: 1080
    strip_metadata: true

rules:
  - name: heic-to-webp
    format: heic
    output: webp
    quality: 90

  - name: raw-to-jpeg
    format: cr2
    output: jpg
    quality: 95

  - name: instagram-square
    glob: "insta_*.jpg"
    output: jpeg
    crop_ratio: "1:1"
    crop_gravity: center
    quality: 90
    sharpen: true

  - name: watermark-proofs
    glob: "proof_*"
    output: jpeg
    watermark_text: "PROOF"
    watermark_pos: center
    watermark_opacity: 0.3
    grayscale: true

  - name: default
    output: jpg
    quality: 92
```

Rules support all transform, filter, and encoding options: `width`, `height`, `max_dim`, `auto_rotate`, `crop_width`, `crop_height`, `crop_ratio`, `crop_gravity`, `watermark_text/pos/opacity/size/color/bg`, `grayscale`, `sepia`, `brightness`, `contrast`, `sharpen`, `blur`, `invert`, `interpolation`, `png_compression`, `webp_method`, `lossless`, `progressive`, `strip_metadata`, `preserve_metadata`.

Rules are evaluated in order. First match wins. CLI flags override rule values. See [pixshift.yaml.example](pixshift.yaml.example) for more examples.

## Shell Completions

```bash
# Bash
pixshift --completion bash > /etc/bash_completion.d/pixshift

# Zsh
pixshift --completion zsh > "${fpath[1]}/_pixshift"

# Fish
pixshift --completion fish > ~/.config/fish/completions/pixshift.fish
```

## Deployment

### Docker

```bash
# Build
docker build -t pixshift .

# Run (simple mode)
docker run -p 8080:8080 pixshift

# Run (full mode)
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://..." \
  -e SESSION_SECRET="..." \
  -e STRIPE_SECRET_KEY="sk_..." \
  pixshift
```

### DigitalOcean App Platform

The repository includes `.do/app.yaml` for one-click deployment with managed Postgres. See the [deployment guide](https://docs.digitalocean.com/products/app-platform/) for setup.

## Building

Requires Go 1.24+ and CGO (for HEIC, AVIF, WebP, and JXL).

**Linux dependencies**: `libwebp-dev`, `libjxl-dev`

```bash
make build          # Build Go binary (CLI only)
make build-web      # Build frontend (requires Node.js 22+)
make build-all      # Build frontend + Go binary with embedded SPA
make test           # Run tests
make lint           # Run linter
make docker         # Build Docker image
make help           # Show all available targets
make install        # Install to $GOPATH/bin
make bench          # Run benchmarks
make coverage       # Generate HTML coverage report
make fmt            # Format code
make vet            # Run go vet
```

## License

Apache 2.0 — see [LICENSE](LICENSE).
