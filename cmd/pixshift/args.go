package main

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"github.com/DanielTso/pixshift/internal/preset"
	"github.com/DanielTso/pixshift/internal/version"
)

type options struct {
	format        string
	quality       int
	jobs          int
	outputDir     string
	recursive     bool
	metadata      bool
	stripMetadata bool
	watchMode     bool
	configFile    string
	overwrite     bool
	dryRun        bool
	verbose       bool
	width         int
	height        int
	maxDim        int
	template      string
	completionSh  string
	inputs        []string

	// v0.3.0 fields
	autoRotate       bool
	cropWidth        int
	cropHeight       int
	cropRatio        string
	cropGravity      string
	watermarkText    string
	watermarkPos     string
	watermarkOpacity float64
	presetName       string
	backup           bool
	jsonOutput       bool
	treeMode         bool
	dedupMode        bool
	dedupThreshold   int
	ssimFiles        []string
	contactSheet     bool
	contactCols      int
	contactSize      int
	serveAddr        string

	// v0.4.0 fields
	grayscale      bool
	sepia          float64
	brightness     float64
	contrast       float64
	sharpen        bool
	blur           float64
	invert         bool
	progressive    bool
	pngCompression int
	webpMethod     int
	lossless       bool
	watermarkSize  float64
	watermarkColor string
	watermarkBg    string
	interpolation  string

	// v0.5.0 server fields
	apiKey         string
	rateLimit      int
	corsOrigins    string
	requestTimeout int   // seconds
	maxUpload      int64 // bytes

	// v0.6.0 fields
	mcpMode bool

	// v0.5.0 watch fields
	watchDebounce int      // milliseconds
	watchIgnore   []string // repeatable
	watchRetry    int

	// v0.7.0 fields
	scanMode        bool
	paletteCount    int // --palette N (0 = disabled)
	smartCropWidth  int
	smartCropHeight int
}

