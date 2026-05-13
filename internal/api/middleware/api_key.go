package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"os"
)

func InitAPIKey() string {
    key := os.Getenv("KUDA_API_KEY")
    if key != "" {
        return key
    }

    // generate one if not provided
    b := make([]byte, 24)
    rand.Read(b)
    key = "kuda_" + hex.EncodeToString(b)

    slog.Warn("KUDA_API_KEY not set, generated a temporary key",
        "api_key", key,
        "warning", "set this in your environment to make it permanent",
    )
    return key
}