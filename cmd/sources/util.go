package sources

import (
	"encoding/base64"
	"path/filepath"
)

// Shared functions
func getPath(basePath string, id string) string {
	p := filepath.Join(basePath, EncodeId(id)+".json")
	return p
}

func EncodeId(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))

}
