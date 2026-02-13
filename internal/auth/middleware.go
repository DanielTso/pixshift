package auth

import (
	"context"
	"net/http"

	"github.com/DanielTso/pixshift/internal/db"
)

type contextKey string

const (
	userContextKey   contextKey = "user"
	apiKeyContextKey contextKey = "apikey"
)

// UserFromContext extracts the authenticated user from the request context.
func UserFromContext(ctx context.Context) *db.User {
	u, _ := ctx.Value(userContextKey).(*db.User)
	return u
}

// APIKeyFromContext extracts the API key from the request context.
func APIKeyFromContext(ctx context.Context) *db.APIKey {
	k, _ := ctx.Value(apiKeyContextKey).(*db.APIKey)
	return k
}

// RequireSession returns middleware that validates the "session" cookie against
// the database and places the user into the request context. Returns 401 if
// the cookie is missing or the session is invalid/expired.
func RequireSession(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			_, user, err := database.GetSession(r.Context(), cookie.Value)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAPIKey returns middleware that validates the "X-API-Key" header. The
// key is hashed and looked up in the database. Returns 401 if the header is
// missing or the key is invalid/revoked.
func RequireAPIKey(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("X-API-Key")
			if raw == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			hash := HashAPIKey(raw)
			key, user, err := database.GetAPIKeyByHash(r.Context(), hash)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			ctx = context.WithValue(ctx, apiKeyContextKey, key)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalSession returns middleware that validates the "session" cookie if
// present. If the cookie is missing or invalid the request proceeds without a
// user in the context.
func OptionalSession(database *db.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			_, user, err := database.GetSession(r.Context(), cookie.Value)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
