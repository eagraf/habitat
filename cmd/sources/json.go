package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

/*
	JSON files implementation for sources. (key-value pairs)
*/

// Shared functions
func getPath(basePath string, name SourceName) string {
	p := filepath.Join(basePath, string(name)+".json")
	return p
}

// JSON source reader
type JSONReader struct {
	ctx  context.Context
	Path string // base path to the files
}

func NewJSONReader(ctx context.Context, path string) *JSONReader {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error().Msgf("error creating sources path: %s", err.Error())
	}
	return &JSONReader{Path: path}
}

func (R *JSONReader) Read(name SourceName) ([]byte, error) {
	path := getPath(R.Path, name)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var source SourceFile
	if err = json.Unmarshal(bytes, &source); err != nil {
		return nil, err
	}

	return source.Data, err
}

// JSON source writer
type JSONWriter struct {
	ctx  context.Context
	Path string // base path to the files
}

func NewJSONWriter(ctx context.Context, path string) *JSONWriter {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error().Msgf("error creating sources path: %s", err.Error())
	}
	return &JSONWriter{Path: path}
}

func (W *JSONWriter) Write(name SourceName, data []byte) error {
	path := getPath(W.Path, name)
	bytes, err := os.ReadFile(path)

	var source SourceFile
	if err = json.Unmarshal(bytes, &source); err != nil {
		return fmt.Errorf("unable to read source file: %s", err.Error())
	}

	if err = source.ValidateDataAgainstSchema(W.ctx, data); err != nil {
		return fmt.Errorf("validation err: %s", err.Error())
	}

	source.Data = json.RawMessage(string(data))
	if bytes, err = json.Marshal(source); err != nil {
		return fmt.Errorf("error writing to source file: %s", err.Error())
	}

	err = os.WriteFile(path, bytes, fs.FileMode(os.O_RDWR))
	return err
}
