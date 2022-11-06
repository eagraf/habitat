package sources

import (
	"fmt"
	"hash/maphash"
)

func HashBytes(b []byte) string {
	var h maphash.Hash
	h.Write(b)
	str := fmt.Sprintf("%#x", h.Sum64())
	return str[2:]
}
