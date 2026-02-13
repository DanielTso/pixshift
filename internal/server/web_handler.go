package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
	"github.com/DanielTso/pixshift/internal/db"
)

// handleWebConvert handles POST /internal/convert with session auth.
func (s *Server) handleWebConvert(w http.ResponseWriter, r *http.Request) {
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

	maxSize := int64(limits.MaxFileSizeMB) << 20

	s.executeConvert(w, r, maxSize, user, nil, "web")
}

// handleGetUser handles GET /internal/user.
func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"name":       user.Name,
		"provider":   user.Provider,
		"tier":       user.Tier,
		"created_at": user.CreatedAt,
	})
}

// handleKeys handles GET/POST /internal/keys.
func (s *Server) handleKeys(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.listKeys(w, r, user)
	case http.MethodPost:
		s.createKey(w, r, user)
	default:
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) listKeys(w http.ResponseWriter, r *http.Request, user *db.User) {
	keys, err := s.DB.ListAPIKeys(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	type keyResponse struct {
		ID        string `json:"id"`
		Prefix    string `json:"prefix"`
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
	}

	result := make([]keyResponse, 0, len(keys))
	for _, k := range keys {
		result = append(result, keyResponse{
			ID:        k.ID,
			Prefix:    k.Prefix,
			Name:      k.Name,
			CreatedAt: k.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) createKey(w http.ResponseWriter, r *http.Request, user *db.User) {
	limits := billing.GetLimits(user.Tier)

	count, err := s.DB.CountActiveKeys(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	if count >= limits.MaxAPIKeys {
		writeError(w, http.StatusForbidden, "KEY_LIMIT", "maximum API keys reached for your plan")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}
	if body.Name == "" {
		body.Name = "default"
	}

	prefix, fullKey, hash, err := auth.GenerateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	if _, err := s.DB.CreateAPIKey(r.Context(), user.ID, hash, prefix, body.Name); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"key":    fullKey,
		"prefix": prefix,
		"name":   body.Name,
	})
}

// handleRevokeKey handles DELETE /internal/keys/{id}.
func (s *Server) handleRevokeKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	// Extract key ID from path: /internal/keys/{id}
	keyID := strings.TrimPrefix(r.URL.Path, "/internal/keys/")
	if keyID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELD", "missing key ID")
		return
	}

	if err := s.DB.RevokeAPIKey(r.Context(), keyID, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
}

// handleUsage handles GET /internal/usage.
func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}

	limits := billing.GetLimits(user.Tier)

	usage, err := s.DB.GetDailyUsage(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	apiUsage, _ := s.DB.GetMonthlyAPIUsage(r.Context(), user.ID)

	conversions, err := s.DB.ListConversions(r.Context(), user.ID, 20, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}

	type convResponse struct {
		ID           string `json:"id"`
		InputFormat  string `json:"input_format"`
		OutputFormat string `json:"output_format"`
		InputSize    int64  `json:"input_size"`
		OutputSize   int64  `json:"output_size"`
		DurationMS   int    `json:"duration_ms"`
		Source       string `json:"source"`
		CreatedAt    string `json:"created_at"`
	}

	convList := make([]convResponse, 0, len(conversions))
	for _, c := range conversions {
		convList = append(convList, convResponse{
			ID:           c.ID,
			InputFormat:  c.InputFormat,
			OutputFormat: c.OutputFormat,
			InputSize:    c.InputSize,
			OutputSize:   c.OutputSize,
			DurationMS:   c.DurationMS,
			Source:       c.Source,
			CreatedAt:    c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"daily_count":       usage,
		"daily_limit":       limits.MaxConversionsPerDay,
		"tier":              user.Tier,
		"conversions":       convList,
		"monthly_api_count": apiUsage,
		"monthly_api_limit": limits.MaxAPIRequestsPerMonth,
		"max_file_size_mb":  limits.MaxFileSizeMB,
		"max_batch_size":    limits.MaxBatchSize,
	})
}

