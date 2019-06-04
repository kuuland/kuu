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
	Paths      map[string](DocPathItems)
	Components DocComponent
}

type DocPathItems map[string]DocPathItem

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
	Title      string
	Type       string
	Ref        string `yaml:"$ref"`
	Properties *DocPathSchema
	Items      *DocSchemaProperty
	Format     string
	Enum       []interface{}
	Default    interface{}
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
	Url         string
	Description string
}

type DocTag struct {
	Name        string
	Description string
}

type DocPathItem struct {
	Tags        []string
	Summary     string
	Description string
	OperationID string             `yaml:"operationId"`
	RequestBody DocPathRequestBody `yaml:"requestBody"`
	Responses   map[int]DocPathResponse
	Parameters  []DocPathParameter
	Deprecated  bool
	Security    []DocPathItemSecurity
}

type DocPathItemSecurity map[string]([]string)

type DocSecurityScheme struct {
	Type string
	Name string
	In   string
}

type DocPathParameter struct {
	Name            string
	In              string
	Description     string
	Required        bool
	Style           string
	Explode         bool
	Schema          DocPathSchema
	AllowEmptyValue bool `yaml:"allowEmptyValue"`
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
	Type        string
	Description string
	Ref         string `yaml:"$ref"`
	Properties  map[string]DocPathSchema
	Items       *DocPathSchema
	Enum        []interface{}
	Default     interface{}
	Required    bool
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
	result = strings.ReplaceAll(result, "items: null", "")
	result = strings.ReplaceAll(result, "email: \"\"", "")
	result = strings.ReplaceAll(result, "description: \"\"", "")
	result = strings.ReplaceAll(result, "properties: null", "")
	result = strings.ReplaceAll(result, "properties: {}", "")
	result = strings.ReplaceAll(result, "content: {}", "")
	result = strings.ReplaceAll(result, "style: \"\"", "")
	result = strings.ReplaceAll(result, "explode: false", "")
	result = strings.ReplaceAll(result, "required: false", "")
	result = strings.ReplaceAll(result, "deprecated: false", "")
	result = strings.ReplaceAll(result, "allowEmptyValue: false", "")
	result = regexp.MustCompile(`(\s*.*)\s*\n\s*\n`).ReplaceAllString(result, "$1\n")
	result = regexp.MustCompile(`\s*requestBody.*\n(\s*responses.*)`).ReplaceAllString(result, "\n$1")
	return
}
