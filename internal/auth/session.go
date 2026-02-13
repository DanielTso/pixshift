package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateSessionToken returns a cryptographically random 64-character hex string.
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
