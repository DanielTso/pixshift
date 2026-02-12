package server

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DanielTso/pixshift/internal/codec"
)

func newTestServer() *Server {
	reg := codec.DefaultRegistry()
	return New(reg, ":0")
}

func TestHandleHealth_GET_Returns200(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("status = %q, want %q", body["status"], "ok")
	}
}

func TestHandleHealth_POST_Returns405(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()

	srv.handleHealth(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleFormats_GET_ReturnsJSON(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/formats", nil)
	w := httptest.NewRecorder()

	srv.handleFormats(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	if len(body["decode"]) == 0 {
		t.Error("decode formats should not be empty")
	}
	if len(body["encode"]) == 0 {
		t.Error("encode formats should not be empty")
	}
}

func TestHandleFormats_POST_Returns405(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/formats", nil)
	w := httptest.NewRecorder()

	srv.handleFormats(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleFormats_ContainsJPEG(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/formats", nil)
	w := httptest.NewRecorder()

	srv.handleFormats(w, req)

	var body map[string][]string
	if err := json.NewDecoder(w.Result().Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}

	found := false
	for _, f := range body["decode"] {
		if f == "jpeg" {
			found = true
			break
		}
	}
	if !found {
		t.Error("formats should include jpeg in decode list")
	}
}

// createTestJPEG creates a small JPEG image in memory and returns its bytes.
func createTestJPEG() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 25), G: uint8(y * 25), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestHandleConvert_ValidJPEGtoPNG(t *testing.T) {
	srv := newTestServer()

	jpegData, err := createTestJPEG()
	if err != nil {
		t.Fatalf("create test jpeg: %v", err)
	}

	// Build multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "test.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(jpegData); err != nil {
		t.Fatalf("write jpeg data: %v", err)
	}
	if err := writer.WriteField("format", "png"); err != nil {
		t.Fatalf("write format field: %v", err)
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200, body: %s", resp.StatusCode, string(respBody))
	}

	// Verify response has content
	respData, _ := io.ReadAll(resp.Body)
	if len(respData) == 0 {
		t.Error("response body should not be empty")
	}
}

func TestHandleConvert_GET_Returns405(t *testing.T) {
	srv := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/convert", nil)
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", w.Code)
	}
}

func TestHandleConvert_MissingFile_Returns400(t *testing.T) {
	srv := newTestServer()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("format", "png")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleConvert_MissingFormat_Returns400(t *testing.T) {
	srv := newTestServer()

	jpegData, err := createTestJPEG()
	if err != nil {
		t.Fatalf("create test jpeg: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	_, _ = part.Write(jpegData)
	// Intentionally omit format field
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandleConvert_InvalidFormat_Returns400(t *testing.T) {
	srv := newTestServer()

	jpegData, err := createTestJPEG()
	if err != nil {
		t.Fatalf("create test jpeg: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	_, _ = part.Write(jpegData)
	_ = writer.WriteField("format", "notaformat")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestNew_Defaults(t *testing.T) {
	reg := codec.DefaultRegistry()
	srv := New(reg, ":8080")
	if srv.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", srv.Addr)
	}
	if srv.Registry != reg {
		t.Error("Registry should match the provided registry")
	}
	if srv.MaxFileSize != 50<<20 {
		t.Errorf("MaxFileSize = %d, want %d", srv.MaxFileSize, 50<<20)
	}
}

func TestHandleConvert_WithQuality(t *testing.T) {
	srv := newTestServer()

	jpegData, err := createTestJPEG()
	if err != nil {
		t.Fatalf("create test jpeg: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, _ := writer.CreateFormFile("file", "test.jpg")
	_, _ = part.Write(jpegData)
	_ = writer.WriteField("format", "jpeg")
	_ = writer.WriteField("quality", "50")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/convert", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	srv.handleConvert(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 200, body: %s", resp.StatusCode, string(respBody))
	}
}

func TestServer_Integration_HTTPTest(t *testing.T) {
	srv := newTestServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/formats", srv.handleFormats)
	mux.HandleFunc("/convert", srv.handleConvert)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	// Test health endpoint via httptest.NewServer
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health: status = %d, want 200", resp.StatusCode)
	}

	var health map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health["status"] != "ok" {
		t.Errorf("health status = %q, want ok", health["status"])
	}
}
