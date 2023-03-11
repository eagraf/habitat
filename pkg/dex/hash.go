package dex

import (
	"crypto/sha1"
	"encoding/hex"
)

func hash(buf []byte) string {
	h := sha1.Sum(buf)
	return hex.EncodeToString(h[:])
}
