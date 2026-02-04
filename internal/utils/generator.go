// Package utils provides utility functions for the application.
package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// MaxAttempts defines the maximum number of attempts to generate a unique short URL.
const MaxAttempts = 10

// tamerLength defines the length of the generated short URL.
const tamerLength = 8

// GenerateShortURL generates a random short URL string.
func GenerateShortURL() string {
	buffer := make([]byte, tamerLength)
	rand.Read(buffer)
	tamer := base64.RawURLEncoding.EncodeToString(buffer)
	return tamer[:tamerLength]
}
