package sources

import (
	"context"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
)

/*
	JSON files implementation for sources. (key-value pairs)
*/

// Shared functions
func getPath(basePath string, source Source) string {
	return basePath + source.Name + ".json"
}

// JSON source reader
type JSONReader struct {
	Path string // base path to the files
}

func NewJSONReader(path string) *JSONReader {
	return &JSONReader{Path: path}
}

func (R *JSONReader) Read(req Source) (error, SourceData) {
	path := getPath(R.Path, req)
	bytes, err := os.ReadFile(path)
	return err, bytes
}

// JSON source writer
type JSONWriter struct {
	Path string // base path to the files
}

func NewJSONWriter(path string) *JSONWriter {
	return &JSONWriter{Path: path}
}

func (W *JSONWriter) Write(source Source, data SourceData) error {
	path := getPath(W.Path, source)
	verrs, err := source.Schema.ValidateBytes(context.Background(), data)
	if err != nil {
		log.Error("Error validating schema bytes: %s", err.Error())
	} else if len(verrs) > 0 {
		for _, e := range verrs {
			log.Error("KeyError when validating source data against schema: %s", e.Message)
		}
	}
	err = os.WriteFile(path, data, fs.FileMode(os.O_RDWR))
	return err
}
