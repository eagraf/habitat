package sources

import (
	"encoding/base64"
	"path/filepath"
)

// Shared functions

// getSourcePath returns the path to a source given it's schema $id field and the
// base path where sources are stored on the node
// id is the %id field of the Schema
// locally, sources are stored by the base64($id) since file paths != URL paths
func getSourcePath(basePath string, rawid string) string {
	p := filepath.Join(basePath, EncodeId(rawid)+".json")
	return p
}

func EncodeId(id string) string {
	return base64.StdEncoding.EncodeToString([]byte(id))
}

func DecodeId(b64id string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(b64id)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
