package intl

import (
	"bytes"
	"regexp"
	"text/template"
)

type MessageDescriptor struct {
	ID             string `json:"id"`
	DefaultMessage string `json:"defaultMessage"`
	Description    string `json:"description"`
}

func FormatMessage(translations map[string]string, id, defaultMessage string, contextValues ...interface{}) string {
	if translations == nil {
		return defaultMessage
	}
	reg := regexp.MustCompile(`{{ *([\w\d]+) *}}`)
	str := defaultMessage
	str = reg.ReplaceAllString(str, "{{.$1}}")

	var values interface{}
	if len(contextValues) > 0 && contextValues[0] != nil {
		values = contextValues[0]
	}

	v := template.Must(template.New("").Parse(str))
	var b bytes.Buffer
	if err := v.Execute(&b, values); err == nil {
		return b.String()
	}
	return id
}
