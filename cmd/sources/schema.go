package sources

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/qri-io/jsonschema"
)

/*
Basic types about schema, shared types.
*/
func ReadSchemaFromPath(path string) (*jsonschema.Schema, error) {

	// schema doesn't exist
	if _, err := os.Stat(path); err != nil {
		return nil, nil
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	sch := &jsonschema.Schema{}
	if err = sch.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}

	return sch, err
}

func WriteSchemaToPath(path string, sch *jsonschema.Schema) error {
	bytes, err := sch.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshalling: %s", err.Error())
	}
	err = os.WriteFile(path, bytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("error writing to file: %s", err.Error())
	}
	return nil
}

func ValidateSchemaBytes(ctx context.Context, sch *jsonschema.Schema, data []byte) error {
	kerr, err := sch.ValidateBytes(ctx, data)
	es := make([]string, len(kerr))
	for i, e := range kerr {
		es[i] = e.Error()
	}
	if len(kerr) > 0 {
		return fmt.Errorf("key errors: %s", strings.Join(es, ","))
	}
	return err
}

func GetSchemaId(sch *jsonschema.Schema) string {
	if idprop := sch.JSONProp("$id"); idprop != nil {
		return string(*idprop.(*jsonschema.ID))
	}
	return ""
}

func CheckSchemaId(sch *jsonschema.Schema, id string) bool {
	sid := GetSchemaId(sch)
	if sid != "" && sid == id {
		return true
	}
	return false
}

type SchemaRegistry struct {
	Path string // path to where schemas are stored
}

func NewSchemaRegistry(path string) *SchemaRegistry {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		panic(err)
	}
	if _, err = os.Stat(path); err != nil {
		panic(fmt.Sprintf("no directory at %s", path))
	}
	return &SchemaRegistry{Path: path}
}

func (sr *SchemaRegistry) Lookup(id string) (*jsonschema.Schema, error) {
	sch, err := ReadSchemaFromPath(filepath.Join(sr.Path, id+".json"))
	if err != nil {
		return nil, fmt.Errorf("error reading schema from path: %s", err.Error())
	}

	if sch != nil && !CheckSchemaId(sch, id) {
		return nil, fmt.Errorf("id supplied (%s) and schema id don't match", id)
	}
	return sch, nil

}

func (sr *SchemaRegistry) Add(id string, sch *jsonschema.Schema) error {
	if id == "" {
		return fmt.Errorf("nil id supplied")
	}
	return WriteSchemaToPath(filepath.Join(sr.Path, id+".json"), sch)
}

func (sr *SchemaRegistry) Delete(id string) error {
	return os.Remove(filepath.Join(sr.Path, id+".json"))
}

// TODO: NewSchema()
// functions for evolving schema
/*
func NewSchema(schema []byte) (error, jsonschema.Schema) {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schema, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}
*/
