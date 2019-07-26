package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"github.com/jinzhu/gorm"
	"strings"
)

var (
	// LanguageMessagesCache
	languageMessagesCache = map[string]LanguageMessagesMap{}
	// RequestLangKey
	RequestLangKey = "Lang"
)

// 需求点：
// 1.缓存LanguageMessage到内存中，每次修改后更新缓存
// 2.保存用户的上一次语言设置，根据请求中的Lang参数自动切换语言

type LanguageMessagesMap map[string]LanguageMessage

// Language
type Language struct {
	ModelExOrg `rest:"*" displayName:"国际化语言列表"`
	LangCode   string `name:"语言编码"`
	LangName   string `name:"语言名称"`
}

// LanguageMessage
type LanguageMessage struct {
	ModelExOrg       `rest:"*" displayName:"国际化语言条目"`
	LangCode         string      `name:"语言编码"`
	Key              string      `name:"消息键"`
	Value            string      `name:"翻译值"`
	DefaultMessage   string      `name:"默认消息" json:"-,omitempty"`
	FormattedContext interface{} `name:"格式化上下文" json:"-,omitempty" gorm:"-"`
	Group            string      `name:"分组"`
	Sort             int         `name:"排序值"`
}

// Render
func (m *LanguageMessage) Render() string {
	messages := GetUserLanguageMessages()
	var template string
	if v, has := messages[m.Key]; has {
		template = v.Value
	} else {
		template = m.DefaultMessage
	}
	return mustache.Render(template, m.FormattedContext)
}

// LangRegister
type LangRegister struct {
	DB  *gorm.DB
	Key string
}

// NewLangRegister
func NewLangRegister(db *gorm.DB, key ...string) *LangRegister {
	r := &LangRegister{DB: db}
	if len(key) > 0 {
		r.Key = key[0]
	}
	return r
}

// SetKey
func (r *LangRegister) SetKey(key string) *LangRegister {
	r.Key = key
	return r
}

// SetDB
func (r *LangRegister) SetDB(db *gorm.DB) *LangRegister {
	r.DB = db
	return r
}

// Add
func (r *LangRegister) Add(enUS string, zhCN string, zhTW string) *LangRegister {
	r.DB.Create(&LanguageMessage{
		LangCode: "en-US",
		Key:      r.Key,
		Value:    enUS,
	})
	r.DB.Create(&LanguageMessage{
		LangCode: "zh-CN",
		Key:      r.Key,
		Value:    zhCN,
	})
	r.DB.Create(&LanguageMessage{
		LangCode: "zh-TW",
		Key:      r.Key,
		Value:    zhTW,
	})
	return r
}

// TranslatedList
type TranslatedList []map[string]interface{}

// Len
func (l TranslatedList) Len() int {
	return len(l)
}

// Less
func (l TranslatedList) Less(i, j int) bool {
	return l[i]["Sort"].(int) < l[j]["Sort"].(int)
}

// Swap
func (l TranslatedList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// L
func L(key string, defaultMessage string, formattedContext ...interface{}) *LanguageMessage {
	if key == "" || defaultMessage == "" {
		PANIC("i18n message key and default message are required: %s, %s", key, defaultMessage)
	}
	msg := &LanguageMessage{Key: key, DefaultMessage: defaultMessage}
	if len(formattedContext) > 0 {
		msg.FormattedContext = formattedContext[0]
	}
	return msg
}

// RefreshLanguageMessagesCache
func RefreshLanguageMessagesCache() {
	var list []LanguageMessage
	if err := DB().Find(&list).Error; err != nil {
		ERROR("Refreshing i18n cache failed: %s", err.Error())
		return
	}
	languageMessagesCache = map[string]LanguageMessagesMap{}
	for _, item := range list {
		if languageMessagesCache[item.LangCode] == nil {
			languageMessagesCache[item.LangCode] = make(LanguageMessagesMap)
		}
		languageMessagesCache[item.LangCode][item.Key] = item
	}
}

// GetUserLanguageMessages
func GetUserLanguageMessages() LanguageMessagesMap {
	var messages LanguageMessagesMap
	if desc := GetRoutinePrivilegesDesc(); desc.IsValid() {
		if len(languageMessagesCache) == 0 {
			RefreshLanguageMessagesCache()
		}
		messages = languageMessagesCache[desc.SignInfo.Lang]
	}
	return messages
}

// ParseLang
var ParseLang = func(langOrContext interface{}) string {
	if v, ok := langOrContext.(string); ok {
		return v
	}

	var lang string
	if c, ok := langOrContext.(*gin.Context); ok {
		keys := []string{"Lang", "lang", "klang", "l"}
		// querystring > header > cookie
		for _, key := range keys {
			if val := c.Query(key); val != "" {
				lang = val
				break
			}
		}
		if lang == "" {
			for _, key := range keys {
				if val := c.GetHeader(key); val != "" {
					lang = val
					break
				}
			}
		}
		if lang == "" {
			for _, key := range keys {
				if val, _ := c.Cookie(key); val != "" {
					lang = val
					break
				}
			}
		}
		if lang == "" {
			lang = parseAcceptLanguage(c)
		}
	}
	if lang == "" {
		lang = "en-US"
	}
	return lang
}

func parseAcceptLanguage(c *gin.Context) (lang string) {
	header := c.GetHeader("Accept-Language")
	split := strings.Split(header, ",")
	// zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7
	for _, item := range split {
		item = strings.TrimSpace(item)
		s := strings.TrimSpace(strings.Split(item, ";")[0])
		return s
	}
	return
}
