package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
)

// Server provides an HTTP API for image conversion.
type Server struct {
	Addr        string
	Registry    *codec.Registry
	MaxFileSize int64
}

// New creates a Server with the given registry and listen address.
func New(reg *codec.Registry, addr string) *Server {
	return &Server{
		Addr:        addr,
		Registry:    reg,
		MaxFileSize: 50 << 20, // 50 MB
	}
}

// Start starts the HTTP server and blocks until ctx is cancelled.
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/formats", s.handleFormats)
	mux.HandleFunc("/convert", s.handleConvert)

	srv := &http.Server{
		Addr:    s.Addr,
		Handler: mux,
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

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleFormats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
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

func (s *Server) handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(s.MaxFileSize); err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing required field: file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	outputFmt := r.FormValue("format")
	if outputFmt == "" {
		http.Error(w, "missing required field: format", http.StatusBadRequest)
		return
	}

	outFormat, err := codec.ParseFormat(outputFmt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, header.Filename)
	tmpFile, err := os.Create(inputPath)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	if _, err := tmpFile.ReadFrom(file); err != nil {
		tmpFile.Close()
		http.Error(w, "failed to save upload", http.StatusInternalServerError)
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

	if _, _, err := pipe.Execute(job); err != nil {
		http.Error(w, fmt.Sprintf("conversion failed: %v", err), http.StatusInternalServerError)
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
