package utils

import (
	"crypto/rand"
	"encoding/base64"
)

const MaxAttempts = 10
const tamerLength = 8

func GenerateShortURL() string {
	buffer := make([]byte, tamerLength)
	rand.Read(buffer)
	tamer := base64.RawURLEncoding.EncodeToString(buffer)
	return tamer[:tamerLength]
}
