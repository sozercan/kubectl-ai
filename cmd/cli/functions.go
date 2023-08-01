package cli

import (
	"encoding/json"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

type schemaNames struct {
	ResourceName string `json:"resourceName"`
}

var findSchemaNames openai.FunctionDefinition = openai.FunctionDefinition{
	Name:        "findSchemaNames",
	Description: "Get the list of possible fully-namespaced names for a specific Kubernetes resource. E.g. given `Container` return `io.k8s.api.core.v1.Container`. Given `EnvVarSource` return `io.k8s.api.core.v1.EnvVarSource`",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"resourceName": {
				Type:        jsonschema.String,
				Description: "The name of a Kubernetes resource or field.",
			},
		},
		Required: []string{"resourceName"},
	},
}

func (s *schemaNames) Run() (content string, err error) {
	names, err := fetchResourceNames(s.ResourceName)
	if err != nil {
		return "", err
	}

	return strings.Join(names, "\n"), nil
}

type schema struct {
	ResourceType string `json:"resourceType"`
}

var getSchema openai.FunctionDefinition = openai.FunctionDefinition{
	Name:        "getSchema",
	Description: "Get the OpenAPI schema for a Kubernetes resource",
	Parameters: jsonschema.Definition{
		Type: jsonschema.Object,
		Properties: map[string]jsonschema.Definition{
			"resourceType": {
				Type:        jsonschema.String,
				Description: "The type of the Kubernetes resource or object (e.g. subresource). Must be fully namespaced, as returned by findSchemaNames",
			},
		},
		Required: []string{"resourceType"},
	},
}

func (s *schema) Run() (content string, err error) {
	schema, err := fetchSchemaForResource(s.ResourceType)
	if err != nil {
		return "", err
	}

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return "", err
	}

	return string(schemaBytes), nil
}

func funcCall(call *openai.FunctionCall) (string, error) {
	switch call.Name {
	case findSchemaNames.Name:
		var f schemaNames
		if err := json.Unmarshal([]byte(call.Arguments), &f); err != nil {
			return "", err
		}
		return f.Run()
	case getSchema.Name:
		var f schema
		if err := json.Unmarshal([]byte(call.Arguments), &f); err != nil {
			return "", err
		}
		return f.Run()
	}
	return "", nil
}
