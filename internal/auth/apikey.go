package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateAPIKey creates a new API key and returns the display prefix, the full
// key, and the SHA-256 hash of the full key. The full key has the format
// "pxs_" followed by 64 random hex characters (256-bit entropy).
func GenerateAPIKey() (prefix, fullKey, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("generate api key: %w", err)
	}
	hexPart := hex.EncodeToString(b)
	fullKey = "pxs_" + hexPart
	prefix = hexPart[:12]
	hash = HashAPIKey(fullKey)
	return prefix, fullKey, hash, nil
}

// HashAPIKey returns the hex-encoded SHA-256 hash of the given key.
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
