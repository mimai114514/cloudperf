package db

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func NewID(prefix string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(buf))
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
