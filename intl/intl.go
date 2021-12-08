package intl

import (
	"bytes"
	"regexp"
	"text/template"
)

func FormatMessage(content string, contextValues ...interface{}) string {
	if content != "" {
		reg := regexp.MustCompile(`{{ *([\w\d]+) *}}`)
		content = reg.ReplaceAllString(content, "{{.$1}}")

		var values interface{}
		if len(contextValues) > 0 && contextValues[0] != nil {
			values = contextValues[0]
		}

		v := template.Must(template.New("").Parse(content))
		var b bytes.Buffer
		if err := v.Execute(&b, values); err == nil {
			return b.String()
		}
	}
	return content
}
