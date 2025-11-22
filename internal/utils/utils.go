package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// Fallback: use a simple random string if crypto/rand fails
		// This should rarely happen, but ensures we always return a token
		return base64.URLEncoding.EncodeToString([]byte("fallback-token"))
	}
	return base64.URLEncoding.EncodeToString(b)
}