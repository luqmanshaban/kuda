package utils

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateAPIKey() string {
	b := make([]byte, 24)
	rand.Read(b)
	return "kuda"+hex.EncodeToString(b)
}