func parseArgs(args []string) *options {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	opts := &options{
		quality:          defaultQuality,
		jobs:             runtime.NumCPU(),
		watermarkOpacity: 0.5,
		dedupThreshold:   10,
		contactCols:      4,
		contactSize:      200,
	}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "-f", "--format":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.format = args[i+1]
			i += 2
		case "-q", "--quality":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			q, err := strconv.Atoi(args[i+1])
			if err != nil || q < 1 || q > 100 {
				fatal("quality must be an integer between 1 and 100")
			}
			opts.quality = q
			i += 2
		case "-j", "--jobs":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			j, err := strconv.Atoi(args[i+1])
			if err != nil || j < 1 {
				fatal("jobs must be a positive integer")
			}
			opts.jobs = j
			i += 2
		case "-o", "--output":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.outputDir = args[i+1]
			i += 2
		case "-c", "--config":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.configFile = args[i+1]
			i += 2
		case "-r", "--recursive":
			opts.recursive = true
			i++
		case "-m", "--preserve-metadata":
			opts.metadata = true
			i++
		case "-s", "--strip-metadata":
			opts.stripMetadata = true
			i++
		case "-w", "--watch":
			opts.watchMode = true
			i++
		case "--overwrite":
			opts.overwrite = true
			i++
		case "--dry-run":
			opts.dryRun = true
			i++
		case "-v", "--verbose":
			opts.verbose = true
			i++
		case "--width":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			w, err := strconv.Atoi(args[i+1])
			if err != nil || w < 1 {
				fatal("width must be a positive integer")
			}
			opts.width = w
			i += 2
		case "--height":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			h, err := strconv.Atoi(args[i+1])
			if err != nil || h < 1 {
				fatal("height must be a positive integer")
			}
			opts.height = h
			i += 2
		case "--max-dim":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			d, err := strconv.Atoi(args[i+1])
			if err != nil || d < 1 {
				fatal("max-dim must be a positive integer")
			}
			opts.maxDim = d
			i += 2
		case "--template":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.template = args[i+1]
			i += 2
		case "--completion":
			if i+1 >= len(args) {
				fatal("missing value for %s (bash, zsh, or fish)", args[i])
			}
			opts.completionSh = args[i+1]
			i += 2
		case "--auto-rotate":
			opts.autoRotate = true
			i++
		case "--crop":
			if i+1 >= len(args) {
				fatal("missing value for %s (WxH, e.g. 800x600)", args[i])
			}
			parts := strings.SplitN(args[i+1], "x", 2)
			if len(parts) != 2 {
				fatal("crop must be WxH (e.g. 800x600)")
			}
			cw, err1 := strconv.Atoi(parts[0])
			ch, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || cw < 1 || ch < 1 {
				fatal("crop dimensions must be positive integers")
			}
			opts.cropWidth = cw
			opts.cropHeight = ch
			i += 2
		case "--crop-ratio":
			if i+1 >= len(args) {
				fatal("missing value for %s (e.g. 16:9)", args[i])
			}
			opts.cropRatio = args[i+1]
			i += 2
		case "--crop-gravity":
			if i+1 >= len(args) {
				fatal("missing value for %s (center, north, south, east, west)", args[i])
			}
			opts.cropGravity = args[i+1]
			i += 2
		case "--watermark":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkText = args[i+1]
			i += 2
		case "--watermark-pos":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkPos = args[i+1]
			i += 2
		case "--watermark-opacity":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			o, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || o < 0 || o > 1 {
				fatal("watermark-opacity must be a number between 0 and 1")
			}
			opts.watermarkOpacity = o
			i += 2
		case "--preset":
			if i+1 >= len(args) {
				fatal("missing value for %s (available: %v)", args[i], preset.List())
			}
			opts.presetName = args[i+1]
			i += 2
		case "--backup":
			opts.backup = true
			i++
		case "--json":
			opts.jsonOutput = true
			i++
		case "--tree":
			opts.treeMode = true
			i++
		case "--dedup":
			opts.dedupMode = true
			i++
		case "--dedup-threshold":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			t, err := strconv.Atoi(args[i+1])
			if err != nil || t < 0 {
				fatal("dedup-threshold must be a non-negative integer")
			}
			opts.dedupThreshold = t
			i += 2
		case "--ssim":
			if i+2 >= len(args) {
				fatal("--ssim requires two file arguments")
			}
			opts.ssimFiles = []string{args[i+1], args[i+2]}
			i += 3
		case "--contact-sheet":
			opts.contactSheet = true
			i++
		case "--contact-cols":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			c, err := strconv.Atoi(args[i+1])
			if err != nil || c < 1 {
				fatal("contact-cols must be a positive integer")
			}
			opts.contactCols = c
			i += 2
		case "--contact-size":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			s, err := strconv.Atoi(args[i+1])
			if err != nil || s < 1 {
				fatal("contact-size must be a positive integer")
			}
			opts.contactSize = s
			i += 2
		case "--grayscale":
			opts.grayscale = true
			i++
		case "--sepia":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				s, err := strconv.ParseFloat(args[i+1], 64)
				if err != nil || s < 0 || s > 1 {
					fatal("sepia intensity must be a number between 0 and 1")
				}
				opts.sepia = s
				i += 2
			} else {
				opts.sepia = 0.8
				i++
			}
		case "--brightness":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			b, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || b < -100 || b > 100 {
				fatal("brightness must be a number between -100 and 100")
			}
			opts.brightness = b
			i += 2
		case "--contrast":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			c, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || c < -100 || c > 100 {
				fatal("contrast must be a number between -100 and 100")
			}
			opts.contrast = c
			i += 2
		case "--sharpen":
			opts.sharpen = true
			i++
		case "--blur":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			bl, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || bl < 0 {
				fatal("blur radius must be a non-negative number")
			}
			opts.blur = bl
			i += 2
		case "--invert":
			opts.invert = true
			i++
		case "--progressive":
			opts.progressive = true
			i++
		case "--png-compression":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			pc, err := strconv.Atoi(args[i+1])
			if err != nil || pc < 0 || pc > 3 {
				fatal("png-compression must be 0-3")
			}
			opts.pngCompression = pc
			i += 2
		case "--webp-method":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			wm, err := strconv.Atoi(args[i+1])
			if err != nil || wm < 0 || wm > 6 {
				fatal("webp-method must be 0-6")
			}
			opts.webpMethod = wm
			i += 2
		case "--lossless":
			opts.lossless = true
			i++
		case "--watermark-size":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			ws, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil || ws <= 0 {
				fatal("watermark-size must be a positive number")
			}
			opts.watermarkSize = ws
			i += 2
		case "--watermark-color":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkColor = args[i+1]
			i += 2
		case "--watermark-bg":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watermarkBg = args[i+1]
			i += 2
		case "--interpolation":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			interp := strings.ToLower(args[i+1])
			if interp != "nearest" && interp != "bilinear" && interp != "catmullrom" {
				fatal("interpolation must be nearest, bilinear, or catmullrom")
			}
			opts.interpolation = interp
			i += 2
		case "--api-key":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.apiKey = args[i+1]
			i += 2
		case "--rate-limit":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			rl, err := strconv.Atoi(args[i+1])
			if err != nil || rl < 0 {
				fatal("rate-limit must be a non-negative integer")
			}
			opts.rateLimit = rl
			i += 2
		case "--cors-origins":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.corsOrigins = args[i+1]
			i += 2
		case "--request-timeout":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			rt, err := strconv.Atoi(args[i+1])
			if err != nil || rt < 1 {
				fatal("request-timeout must be a positive integer (seconds)")
			}
			opts.requestTimeout = rt
			i += 2
		case "--max-upload":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			mu, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || mu < 1 {
				fatal("max-upload must be a positive integer (MB)")
			}
			opts.maxUpload = mu * 1024 * 1024 // convert MB to bytes
			i += 2
		case "--watch-debounce":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			wd, err := strconv.Atoi(args[i+1])
			if err != nil || wd < 0 {
				fatal("watch-debounce must be a non-negative integer (milliseconds)")
			}
			opts.watchDebounce = wd
			i += 2
		case "--watch-ignore":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			opts.watchIgnore = append(opts.watchIgnore, args[i+1])
			i += 2
		case "--watch-retry":
			if i+1 >= len(args) {
				fatal("missing value for %s", args[i])
			}
			wr, err := strconv.Atoi(args[i+1])
			if err != nil || wr < 0 {
				fatal("watch-retry must be a non-negative integer")
			}
			opts.watchRetry = wr
			i += 2
		case "--scan":
			opts.scanMode = true
			i++
		case "--palette":
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p > 0 {
					opts.paletteCount = p
					i += 2
				} else {
					// Next arg is a file, not a count — use default
					opts.paletteCount = 5
					i++
				}
			} else {
				opts.paletteCount = 5
				i++
			}
		case "--smart-crop":
			if i+1 >= len(args) {
				fatal("missing value for %s (WxH, e.g. 800x600)", args[i])
			}
			parts := strings.SplitN(args[i+1], "x", 2)
			if len(parts) != 2 {
				fatal("smart-crop must be WxH (e.g. 800x600)")
			}
			sw, err1 := strconv.Atoi(parts[0])
			sh, err2 := strconv.Atoi(parts[1])
			if err1 != nil || err2 != nil || sw < 1 || sh < 1 {
				fatal("smart-crop dimensions must be positive integers")
			}
			opts.smartCropWidth = sw
			opts.smartCropHeight = sh
			i += 2
		case "mcp":
			opts.mcpMode = true
			i++
		case "serve":
			addr := ":8080"
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				addr = args[i+1]
				i += 2
			} else {
				i++
			}
			opts.serveAddr = addr
		case "-V", "--version":
			fmt.Println(version.String())
			return nil
		case "-h", "--help":
			printUsage()
			return nil
		default:
			if strings.HasPrefix(args[i], "-") {
				fatal("unknown flag: %s", args[i])
			}
			opts.inputs = append(opts.inputs, args[i])
			i++
		}
	}

	if opts.metadata && opts.stripMetadata {
		fatal("--preserve-metadata and --strip-metadata are mutually exclusive")
	}

	needsInput := !opts.watchMode && opts.configFile == "" && opts.completionSh == "" &&
		opts.serveAddr == "" && len(opts.ssimFiles) == 0 && !opts.mcpMode && !opts.scanMode
	if len(opts.inputs) == 0 && needsInput {
		fatal("no input files or directories specified")
	}

	return opts
}

