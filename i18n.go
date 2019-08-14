package kuu

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
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
	c                *gin.Context `json:"-" gorm:"-"`
	lang             string       `json:"-" gorm:"-"`
	ModelExOrg       `rest:"*" displayName:"国际化语言条目"`
	LangCode         string      `name:"语言编码"`
	Key              string      `name:"消息键"`
	Value            string      `name:"翻译值"`
	DefaultMessage   string      `name:"默认消息" json:"-,omitempty"`
	FormattedContext interface{} `name:"格式化上下文" json:"-,omitempty" gorm:"-"`
	Group            string      `name:"分组"`
	Sort             int         `name:"排序值"`
	IsBuiltIn        null.Bool   `name:"是否预置"`
}

// SetContext
func (m *LanguageMessage) SetContext(c *gin.Context) *LanguageMessage {
	m.c = c
	return m
}

// SetLang
func (m *LanguageMessage) SetLang(lang string) *LanguageMessage {
	m.lang = lang
	return m
}

// Render
func (m *LanguageMessage) Render() string {
	messages := GetUserLanguageMessages(m.c, m.lang)
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
	DB   *gorm.DB
	Key  string
	list []*LanguageMessage
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
	r.list = append(r.list, &LanguageMessage{
		LangCode:  "en-US",
		Key:       r.Key,
		Value:     enUS,
		IsBuiltIn: null.NewBool(true, true),
	})
	r.list = append(r.list, &LanguageMessage{
		LangCode:  "zh-CN",
		Key:       r.Key,
		Value:     zhCN,
		IsBuiltIn: null.NewBool(true, true),
	})
	r.list = append(r.list, &LanguageMessage{
		LangCode:  "zh-TW",
		Key:       r.Key,
		Value:     zhTW,
		IsBuiltIn: null.NewBool(true, true),
	})
	return r
}

// Exec
func (r *LangRegister) Exec(db ...*gorm.DB) *LangRegister {
	if len(db) > 0 {
		r.DB = db[0]
	}
	if len(r.list) == 0 {
		return r
	}

	var (
		base      = `INSERT INTO "sys_LanguageMessage" (created_at, updated_at, created_by_id, updated_by_id, lang_code, key, value, is_built_in) VALUES `
		now       = time.Now().Format("2006-01-02 15:04:05")
		batchSize = 200
		buffer    bytes.Buffer
		vars      []interface{}
	)
	buffer.WriteString(base)
	for index, item := range r.list {
		buffer.WriteString("(?, ?, ?, ?, ?, ?, ?, TRUE)")
		vars = append(vars, now, now, RootUID(), RootUID(), item.LangCode, item.Key, item.Value)
		if (index+1)%batchSize == 0 || index == len(r.list)-1 {
			if sql := buffer.String(); sql != "" {
				r.DB.Exec(sql, vars...)
				buffer.Reset()
				buffer.WriteString(base)
				vars = vars[0:0]
			}
		} else {
			buffer.WriteString(", ")
		}
	}
	r.list = r.list[0:0]
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
	languageMessagesCache = make(map[string]LanguageMessagesMap)
	for _, item := range list {
		if languageMessagesCache[item.LangCode] == nil {
			languageMessagesCache[item.LangCode] = make(LanguageMessagesMap)
		}
		languageMessagesCache[item.LangCode][item.Key] = item
	}
}

// GetUserLanguageMessages
func GetUserLanguageMessages(c *gin.Context, userLang ...string) LanguageMessagesMap {
	if len(languageMessagesCache) == 0 {
		RefreshLanguageMessagesCache()
	}
	var (
		messages LanguageMessagesMap
		lang     string
	)
	if len(userLang) > 0 && userLang[0] != "" {
		lang = userLang[0]
	} else if sign := GetSignContext(c); sign.IsValid() && sign.Lang != "" {
		lang = sign.Lang
	} else {
		lang = ParseLang(c)
	}
	messages = languageMessagesCache[lang]
	return messages
}

// ParseLang
var ParseLang = func(langOrContext interface{}) string {
	if v, ok := langOrContext.(string); ok {
		return v
	}

	var (
		lang string
		c    *gin.Context
	)
	if v, ok := langOrContext.(*gin.Context); ok && v != nil {
		c = v
	} else if v, ok := langOrContext.(*Context); ok && v != nil {
		c = v.Context
	}

	if c != nil {
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
