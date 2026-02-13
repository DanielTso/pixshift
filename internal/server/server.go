package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/db"
	"github.com/DanielTso/pixshift/internal/pipeline"
)

// Server provides an HTTP API for image conversion.
type Server struct {
	// Existing fields (simple mode)
	Addr         string
	Registry     *codec.Registry
	MaxFileSize  int64
	APIKey       string        // simple mode bearer token
	RateLimit    int           // requests/min per IP (0=off)
	AllowOrigins string        // CORS allowed origins ("*" default)
	Timeout      time.Duration // request timeout (60s default)

	// New fields for full mode (DB != nil)
	DB                  *db.DB
	StripeWebhookSecret string
	OAuthConfig         *oauth2.Config
	SessionSecret       string
	BaseURL             string
	WebFS               fs.FS // embedded SPA filesystem (can be nil)

	// webhookEvents tracks processed Stripe event IDs for idempotency.
	webhookEvents   map[string]time.Time
	webhookEventsMu sync.Mutex
}

// ErrorResponse is the structured JSON error returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

// New creates a Server with the given registry and listen address.
func New(reg *codec.Registry, addr string) *Server {
	return &Server{
		Addr:          addr,
		Registry:      reg,
		MaxFileSize:   50 << 20, // 50 MB
		webhookEvents: make(map[string]time.Time),
	}
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg, Code: code})
}

// statusRecorder wraps http.ResponseWriter to capture the status code for logging.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware logs method, path, status, and duration to stderr as JSON.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(sr, r)
		duration := time.Since(start)
		logEntry := map[string]interface{}{
			"method":   r.Method,
			"path":     r.URL.Path,
			"status":   sr.statusCode,
			"duration": duration.String(),
			"remote":   r.RemoteAddr,
		}
		data, _ := json.Marshal(logEntry)
		fmt.Fprintf(os.Stderr, "%s\n", data)
	})
}

