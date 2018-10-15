package kuu

import (
	"bytes"
	"html/template"
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
		key  string
		data H
		lang string
	)
	switch len(args) {
	case 1:
		key = args[0].(string)
	case 2:
		key = args[0].(string)
		if args[1] != nil {
			lang = args[1].(string)
		}
	case 3:
		key = args[0].(string)
		if args[1] != nil {
			lang = args[1].(string)
		}
		if args[2] != nil {
			data = args[2].(H)
		}
	}
	return Locale(key, lang, data)
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
