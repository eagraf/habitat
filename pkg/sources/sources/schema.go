package sources

import "github.com/qri-io/jsonschema"

/*
Basic types about schema and sources, shared types.
*/

type Schema jsonschema.Schema // use library that supplies validator
type SourceData []byte        // maybe this will evolve into a more fancy type in the future - for now json data stoed as bytes

type Source struct {
	// name and description are here just for easy access rather than getting it from Schema
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schema      jsonschema.Schema `json:"schema"`
	Data        []byte            `json:"data"` // Data should always be validated against Schame, store in []byte form for ease
}

// unused right now, use later for token
type RequestToken struct {
	Token string `json:"token"`
}

// TODO: NewSchema()
// functions for evolving schema
/*
func NewSchema(schema []byte) (err, jsonschema.Schema) {
	rs := &jsonschema.Schema{}
	if err := json.Unmarshal(schema, rs); err != nil {
		panic("unmarshal schema: " + err.Error())
	}
}

func ValidateAgainstSchema() {

}
*/
