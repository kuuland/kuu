package kuu

import (
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
)

type Doc struct {
	Openapi    string
	Info       DocInfo
	Servers    []DocServer
	Tags       []DocTag
	Paths      map[string](DocPathItem)
	Components DocComponent
}

type DocPathItem map[string]DocPathDesc

type DocComponent struct {
	Schemas         map[string]DocComponentSchema
	SecuritySchemes map[string]DocSecurityScheme `yaml:"securitySchemes"`
}
type DocComponentSchema struct {
	Required   []string
	Type       string
	Properties map[string]DocSchemaProperty
}

type DocSchemaProperty struct {
	Type    string
	Ref     string `yaml:"$ref"`
	Format  string
	Enum    []interface{}
	Default interface{}
}

type DocInfo struct {
	Title       string
	Description string
	Contact     DocInfoContact
	Version     string
}

type DocInfoContact struct {
	Email string
}

type DocServer struct {
	Url string
}

type DocTag struct {
	Name        string
	Description string
}

type DocPathDesc struct {
	Tags        []string
	Summary     string
	OperationID string             `yaml:"operationId"`
	RequestBody DocPathRequestBody `yaml:"requestBody"`
	Responses   map[int]DocPathResponse
	Parameters  []DocPathParameter
	Deprecated  bool
	Security    map[string]interface{}
}

type DocSecurityScheme struct {
	Type string
	Name string
	In   string
}

type DocPathParameter struct {
	Name        string
	In          string
	Description string
	Required    bool
	Style       string
	Explode     bool
	Schema      DocPathSchema
}

type DocPathRequestBody struct {
	Description string
	Content     map[string]DocPathContentItem
	Required    bool
}

type DocPathContentItem struct {
	Schema DocPathSchema
}

type DocPathSchema struct {
	Type  string
	Ref   string `yaml:"$ref"`
	Items struct {
		Ref  string `yaml:"$ref"`
		Type string
	}
}

type DocPathResponse struct {
	Description string
	Content     map[string]DocPathContentItem
}

// Marshal
func (d *Doc) Marshal() (result string) {
	out, err := yaml.Marshal(d)
	if err != nil {
		ERROR(err)
		return
	}
	result = string(out)
	result = strings.ReplaceAll(result, "$ref: \"\"", "")
	result = strings.ReplaceAll(result, "type: \"\"", "")
	result = strings.ReplaceAll(result, "format: \"\"", "")
	result = strings.ReplaceAll(result, "default: null", "")
	result = strings.ReplaceAll(result, "enum: []", "")
	result = strings.ReplaceAll(result, "items:\n\n", "")
	result = strings.ReplaceAll(result, "parameters: []", "")
	result = strings.ReplaceAll(result, "security: {}", "")
	result = strings.ReplaceAll(result, "required: []", "")
	result = strings.ReplaceAll(result, "email: \"\"", "")
	result = strings.ReplaceAll(result, "deprecated: false", "")
	result = regexp.MustCompile(`(\s*.*)\s*\n\s*\n`).ReplaceAllString(result, "$1\n")
	return
}
