package ortfodb

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
)

var yamlReflector = jsonschema.Reflector{
	FieldNameTag: "yaml",
	KeyNamer:     func(key string) string { return strings.ToLower(key) },
}

func setSchemaId(schema *jsonschema.Schema) {
	parts := strings.Split(string(schema.ID), "/")
	base := parts[len(parts)-1]
	schema.ID = jsonschema.ID(fmt.Sprintf("https://raw.githubusercontent.com/ortfo/db/v%s/schemas/%s.schema.json", Version, base))
}

func makeJSONSchema(t any, yaml bool) string {
	selectedReflector := jsonschema.Reflector{}
	if yaml {
		selectedReflector = yamlReflector
	}
	schema := selectedReflector.Reflect(t)
	setSchemaId(schema)
	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

func ConfigurationJSONSchema() string {
	return makeJSONSchema(&Configuration{}, true)
}

func DatabaseJSONSchema() string {
	return makeJSONSchema(&Database{}, false)
}

type tags []Tag

func TagsRepositoryJSONSchema() string {
	return makeJSONSchema(&tags{}, true)
}

type technologies []Technology

func TechnologiesRepositoryJSONSchema() string {
	return makeJSONSchema(&technologies{}, true)
}
