package server

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
	"github.com/DanielTso/pixshift/internal/codec"
	pixcolor "github.com/DanielTso/pixshift/internal/color"
	"github.com/DanielTso/pixshift/internal/ssim"
)

// handleAPIPalette handles POST /api/v1/palette with API key auth.
func (s *Server) handleAPIPalette(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limits := billing.GetLimits(user.Tier)
	maxSize := int64(limits.MaxFileSizeMB) << 20

	img, _, err := s.decodeUpload(r, "file", maxSize)
	if err != nil {
		writeError(w, err.status, err.code, err.message)
		return
	}

	count := 5
	if v := r.FormValue("count"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 20 {
			count = n
		}
	}

	colors := pixcolor.ExtractPalette(img, count)

	if s.DB != nil {
		_, _ = s.DB.IncrementMonthlyAPIUsage(r.Context(), user.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"colors": colors,
	})
}

// handleAPIAnalyze handles POST /api/v1/analyze with API key auth.
func (s *Server) handleAPIAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limits := billing.GetLimits(user.Tier)
	maxSize := int64(limits.MaxFileSizeMB) << 20

	img, info, err := s.decodeUpload(r, "file", maxSize)
	if err != nil {
		writeError(w, err.status, err.code, err.message)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if s.DB != nil {
		_, _ = s.DB.IncrementMonthlyAPIUsage(r.Context(), user.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"width":        width,
		"height":       height,
		"format":       string(info.format),
		"size":         info.size,
		"aspect_ratio": aspectRatio(width, height),
	})
}

// handleAPICompare handles POST /api/v1/compare with API key auth.
func (s *Server) handleAPICompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limits := billing.GetLimits(user.Tier)
	maxSize := int64(limits.MaxFileSizeMB) << 20

	if e := r.ParseMultipartForm(maxSize * 2); e != nil {
		writeError(w, http.StatusBadRequest, "FILE_TOO_LARGE", "failed to parse form: "+e.Error())
		return
	}

	// Save both files to temp dir
	tmpDir, e := os.MkdirTemp("", "pixshift-compare-*")
	if e != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	defer os.RemoveAll(tmpDir)

	path1, size1, e := saveFormFile(r, "file1", tmpDir, maxSize)
	if e != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", e.Error())
		return
	}
	path2, size2, e := saveFormFile(r, "file2", tmpDir, maxSize)
	if e != nil {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", e.Error())
		return
	}

	score, e := ssim.CompareFiles(path1, path2, s.Registry)
	if e != nil {
		writeError(w, http.StatusInternalServerError, "COMPARE_FAILED", fmt.Sprintf("comparison failed: %v", e))
		return
	}

	// Decode both for dimensions
	img1, e := decodeImageFile(path1, s.Registry)
	if e != nil {
		writeError(w, http.StatusInternalServerError, "DECODE_FAILED", fmt.Sprintf("decode file1: %v", e))
		return
	}
	img2, e := decodeImageFile(path2, s.Registry)
	if e != nil {
		writeError(w, http.StatusInternalServerError, "DECODE_FAILED", fmt.Sprintf("decode file2: %v", e))
		return
	}

	b1 := img1.Bounds()
	b2 := img2.Bounds()

	if s.DB != nil {
		_, _ = s.DB.IncrementMonthlyAPIUsage(r.Context(), user.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ssim":   score,
		"rating": ssim.Rating(score),
		"file1": map[string]interface{}{
			"width":  b1.Dx(),
			"height": b1.Dy(),
			"size":   size1,
		},
		"file2": map[string]interface{}{
			"width":  b2.Dx(),
			"height": b2.Dy(),
			"size":   size2,
		},
	})
}

// handleSimplePalette handles POST /palette in simple mode (no auth).
func (s *Server) handleSimplePalette(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	img, _, err := s.decodeUpload(r, "file", s.MaxFileSize)
	if err != nil {
		writeError(w, err.status, err.code, err.message)
		return
	}

	count := 5
	if v := r.FormValue("count"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 && n <= 20 {
			count = n
		}
	}

	colors := pixcolor.ExtractPalette(img, count)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"colors": colors,
	})
}

