package kuu

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
)

// LangMessages 国际化字符串对照表
type LangMessages map[string]string

// Langs 语言配置库
var Langs = map[string]LangMessages{}

func mergeLangs(merge map[string]LangMessages) {
	if merge == nil {
		return
	}
	for lang, messages := range merge {
		if Langs[lang] == nil {
			Langs[lang] = messages
		} else {
			for key, value := range messages {
				if Langs[lang][key] != "" {
					continue
				}
				Langs[lang][key] = value
			}
		}
	}
}

func handleOnNew(k *Kuu) {
	config := map[string]LangMessages{}
	if k.Config["i18n"] != nil {
		JSONConvert(k.Config["i18n"], config)
	}
	mergeLangs(config)
}

func handleOnImport(p *Mod) {
	mergeLangs(p.Langs)
}

func init() {
	On("OnNew", func(args ...interface{}) {
		k := args[0].(*Kuu)
		handleOnNew(k)
	})
	On("OnImport", func(args ...interface{}) {
		p := args[0].(*Mod)
		handleOnImport(p)
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

func renderMessage(lang string, key string, defaultMessage string, context H) string {
	if key == "" || Langs[lang] == nil || Langs[lang][key] == "" {
		if defaultMessage != "" {
			return defaultMessage
		}
		return key
	}
	data := Langs[lang][key]
	return mustache.Render(data, context)
}

// L 国际化函数
func L(langOrContext interface{}, key string, defaultMsgAndArgs ...interface{}) string {
	var lang string
	if v, ok := langOrContext.(*gin.Context); ok {
		lang = parseAcceptLanguage(v)
	} else if v, ok := langOrContext.(string); ok {
		lang = v
	}
	var (
		defaultMessage string
		context        H
	)
	if len(defaultMsgAndArgs) > 0 {
		if v, ok := defaultMsgAndArgs[0].(string); ok {
			defaultMessage = v
		} else if v, ok := defaultMsgAndArgs[0].(H); ok {
			context = v
		} else {
			defaultMessage = key
		}
	}
	if len(defaultMsgAndArgs) > 1 {
		if v, ok := defaultMsgAndArgs[1].(H); ok {
			context = v
		}
	}
	if lang == "" {
		if defaultMessage != "" {
			return defaultMessage
		} else {
			return key
		}
	}
	return renderMessage(lang, key, defaultMessage, context)
}
