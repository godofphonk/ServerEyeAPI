package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// GenerateServerID generates a unique server ID
func GenerateServerID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("srv_%s", hex.EncodeToString(b))
}

// GenerateServerKey generates a unique server key
func GenerateServerKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("key_%s", hex.EncodeToString(b))
}

// GenerateSecretKey generates a unique secret key
func GenerateSecretKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("secret_%s", hex.EncodeToString(b))
}
