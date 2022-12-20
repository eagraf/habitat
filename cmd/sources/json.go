package sources

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

/*
	JSON files implementation for sources. (key-value pairs)
*/

// JSON source reader
type JSONReaderWriter struct {
	ctx  context.Context
	Path string // base path to the files
}

func NewJSONReaderWriter(ctx context.Context, path string) *JSONReaderWriter {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		log.Fatal().Msgf("error creating sources path: %s", err.Error())
	}
	return &JSONReaderWriter{ctx: ctx, Path: path}
}

func (rw *JSONReaderWriter) Read(id string) ([]byte, error) {
	path := getPath(rw.Path, id)

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, err
}

func (rw *JSONReaderWriter) Write(id string, sch *Schema, data []byte) error {
	path := getPath(rw.Path, id)

	if err := ValidateSchemaBytes(rw.ctx, sch, data); err != nil {
		return fmt.Errorf("validation err: %s", err.Error())
	}

	err := os.WriteFile(path, data, os.ModePerm)
	if err != nil {
		return err
	}
	log.Info().Msgf("wrote sources to %s", path)
	return nil
}
