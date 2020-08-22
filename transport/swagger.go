package transport

import (
	"github.com/go-openapi/spec"
	"github.com/severgroup-tt/gopkg-app/types"
	logger "github.com/severgroup-tt/gopkg-logger"
	"github.com/utrack/clay/v2/transport/swagger"
	"strings"
)

const int64Type = "int64"

// SetIntegerTypeForInt64 replace type="string" to type="integer" for format="int64"
func SetIntegerTypeForInt64() swagger.Option {
	return func(swagger *spec.Swagger) {
		for _, definition := range swagger.Definitions {
			for propName, property := range definition.Properties {
				if property.Format == int64Type {
					definition.Properties[propName] = *spec.Int64Property()
				}
				if property.Items != nil && property.Items.Schema != nil && property.Items.Schema.Format == int64Type {
					definition.Properties[propName].Items.Schema = spec.Int64Property()
				}
			}
		}
	}
}

func SetDeprecatedFromSummary() swagger.Option {
	return func(swagger *spec.Swagger) {
		for _, path := range swagger.Paths.Paths {
			for _, op := range []*spec.Operation{path.Get, path.Post, path.Put, path.Patch, path.Delete} {
				if op != nil {
					if strings.HasPrefix(op.Summary, "DEPRECATED ") {
						op.Summary = op.Summary[11:]
						op.Deprecated = true
					}
				}
			}
		}
	}
}

func SetErrorResponse() swagger.Option {
	return func(swagger *spec.Swagger) {
		data := &spec.Schema{}
		data.CollectionOf(spec.Schema{SchemaProps: spec.SchemaProps{Type: []string{"object"}, Properties: map[string]spec.Schema{
			"key":   {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
			"value": {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
		}}})
		swagger.Definitions["ErrorResponse"] = spec.Schema{
			SchemaProps: spec.SchemaProps{Type: []string{"object"}, Properties: map[string]spec.Schema{
				"error": {SchemaProps: spec.SchemaProps{Type: []string{"object"}, Properties: map[string]spec.Schema{
					"message": {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
					"code":    {SchemaProps: spec.SchemaProps{Type: []string{"string"}}},
					"data":    *data,
				}}},
			}},
		}
		ref, err := spec.NewRef("#/definitions/ErrorResponse")
		if err != nil {
			logger.Error(logger.App, "Swagger errorResponse error: %v", err)
		}
		errorResponse := spec.Response{
			ResponseProps: spec.ResponseProps{
				Description: "A error response.",
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Ref: ref,
					},
				},
			},
		}
		for _, path := range swagger.Paths.Paths {
			for _, op := range []*spec.Operation{path.Get, path.Post, path.Put, path.Patch, path.Delete} {
				if op != nil {
					op.Responses.StatusCodeResponses[500] = errorResponse
				}
			}
		}
	}
}

// Convert CamelCase names to snake_case
func SetNameSnakeCase() swagger.Option {
	return func(swagger *spec.Swagger) {
		for _, definition := range swagger.Definitions {
			for propName, property := range definition.Properties {
				propNameSnakeCase := types.CamelToSnakeCase(propName)
				if propNameSnakeCase != propName {
					definition.Properties[propNameSnakeCase] = property
					delete(definition.Properties, propName)
				}
			}
		}
		for _, path := range swagger.Paths.Paths {
			for _, op := range []*spec.Operation{path.Get, path.Post, path.Put, path.Patch, path.Delete} {
				if op != nil {
					for i, param := range op.Parameters {
						op.Parameters[i].ParamProps.Name = types.CamelToSnakeCase(param.ParamProps.Name)
					}
				}
			}
		}
	}
}
