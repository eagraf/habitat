package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/qri-io/jsonschema"
	"github.com/rs/zerolog/log"
)

type Schema struct {
	Schema      *jsonschema.Schema `json:"schema"` // the jsonschema.Schema
	ID          string             `json:"id"`     // base64 encoded jsonschema $id field
	Name        string             `json:"name"`   // for easy access
	Description string             `json:"desc"`   // for easy access
}

func (s *Schema) JsonSchema() *jsonschema.Schema {
	// TODO: don't panic or recover() if schema parsing fails
	return s.Schema
}

type SchemaStore interface {
	Add(Schema) error
	Get(string) (Schema, error)
	Delete(string) error
}

// store Schemas locally
type LocalSchemaStore struct {
	path string
}

func NewLocalSchemaStore(path string) *LocalSchemaStore {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		log.Fatal().Msgf("error creating schema store path: %s", err.Error())
	}
	return &LocalSchemaStore{
		path: path,
	}
}

func (s *LocalSchemaStore) Add(sch *Schema) error {

	bytes, err := json.Marshal(sch)
	if err != nil {
		return err
	}

	path := getSourcePath(s.path, GetSchemaIdRaw(sch))
	err = os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		log.Error().Msgf("error writing schema to path %s: %s", path, err.Error())
	} else {
		log.Info().Msgf("wrote schema to path %s", path)
	}
	return err
}

func (s *LocalSchemaStore) Get(id string) (*Schema, error) {
	path := getSourcePath(s.path, id)

	// TODO: schema must be explicitly added through schema store: add support in CLI
	// schema doesn't exist - right now just write it and continue
	if _, err := os.Stat(path); err != nil {
		// TODO: properly indicate the schema doesn't exist
		return nil, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sch Schema
	if err = json.Unmarshal(bytes, &sch); err != nil {
		return nil, err
	}

	return &sch, err
}

func (s *LocalSchemaStore) Resolve(ctx context.Context, id string) *Schema {
	// TODO handle errors when .Get() fails
	return &Schema{
		Schema: jsonschema.GetSchemaRegistry().Get(ctx, id),
		ID:     id,
	}
}

func (s *LocalSchemaStore) Delete(id string) error {
	return os.Remove(getSourcePath(s.path, id))
}

func ValidateSchemaBytes(ctx context.Context, sch *Schema, data []byte) error {
	jsonsch := sch.JsonSchema()
	kerr, err := jsonsch.ValidateBytes(ctx, data)
	es := make([]string, len(kerr))
	for i, e := range kerr {
		es[i] = e.Error()
	}
	if len(kerr) > 0 {
		return fmt.Errorf("key errors: %s", strings.Join(es, ","))
	}
	return err
}

func GetSchemaIdRaw(sch *Schema) string {
	jsonsch := sch.JsonSchema()

	if idprop := jsonsch.JSONProp("$id"); idprop != nil {
		return string(*idprop.(*jsonschema.ID))
	}
	return ""
}

func CheckSchemaId(sch *Schema, id string) bool {
	sid := GetSchemaIdRaw(sch)
	if sid != "" && sid == id {
		return true
	}
	return false
}
