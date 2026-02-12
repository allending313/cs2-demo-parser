package util

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateID() string {
	b := make([]byte, 2)
	rand.Read(b)
	return hex.EncodeToString(b)
}
