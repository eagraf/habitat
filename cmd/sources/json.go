package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
)

/*
	JSON files implementation for sources. (key-value pairs)
*/

// Shared functions
func getPath(basePath string, id SourceID) string {
	p := filepath.Join(basePath, string(id)+".json")
	return p
}

// JSON source reader
type JSONReaderWriter struct {
	ctx  context.Context
	Path string // base path to the files
}

func NewJSONReaderWriter(ctx context.Context, path string) *JSONReaderWriter {
	err := os.MkdirAll(path, fs.ModeDir)
	if err != nil {
		log.Error().Msgf("error creating sources path: %s", err.Error())
	}
	return &JSONReaderWriter{ctx: ctx, Path: path}
}

func (R *JSONReaderWriter) Read(id SourceID) ([]byte, error) {
	path := getPath(R.Path, id)
	source, err := ReadSource(path)
	if source == nil {
		return nil, err
	}
	return source.Data, err
}

func (W *JSONReaderWriter) Write(id SourceID, sch *jsonschema.Schema, data []byte) error {
	path := getPath(W.Path, id)
	source, err := ReadSource(path)

	if source == nil {
		source = &SourceFile{}
	}

	if err = ValidateSchemaBytes(W.ctx, sch, data); err != nil {
		return fmt.Errorf("validation err: %s", err.Error())
	}

	source.Data = data
	bytes, err := json.Marshal(source)
	if err != nil {
		return fmt.Errorf("error writing to source file: %s", err.Error())
	}

	err = os.WriteFile(path, bytes, os.ModePerm)
	return err
}
