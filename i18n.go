package kuu

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/gin-gonic/gin"
)

// LangMessageMap 语言配置集合
type LangMessageMap map[string]string

// LangMap 语言集合
type LangMap map[string](LangMessageMap)

var lang = "en"
var langs = LangMap{}

func init() {
	On("OnNew", func(args ...interface{}) {
		k := args[0].(*Kuu)
		var config LangMap
		if k.Config["i18n"] != nil {
			config = k.Config["i18n"].(LangMap)
		}
		if config != nil {
			for key, value := range config {
				langs[key] = value
			}
		}
	})
}

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

// L 国际化函数(*gin.Context, key, data, lang)
func L(args ...interface{}) string {
	var (
		c    *gin.Context
		key  string
		data H
		lang string
	)
	if len(args) > 1 {
		if args[0] != nil {
			c = args[0].(*gin.Context)
		}
		if args[1] != nil {
			key = args[1].(string)
		}
	}
	if len(args) > 2 {
		if args[2] != nil {
			lang = args[2].(string)
		}
	}
	if len(args) > 3 {
		if args[3] != nil {
			data = args[3].(H)
		}
	}
	if c != nil {
		// 解析Accept-Language
		if l := parseAcceptLanguage(c); l != "" {
			lang = l
		}
	}
	return localeMessage(key, lang, data)
}

// SafeL 包含默认值的L函数
func SafeL(defaultMessages map[string]string, args ...interface{}) string {
	value := L(args...)
	key := args[1].(string)
	if value == "" || value == key {
		value = defaultMessages[key]
	}
	return value
}

func localeMessage(key string, l string, data H) string {
	if l == "" {
		l = lang
	}
	if key == "" || langs[l] == nil || langs[l][key] == "" {
		return key
	}

	value := langs[l][key]
	var buf bytes.Buffer
	tmpl, err := template.New(lang).Parse(value)
	if err != nil {
		return key
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return key
	}
	return buf.String()
}

// SetLang 切换当前语言
func SetLang(l string) {
	lang = l
}

// AddLang 添加语言配置
func AddLang(l string, data LangMessageMap) {
	if l != "" && data != nil {
		langs[l] = data
	}
}

// L 应用实例函数
func (k *Kuu) L(args ...interface{}) string {
	return L(args...)
}
