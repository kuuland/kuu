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

type Translations map[string]*MessageDescriptor

func FormatMessage(translations Translations, descriptor *MessageDescriptor, contextValues ...interface{}) string {
	if len(translations) > 0 {
		if v, has := translations[descriptor.ID]; has && v != nil {
			descriptor = v
		}
	}
	reg := regexp.MustCompile(`{ *([\w\d]+) *}`)
	str := descriptor.DefaultMessage
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
	return descriptor.ID
}