// authMiddleware checks for a valid Authorization: Bearer <key> header.
func authMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") || strings.TrimPrefix(auth, "Bearer ") != apiKey {
				writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid or missing API key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// securityHeadersMiddleware adds standard security headers to all responses.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// HSTS: only set when behind HTTPS (X-Forwarded-Proto or TLS)
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware sets CORS headers and handles OPTIONS preflight.
func corsMiddleware(allowOrigins string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowOrigins == "*" {
				// Wildcard: no credentials allowed per spec
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// Validate origin against whitelist
				origin := r.Header.Get("Origin")
				for _, allowed := range strings.Split(allowOrigins, ",") {
					if strings.TrimSpace(allowed) == origin {
						w.Header().Set("Access-Control-Allow-Origin", origin)
						w.Header().Set("Access-Control-Allow-Credentials", "true")
						break
					}
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "3600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// rateLimitMiddleware checks the per-IP rate limit and returns 429 if exceeded.
func rateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			if ip == "" {
				ip = r.RemoteAddr
			}
			if !limiter.Allow(ip) {
				writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Start starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	var handler http.Handler

	if s.DB != nil {
		handler = s.buildFullModeHandler(ctx)
	} else {
		handler = s.buildSimpleModeHandler(ctx)
	}

	timeout := s.Timeout
	if timeout == 0 {
		timeout = 60 * time.Second
	}

	// Wrap with security headers
	handler = securityHeadersMiddleware(handler)

	srv := &http.Server{
		Addr:              s.Addr,
		Handler:           handler,
		ReadTimeout:       timeout,
		WriteTimeout:      timeout,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		return srv.Shutdown(context.Background())
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

// buildSimpleModeHandler creates the handler for simple mode (no DB, backward compatible).
func (s *Server) buildSimpleModeHandler(ctx context.Context) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/formats", s.handleFormats)
	mux.HandleFunc("/convert", s.handleConvert)
	mux.HandleFunc("/internal/convert", s.handleConvert)
	mux.HandleFunc("/palette", s.handleSimplePalette)
	mux.HandleFunc("/analyze", s.handleSimpleAnalyze)

	var handler http.Handler = mux
	handler = loggingMiddleware(handler)
	if s.APIKey != "" {
		handler = authMiddleware(s.APIKey)(handler)
	}
	if s.AllowOrigins != "" {
		handler = corsMiddleware(s.AllowOrigins)(handler)
	}
	if s.RateLimit > 0 {
		limiter := NewRateLimiter(s.RateLimit)
		go limiter.Cleanup(ctx)
		handler = rateLimitMiddleware(limiter)(handler)
	}
	return handler
}

// buildFullModeHandler creates the handler for full mode (with DB, auth, billing).
func (s *Server) buildFullModeHandler(ctx context.Context) http.Handler {
	mux := http.NewServeMux()

	// Shared rate limiter for full mode
	limiter := NewRateLimiter(60) // default 60/min, overridden per-route by tier
	go limiter.Cleanup(ctx)

	// --- Public routes (no auth) ---
	mux.HandleFunc("/health", s.handleHealth)

	// --- API routes (API key auth) ---
	apiKeyAuth := auth.RequireAPIKey(s.DB)
	apiRateLimit := s.tierRateLimitMiddleware(limiter, apiKeyRateKey)

	mux.Handle("/api/v1/convert", apiRateLimit(apiKeyAuth(http.HandlerFunc(s.handleAPIConvert))))
	mux.Handle("/api/v1/palette", apiRateLimit(apiKeyAuth(http.HandlerFunc(s.handleAPIPalette))))
	mux.Handle("/api/v1/analyze", apiRateLimit(apiKeyAuth(http.HandlerFunc(s.handleAPIAnalyze))))
	mux.Handle("/api/v1/compare", apiRateLimit(apiKeyAuth(http.HandlerFunc(s.handleAPICompare))))
	mux.HandleFunc("/api/v1/formats", s.handleFormats)

	// --- Webhook routes (no auth, signature verified internally) ---
	mux.HandleFunc("/api/webhooks/stripe", s.handleStripeWebhook)

	// --- Web internal routes (session auth) ---
	sessionAuth := auth.RequireSession(s.DB)
	webRateLimit := s.tierRateLimitMiddleware(limiter, sessionRateKey)

	mux.Handle("/internal/convert", webRateLimit(sessionAuth(http.HandlerFunc(s.handleWebConvert))))
	mux.HandleFunc("/internal/formats", s.handleFormats)
	mux.Handle("/internal/user", sessionAuth(http.HandlerFunc(s.handleGetUser)))
	mux.Handle("/internal/keys", sessionAuth(http.HandlerFunc(s.handleKeys)))
	mux.Handle("/internal/keys/", sessionAuth(http.HandlerFunc(s.handleRevokeKey)))
	mux.Handle("/internal/usage", sessionAuth(http.HandlerFunc(s.handleUsage)))

	// --- Auth routes (no auth, rate limited by IP) ---
	authRateLimit := s.tierRateLimitMiddleware(limiter, ipRateKey)

	mux.Handle("/internal/auth/signup", authRateLimit(http.HandlerFunc(s.handleSignup)))
	mux.Handle("/internal/auth/login", authRateLimit(http.HandlerFunc(s.handleLogin)))
	mux.HandleFunc("/internal/auth/logout", s.handleLogout)
	mux.HandleFunc("/internal/auth/google", s.handleGoogleOAuth)
	mux.HandleFunc("/internal/auth/google/callback", s.handleGoogleCallback)

	// --- Billing routes (session auth) ---
	mux.Handle("/internal/billing/checkout", sessionAuth(http.HandlerFunc(s.handleCheckout)))
	mux.Handle("/internal/billing/portal", sessionAuth(http.HandlerFunc(s.handlePortal)))

	// --- SPA handler ---
	if s.WebFS != nil {
		mux.Handle("/", s.spaHandler())
	}

	// Apply global middleware
	var handler http.Handler = mux
	handler = loggingMiddleware(handler)
	if s.AllowOrigins != "" {
		handler = corsMiddleware(s.AllowOrigins)(handler)
	}

	return handler
}

// spaHandler serves the embedded SPA filesystem, falling back to index.html
// for client-side routing.
func (s *Server) spaHandler() http.Handler {
	fileServer := http.FileServer(http.FS(s.WebFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists in the embedded FS
		f, err := s.WebFS.Open(strings.TrimPrefix(path, "/"))
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fall back to index.html for SPA client-side routing
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleFormats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	decode := make([]string, 0)
	for _, f := range s.Registry.SupportedDecoders() {
		decode = append(decode, string(f))
	}
	encode := make([]string, 0)
	for _, f := range s.Registry.SupportedEncoders() {
		encode = append(encode, string(f))
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string][]string{
		"decode": decode,
		"encode": encode,
	})
}

// handleConvert is the simple mode conversion handler (backward compatible).
func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	if err := r.ParseMultipartForm(s.MaxFileSize); err != nil {
		writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "missing required field: file")
		return
	}
	defer file.Close()

	outputFmt := r.FormValue("format")
	if outputFmt == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "missing required field: format")
		return
	}

	outFormat, err := codec.ParseFormat(outputFmt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FORMAT", err.Error())
		return
	}

	quality := 92
	if q := r.FormValue("quality"); q != "" {
		if v, err := strconv.Atoi(q); err == nil && v > 0 && v <= 100 {
			quality = v
		}
	}

	var width, height, maxDim int
	if v := r.FormValue("width"); v != "" {
		width, _ = strconv.Atoi(v)
	}
	if v := r.FormValue("height"); v != "" {
		height, _ = strconv.Atoi(v)
	}
	if v := r.FormValue("max_dim"); v != "" {
		maxDim, _ = strconv.Atoi(v)
	}

	// Save uploaded file to temp dir
	tmpDir, err := os.MkdirTemp("", "pixshift-convert-*")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, header.Filename)
	tmpFile, err := os.Create(inputPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	if _, err := tmpFile.ReadFrom(file); err != nil {
		tmpFile.Close()
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to save upload")
		return
	}
	tmpFile.Close()

	// Build output filename
	baseName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	outExt := codec.DefaultExtension(outFormat)
	outputPath := filepath.Join(tmpDir, baseName+outExt)

	pipe := pipeline.NewPipeline(s.Registry)
	job := pipeline.Job{
		InputPath:    inputPath,
		OutputPath:   outputPath,
		OutputFormat: outFormat,
		Quality:      quality,
		Width:        width,
		Height:       height,
		MaxDim:       maxDim,
	}

	// Boolean form fields
	job.AutoRotate = r.FormValue("auto_rotate") == "true"
	job.Grayscale = r.FormValue("grayscale") == "true"
	job.Sharpen = r.FormValue("sharpen") == "true"
	job.Invert = r.FormValue("invert") == "true"
	job.PreserveMetadata = r.FormValue("preserve_metadata") == "true"
	job.StripMetadata = r.FormValue("strip_metadata") == "true"
	job.BackupOriginal = false // never backup in server mode

	// String form fields
	job.CropAspectRatio = r.FormValue("crop_ratio")
	job.CropGravity = r.FormValue("crop_gravity")
	job.WatermarkText = r.FormValue("watermark_text")
	job.WatermarkPos = r.FormValue("watermark_pos")
	job.WatermarkColor = r.FormValue("watermark_color")
	job.WatermarkBg = r.FormValue("watermark_bg")
	job.Interpolation = r.FormValue("interpolation")

	// Int form fields
	if v := r.FormValue("crop_width"); v != "" {
		job.CropWidth, _ = strconv.Atoi(v)
	}
	if v := r.FormValue("crop_height"); v != "" {
		job.CropHeight, _ = strconv.Atoi(v)
	}

	// Float form fields
	if v := r.FormValue("watermark_opacity"); v != "" {
		job.WatermarkOpacity, _ = strconv.ParseFloat(v, 64)
	}
	if v := r.FormValue("watermark_size"); v != "" {
		job.WatermarkSize, _ = strconv.ParseFloat(v, 64)
	}
	if v := r.FormValue("sepia"); v != "" {
		job.Sepia, _ = strconv.ParseFloat(v, 64)
	}
	if v := r.FormValue("brightness"); v != "" {
		job.Brightness, _ = strconv.ParseFloat(v, 64)
	}
	if v := r.FormValue("contrast"); v != "" {
		job.Contrast, _ = strconv.ParseFloat(v, 64)
	}
	if v := r.FormValue("blur"); v != "" {
		job.Blur, _ = strconv.ParseFloat(v, 64)
	}

	// Encoding options
	job.EncodeOpts.Quality = quality
	job.EncodeOpts.Progressive = r.FormValue("progressive") == "true"
	job.EncodeOpts.Lossless = r.FormValue("lossless") == "true"
	if v := r.FormValue("png_compression"); v != "" {
		job.EncodeOpts.Compression, _ = strconv.Atoi(v)
	}
	if v := r.FormValue("webp_method"); v != "" {
		job.EncodeOpts.WebPMethod, _ = strconv.Atoi(v)
	}

	if _, _, err := pipe.Execute(job); err != nil {
		writeError(w, http.StatusInternalServerError, "CONVERSION_FAILED", fmt.Sprintf("conversion failed: %v", err))
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType(outFormat))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, baseName+outExt))

	http.ServeFile(w, r, outputPath)
}

func contentType(f codec.Format) string {
	switch f {
	case codec.JPEG:
		return "image/jpeg"
	case codec.PNG:
		return "image/png"
	case codec.GIF:
		return "image/gif"
	case codec.WebP:
		return "image/webp"
	case codec.TIFF:
		return "image/tiff"
	case codec.BMP:
		return "image/bmp"
	case codec.HEIC:
		return "image/heic"
	case codec.AVIF:
		return "image/avif"
	default:
		return "application/octet-stream"
	}
}
