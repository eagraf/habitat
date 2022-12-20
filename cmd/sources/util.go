package sources

import (
	"encoding/base64"
	"path/filepath"
)

// Shared functions
func getPath(basePath string, b64id string) string {
	p := filepath.Join(basePath, string(b64id)+".json")
	return p
}

func EncodeId(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))

}
