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
	Schema      []byte `json:"schema"`    // the jsonschema.Schema
	B64ID       string `json:"base64_id"` // base64 encoded jsonschema $id field
	Name        string `json:"name"`      // for easy access
	Description string `json:"desc"`      // for easy access
}

func (s *Schema) JsonSchema() *jsonschema.Schema {
	// TODO: don't panic or recover() if schema parsing fails
	return jsonschema.Must(string(s.Schema))
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

	err = os.WriteFile(getPath(s.path, sch.B64ID), bytes, os.ModePerm)
	if err != nil {
		log.Error().Msgf("error writing schema to path %s: %s", getPath(s.path, sch.B64ID), err.Error())
	} else {
		log.Info().Msgf("error wrote schema to path %s", getPath(s.path, sch.B64ID))
	}
	return err
}

func (s *LocalSchemaStore) Get(id string) (*Schema, error) {

	path := getPath(s.path, id)

	// schema doesn't exist
	if _, err := os.Stat(path); err != nil {
		return nil, err
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

func (s *LocalSchemaStore) Delete(id string) error {
	return os.Remove(getPath(s.path, id))
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