// handleSimpleAnalyze handles POST /analyze in simple mode (no auth).
func (s *Server) handleSimpleAnalyze(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	img, info, err := s.decodeUpload(r, "file", s.MaxFileSize)
	if err != nil {
		writeError(w, err.status, err.code, err.message)
		return
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"width":        width,
		"height":       height,
		"format":       string(info.format),
		"size":         info.size,
		"aspect_ratio": aspectRatio(width, height),
	})
}

// uploadInfo holds metadata about an uploaded file.
type uploadInfo struct {
	format codec.Format
	size   int64
}

// uploadError is a structured error for upload processing.
type uploadError struct {
	status  int
	code    string
	message string
}

// decodeUpload parses a multipart form, extracts the named file field, detects
// its format, and decodes the image. Returns the decoded image and file info.
func (s *Server) decodeUpload(r *http.Request, field string, maxSize int64) (image.Image, *uploadInfo, *uploadError) {
	if e := r.ParseMultipartForm(maxSize); e != nil {
		return nil, nil, &uploadError{http.StatusBadRequest, "FILE_TOO_LARGE", "failed to parse form: " + e.Error()}
	}

	file, header, e := r.FormFile(field)
	if e != nil {
		return nil, nil, &uploadError{http.StatusBadRequest, "MISSING_FIELD", "missing required field: " + field}
	}
	defer file.Close()

	if header.Size > maxSize {
		return nil, nil, &uploadError{http.StatusBadRequest, "FILE_TOO_LARGE",
			fmt.Sprintf("file size %d exceeds limit of %d bytes", header.Size, maxSize)}
	}

	format, e := codec.DetectFormat(file, header.Filename)
	if e != nil {
		return nil, nil, &uploadError{http.StatusBadRequest, "INVALID_FORMAT", "unable to detect image format"}
	}

	dec, e := s.Registry.Decoder(format)
	if e != nil {
		return nil, nil, &uploadError{http.StatusBadRequest, "INVALID_FORMAT", fmt.Sprintf("unsupported format: %s", format)}
	}

	img, e := dec.Decode(file)
	if e != nil {
		return nil, nil, &uploadError{http.StatusInternalServerError, "DECODE_FAILED", "failed to decode image"}
	}

	return img, &uploadInfo{format: format, size: header.Size}, nil
}

// saveFormFile extracts a file from the multipart form and saves it to dir.
// Returns the path to the saved file and its size.
func saveFormFile(r *http.Request, field, dir string, maxSize int64) (string, int64, error) {
	file, header, err := r.FormFile(field)
	if err != nil {
		return "", 0, fmt.Errorf("missing required field: %s", field)
	}
	defer file.Close()

	if header.Size > maxSize {
		return "", 0, fmt.Errorf("file size %d exceeds limit of %d bytes", header.Size, maxSize)
	}

	path := filepath.Join(dir, field+"_"+sanitizeFilename(header.Filename))
	out, err := os.Create(path)
	if err != nil {
		return "", 0, fmt.Errorf("internal error")
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", 0, fmt.Errorf("failed to save upload")
	}

	return path, header.Size, nil
}

// decodeImageFile opens and decodes an image file using the registry.
func decodeImageFile(path string, reg *codec.Registry) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	format, err := codec.DetectFormat(f, path)
	if err != nil {
		return nil, err
	}

	dec, err := reg.Decoder(format)
	if err != nil {
		return nil, err
	}

	return dec.Decode(f)
}

// aspectRatio computes the aspect ratio string for given dimensions.
// If the GCD-reduced ratio gives small numbers (both < 100), returns "W:H".
// Otherwise returns the decimal ratio rounded to 2 places.
func aspectRatio(w, h int) string {
	if w == 0 || h == 0 {
		return "0:0"
	}
	g := gcd(w, h)
	rw, rh := w/g, h/g
	if rw <= 100 && rh <= 100 {
		return fmt.Sprintf("%d:%d", rw, rh)
	}
	return fmt.Sprintf("%.2f", float64(w)/float64(h))
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}
