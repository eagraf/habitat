package sources

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

/*
	JSON files implementation for sources. (key-value pairs)
*/

// Shared functions
func getPath(basePath string, source Source) string {
	return filepath.Join(basePath, source.Name+".json")
}

// JSON source reader
type JSONReader struct {
	Path string // base path to the files
}

func NewJSONReader(path string) *JSONReader {
	return &JSONReader{Path: path}
}

func (R *JSONReader) Read(req Source) (SourceData, error) {
	path := getPath(R.Path, req)
	bytes, err := os.ReadFile(path)
	return SourceData(bytes), err
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
	verrs, err := source.Schema.ValidateBytes(context.Background(), []byte(data))
	if err != nil {
		log.Error().Msgf("Error validating schema bytes: %s", err.Error())
	} else if len(verrs) > 0 {
		for _, e := range verrs {
			log.Error().Msgf("KeyError when validating source data against schema: %s", e.Error())
		}
		return errors.New("Unable to validate schema")
	}
	// TODO: is this the right permissions?
	err = os.WriteFile(path, []byte(data), fs.FileMode(0644))
	return err
}
