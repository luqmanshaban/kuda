package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

func WriteJson(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)	
}

func GenerateBatchId() string {
	b := make([]byte, 24)
	rand.Read(b)
	uuid := hex.EncodeToString(b)
	return uuid
}