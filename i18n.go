package kuu

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// DefaultLang 默认国际化语言编码
	DefaultLang = "en"
	// Langs 国际化语言集合
	Langs = map[string](LangMessages){}
)

func init() {
	On("OnNew", func(args ...interface{}) {
		k := args[0].(*Kuu)
		var config = map[string](LangMessages){}
		if k.Config["i18n"] != nil {
			if b, e := json.Marshal(k.Config["i18n"]); e == nil {
				json.Unmarshal(b, &config)
			}
		}
		if config != nil {
			for key, value := range config {
				Langs[key] = value
			}
		}
	})
}

// LangMessages 语言消息集合
type LangMessages map[string]string

func parseAcceptLanguage(c *gin.Context) string {
	header := c.GetHeader("Accept-Language")
	split := strings.Split(header, ",")
	// zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7
	for _, item := range split {
		item = strings.TrimSpace(item)
		s := strings.TrimSpace(strings.Split(item, ";")[0])
		return s
	}
	return ""
}

// L 获取国际化信息值，可选参数为语言编码和模板数据：(language string, data H)
func L(c *gin.Context, key string, args ...interface{}) string {
	var (
		language string
		data     H
	)
	if len(args) > 0 {
		if args[0] != nil {
			language = args[0].(string)
		}
	}
	if len(args) > 1 {
		if args[1] != nil {
			data = args[1].(H)
		}
	}
	if language == "" && c != nil {
		language = parseAcceptLanguage(c)
	}
	return localeMessage(key, language, data)
}

// SafeL 包含默认值的L函数
func SafeL(defaultMessages map[string]string, c *gin.Context, key string, args ...interface{}) string {
	value := L(c, key, args...)
	if (value == "" || value == key) && defaultMessages[key] != "" {
		value = defaultMessages[key]
	}
	return value
}

func localeMessage(key string, l string, data H) string {
	if l == "" {
		l = DefaultLang
	}
	if key == "" || Langs[l] == nil || Langs[l][key] == "" {
		return key
	}

	value := Langs[l][key]
	var buf bytes.Buffer
	tmpl, err := template.New(l).Parse(value)
	if err != nil {
		return key
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return key
	}
	return buf.String()
}
