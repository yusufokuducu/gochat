package websocket

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID generates a random string that can be used as a unique identifier
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
