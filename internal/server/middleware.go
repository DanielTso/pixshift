package server

import (
	"net"
	"net/http"

	"github.com/DanielTso/pixshift/internal/auth"
	"github.com/DanielTso/pixshift/internal/billing"
)

// tierRateLimitMiddleware returns middleware that rate-limits using a key and
// per-minute limit returned by keyFunc for each request.
func (s *Server) tierRateLimitMiddleware(limiter *RateLimiter, keyFunc func(*http.Request) (string, int)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, limit := keyFunc(r)
			if limit > 0 && !limiter.AllowN(key, limit) {
				writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// apiKeyRateKey returns the API key prefix and the tier-based rate limit.
func apiKeyRateKey(r *http.Request) (string, int) {
	user := auth.UserFromContext(r.Context())
	key := auth.APIKeyFromContext(r.Context())
	if user == nil || key == nil {
		return clientIP(r), billing.GetLimits(billing.TierFree).RateLimitPerMin
	}
	limits := billing.GetLimits(user.Tier)
	return "apikey:" + key.Prefix, limits.RateLimitPerMin
}

// sessionRateKey returns the session user ID and the tier-based rate limit.
func sessionRateKey(r *http.Request) (string, int) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		return clientIP(r), billing.GetLimits(billing.TierFree).RateLimitPerMin
	}
	limits := billing.GetLimits(user.Tier)
	return "session:" + user.ID, limits.RateLimitPerMin
}

// ipRateKey returns the client IP and a default rate limit for unauthenticated routes.
func ipRateKey(r *http.Request) (string, int) {
	return clientIP(r), 30
}

// clientIP extracts the client IP from RemoteAddr.
func clientIP(r *http.Request) string {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
