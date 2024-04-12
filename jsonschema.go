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

func makeJSONSchema(t any) string {
	schema := yamlReflector.Reflect(t)
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
	schema := jsonschema.Reflect(&Database{})
	setSchemaId(schema)

	metaWorkProperties := jsonschema.NewProperties()
	metaWorkProperties.Set("Partial", &jsonschema.Schema{
		Type: "boolean",
	})

	schema.Definitions["MetaWork"] = &jsonschema.Schema{
		Type:       "object",
		Properties: metaWorkProperties,
	}

	dbWithWorkProperties := jsonschema.NewProperties()
	dbWithWorkProperties.Set("#meta", &jsonschema.Schema{
		Ref: "#/$defs/MetaWork",
	})

	schema.Definitions["DatabaseWithMetaWork"] = &jsonschema.Schema{
		Type:       "object",
		Properties: dbWithWorkProperties,
		PatternProperties: map[string]*jsonschema.Schema{
			"^(?!#meta).*$": {
				Ref: "#/$defs/AnalyzedWork",
			},
		},
	}
	schema.Ref = "#/$defs/DatabaseWithMetaWork"

	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(out)
}

type tags []Tag

func TagsRepositoryJSONSchema() string {
	return makeJSONSchema(&tags{})
}

type technologies []Technology

func TechnologiesRepositoryJSONSchema() string {
	return makeJSONSchema(&technologies{})
}
