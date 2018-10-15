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

// LocaleHook 国际化钩子
type LocaleHook struct {
	Fired bool
}

// Fire 国际化钩子触发函数
func (hook *LocaleHook) Fire(k *Kuu, args ...interface{}) error {
	var config LangMap
	if k.Config["i18n"] != nil {
		config = k.Config["i18n"].(LangMap)
	}
	if config != nil {
		for key, value := range config {
			langs[key] = value
		}
	}
	hook.Fired = true
	return nil
}

func init() {
	AddHook("OnNew", new(LocaleHook))
}

// 加载语言库的方式：
// 1.应用配置
// 2.HTTP

// L Locale的快捷调用
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
	return Locale(key, lang, data)
}

// parseAcceptLanguage 解析Accept-Language并转换成lang
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

// Locale 获取指定语言的国际化内容
func Locale(key string, l string, data H) string {
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
