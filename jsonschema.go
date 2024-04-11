package ortfodb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

var reflector = jsonschema.Reflector{
	FieldNameTag: "yaml",
	KeyNamer:     func(key string) string { return strings.ToLower(key) },
}

func setSchemaId(schema *jsonschema.Schema) {
	parts := strings.Split(string(schema.ID), "/")
	base := parts[len(parts)-1]
	schema.ID = jsonschema.ID(fmt.Sprintf("https://raw.githubusercontent.com/ortfo/db/v%s/schemas/%s.schema.json", Version, base))
}

func makeJSONSchema(t any) string {
	schema := reflector.Reflect(t)
	setSchemaId(schema)
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

func ConfigurationJSONSchema() string {
	return makeJSONSchema(&Configuration{})
}

func DatabaseJSONSchema() string {
	return makeJSONSchema(&Database{})
}

type tags []Tag

func TagsRepositoryJSONSchema() string {
	return makeJSONSchema(&tags{})
}

type technologies []Technology

func TechnologiesRepositoryJSONSchema() string {
	return makeJSONSchema(&technologies{})
}
