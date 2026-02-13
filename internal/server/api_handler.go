package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/db"
	"github.com/DanielTso/pixshift/internal/pipeline"
)

// handleAPIConvert handles POST /api/v1/convert with API key auth.
func (s *Server) handleAPIConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	apiKey := auth.APIKeyFromContext(r.Context())
	if user == nil || apiKey == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limits := billing.GetLimits(user.Tier)

	// Check daily usage limit
	if limits.MaxConversionsPerDay > 0 {
		usage, err := s.DB.GetDailyUsage(r.Context(), user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			return
		}
		if usage >= limits.MaxConversionsPerDay {
			writeError(w, http.StatusTooManyRequests, "DAILY_LIMIT", "daily conversion limit reached")
			return
		}
	}

	// Check monthly API request limit
	if limits.MaxAPIRequestsPerMonth > 0 {
		apiUsage, err := s.DB.GetMonthlyAPIUsage(r.Context(), user.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			return
		}
		if apiUsage >= limits.MaxAPIRequestsPerMonth {
			writeError(w, http.StatusTooManyRequests, "MONTHLY_API_LIMIT", "monthly API request limit reached")
			return
		}
	}

	// Override max file size based on tier
	maxSize := int64(limits.MaxFileSizeMB) << 20

	s.executeConvert(w, r, maxSize, user, apiKey, "api")
}

// executeConvert is the shared conversion logic used by both API and web handlers.
func (s *Server) executeConvert(w http.ResponseWriter, r *http.Request, maxSize int64, user *db.User, apiKey *db.APIKey, source string) {
	start := time.Now()

	if err := r.ParseMultipartForm(maxSize); err != nil {
		writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "failed to parse form: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "missing required field: file")
		return
	}
	defer file.Close()

	// Check file size against tier limit
	if header.Size > maxSize {
		writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE",
			fmt.Sprintf("file size %d exceeds limit of %d bytes", header.Size, maxSize))
		return
	}

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

	// Detect input format from extension
	inputFormat := codec.Format(strings.TrimPrefix(strings.ToLower(filepath.Ext(header.Filename)), "."))

	// Build output filename
	baseName := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))
	outExt := codec.DefaultExtension(outFormat)
	outputPath := filepath.Join(tmpDir, baseName+outExt)

	pipe := pipeline.NewPipeline(s.Registry)
	job := buildConvertJob(r, inputPath, outputPath, outFormat, quality, width, height, maxDim)

	if _, _, err := pipe.Execute(job); err != nil {
		writeError(w, http.StatusInternalServerError, "CONVERSION_FAILED", fmt.Sprintf("conversion failed: %v", err))
		return
	}

	duration := time.Since(start)

	// Record conversion in DB if available
	if s.DB != nil && user != nil {
		outputInfo, _ := os.Stat(outputPath)
		var outputSize int64
		if outputInfo != nil {
			outputSize = outputInfo.Size()
		}

		conv := &db.Conversion{
			UserID:       sql.NullString{String: user.ID, Valid: true},
			InputFormat:  string(inputFormat),
			OutputFormat: string(outFormat),
			InputSize:    header.Size,
			OutputSize:   outputSize,
			DurationMS:   int(duration.Milliseconds()),
			Source:       source,
		}
		if apiKey != nil {
			conv.APIKeyID = sql.NullString{String: apiKey.ID, Valid: true}
		}
		_ = s.DB.RecordConversion(r.Context(), conv)
		_, _ = s.DB.IncrementUsage(r.Context(), user.ID)
		if source == "api" {
			_, _ = s.DB.IncrementMonthlyAPIUsage(r.Context(), user.ID)
		}
	}

	// Set response headers
	w.Header().Set("Content-Type", contentType(outFormat))
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, baseName+outExt))

	http.ServeFile(w, r, outputPath)
}

// buildConvertJob constructs a pipeline.Job from the HTTP form values.
func buildConvertJob(r *http.Request, inputPath, outputPath string, outFormat codec.Format, quality, width, height, maxDim int) pipeline.Job {
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

	return job
}
