package util

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}