func printUsage() {
	fmt.Print(`pixshift - Universal image converter

Usage:
  pixshift [options] <files or directories...>
  pixshift serve [addr]
  pixshift mcp
  pixshift --scan [dir]
  pixshift --tree [dir]
  pixshift --dedup [dir]
  pixshift --ssim <file1> <file2>
  pixshift --contact-sheet [dir]
  pixshift --palette [N] <files...>

Conversion options:
  -f, --format <fmt>        Output format: jpg, png, gif, webp, tiff, bmp, heic, avif
  -q, --quality <1-100>     Encoding quality (default: 92)
  -j, --jobs <N>            Parallel workers (default: number of CPUs)
  -o, --output <dir>        Output directory (default: same as input)
  -r, --recursive           Process directories recursively
  -m, --preserve-metadata   Preserve EXIF metadata
  -s, --strip-metadata      Strip all EXIF/GPS metadata from output
  -w, --watch               Watch mode: auto-convert new files
  -c, --config <file>       Rules mode: use YAML config file
      --overwrite           Overwrite existing output files
      --dry-run             Preview what would happen
      --template <pattern>  Output naming template (placeholders: {name}, {ext}, {format})
      --preset <name>       Named preset: web, thumbnail, print, archive

Image transforms:
      --auto-rotate          Auto-rotate based on EXIF orientation
      --smart-crop <WxH>    Smart crop to target dimensions (entropy-based)
      --crop <WxH>           Crop to exact pixel dimensions (e.g. 800x600)
      --crop-ratio <W:H>     Crop to aspect ratio (e.g. 16:9)
      --crop-gravity <pos>   Crop anchor: center, north, south, east, west
      --watermark <text>     Add text watermark
      --watermark-pos <pos>  Watermark position: bottom-right, bottom-left, top-right, top-left, center
      --watermark-opacity <N> Watermark opacity 0.0-1.0 (default: 0.5)
      --watermark-size <N>   Watermark font scale (default: 1.0, e.g. 3.0 = 3x)
      --watermark-color <hex> Watermark text color (default: #FFFFFF)
      --watermark-bg <hex>   Watermark background color (default: #000000)

Filters:
      --grayscale            Convert to grayscale
      --sepia [intensity]    Apply sepia tone (default: 0.8, range: 0.0-1.0)
      --brightness <N>       Adjust brightness (-100 to +100)
      --contrast <N>         Adjust contrast (-100 to +100)
      --sharpen              Apply sharpen filter
      --blur <radius>        Apply blur filter (radius in pixels)
      --invert               Invert colors

Resize:
      --width <N>           Target width (preserves aspect ratio)
      --height <N>          Target height (preserves aspect ratio)
      --max-dim <N>         Max dimension (scale to fit)
      --interpolation <m>   Resize method: nearest, bilinear, catmullrom (default)

Encoding options:
      --progressive          JPEG progressive encoding (reserved for future encoder)
      --png-compression <N>  PNG compression: 0=default, 1=none, 2=fast, 3=best
      --webp-method <N>     WebP encoding method: 0-6 (0=fast, 6=best)
      --lossless             WebP lossless mode

Analysis tools:
      --scan                Scan directory: count images by format with sizes
      --palette [N]         Extract color palette (default: 5 colors)
      --tree                Show directory tree of supported images
      --dedup               Find duplicate images using perceptual hashing
      --dedup-threshold <N> Hamming distance threshold (default: 10)
      --ssim <f1> <f2>      Compare two images using Structural Similarity Index
      --contact-sheet       Generate a contact sheet (thumbnail grid)
      --contact-cols <N>    Contact sheet columns (default: 4)
      --contact-size <N>    Contact sheet thumbnail size in px (default: 200)

Other:
      --backup              Create .bak backup of originals before converting
      --json                Output results as JSON
      --completion <shell>  Generate shell completion (bash, zsh, fish)
  -v, --verbose             Verbose output
  -V, --version             Show version
  -h, --help                Show this help

Watch mode options:
      --watch-debounce <ms> Debounce delay in milliseconds (default: 500)
      --watch-ignore <glob> Ignore files matching glob pattern (repeatable)
      --watch-retry <N>     Retry failed conversions N times with backoff

MCP mode:
  pixshift mcp              Start MCP (Model Context Protocol) server on stdio
                            Exposes tools: convert_image, get_formats, analyze_image, compare_images

Serve mode:
  pixshift serve [addr]     Start HTTP conversion server (default: :8080)
    POST /convert           Upload image with multipart form (file, format, quality)
    GET  /formats           List supported formats
    GET  /health            Health check
      --api-key <key>       Require Bearer token authentication
      --rate-limit <N>      Rate limit: max requests/min per IP (0=off)
      --cors-origins <o>    CORS allowed origins (e.g. "*" or "https://example.com")
      --request-timeout <s> Request timeout in seconds (default: 60)
      --max-upload <MB>     Max upload size in MB (default: 50)

Stdin/stdout:
  cat photo.heic | pixshift -f webp - > photo.webp

Presets:
  web         WebP, q85, max 1920px, strip metadata
  thumbnail   JPEG, q80, max 300px, strip metadata
  print       TIFF, q100, preserve metadata
  archive     PNG, q100, preserve metadata
  (custom presets can be defined in config YAML under "presets:" section)

Examples:
  pixshift photo.heic                            Convert HEIC to JPEG (default)
  pixshift -f webp -q 90 photo.heic             Convert to WebP at quality 90
  pixshift -f png -o converted/ photos/          Batch convert directory to PNG
  pixshift --preset web -o output/ photos/       Use web preset
  pixshift --auto-rotate -f jpg photo.heic       Auto-rotate from EXIF
  pixshift --crop-ratio 16:9 -f webp photo.jpg   Crop to 16:9
  pixshift --watermark "© 2026" -f jpg photos/   Add watermark
  pixshift --grayscale --sharpen -f webp photo.jpg Apply filters
  pixshift --watermark "Test" --watermark-size 3 --watermark-color "#FF0000" -f jpg photo.jpg
  pixshift --png-compression 3 -f png photo.jpg  Best PNG compression
  pixshift --lossless -f webp photo.jpg           Lossless WebP
  pixshift --tree ~/Pictures                     Show image directory tree
  pixshift --dedup ~/Pictures                    Find duplicate images
  pixshift --ssim original.jpg compressed.jpg    Compare image quality
  pixshift --contact-sheet -o output/ photos/    Generate contact sheet
  pixshift --json -f webp photos/                JSON output for scripting
  pixshift --backup -f webp photos/              Backup originals first
  pixshift --scan ~/Pictures                     Scan and count images
  pixshift --palette photos/landscape.jpg        Extract color palette
  pixshift --smart-crop 800x600 -f webp photo.jpg Smart crop to 800x600
  pixshift serve :9090                           Start HTTP server on port 9090
  cat photo.heic | pixshift -f webp - > out.webp Stdin/stdout pipeline
`)
}
