package ortfodb

import (
	"encoding/json"

	ll "github.com/ewen-lbh/label-logger-go"
	"github.com/xeipuuv/gojsonschema"
)

func ValidateAsJSONSchema(typ any, yaml bool, values any) []gojsonschema.ResultError {
	schema := makeJSONSchema(typ, true)
	jsonOpts, _ := json.Marshal(values)
	if jsonOpts == nil {
		return []gojsonschema.ResultError{}
	}
	_, valiationErrors, err := validateWithJSONSchema(string(jsonOpts), schema)
	if err != nil {
		ll.Log("Error", "red", "could not validate as JSON schema: %s", err)
		return []gojsonschema.ResultError{}
	}
	return valiationErrors
}
