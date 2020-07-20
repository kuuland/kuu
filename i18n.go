package kuu

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"github.com/jinzhu/gorm"
	"strings"
	"sync"
	"time"
)

var (
	// LanguageMessagesCache
	languageMessagesCache sync.Map
	// RequestLangKey
	RequestLangKey = "Lang"
)

// 需求点：
// 1.缓存LanguageMessage到内存中，每次修改后更新缓存
// 2.保存用户的上一次语言设置，根据请求中的Lang参数自动切换语言

type LanguageMessagesMap map[string]LanguageMessage

// Language
type Language struct {
	gorm.Model `rest:"*" displayName:"国际化语言列表"`
	LangCode   string `name:"语言编码"`
	LangName   string `name:"语言名称"`
}

// LanguageMessage
type LanguageMessage struct {
	gorm.Model       `rest:"*" displayName:"国际化语言条目"`
	c                *gin.Context `json:"-" gorm:"-"`
	lang             string       `json:"-" gorm:"-"`
	LangCode         string       `name:"语言编码"`
	Key              string       `name:"消息键"`
	Value            string       `name:"翻译值"`
	DefaultMessage   string       `name:"默认消息" json:"-,omitempty"`
	FormattedContext interface{}  `name:"格式化上下文" json:"-,omitempty" gorm:"-"`
	Group            string       `name:"分组"`
	Sort             int          `name:"排序值"`
}

// C
func (m *LanguageMessage) C(c *gin.Context) *LanguageMessage {
	m.c = c
	return m
}

// Lang
func (m *LanguageMessage) Lang(lang string) *LanguageMessage {
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
		LangCode: "en-US",
		Key:      r.Key,
		Value:    enUS,
	})
	r.list = append(r.list, &LanguageMessage{
		LangCode: "zh-CN",
		Key:      r.Key,
		Value:    zhCN,
	})
	r.list = append(r.list, &LanguageMessage{
		LangCode: "zh-TW",
		Key:      r.Key,
		Value:    zhTW,
	})
	return r
}

// Append
func (r *LangRegister) Append(msgs ...*LanguageMessage) *LangRegister {
	r.list = append(r.list, msgs...)
	return r
}

// Reset
func (r *LangRegister) Reset() {
	r.list = r.list[0:0]
}

// Exec
func (r *LangRegister) Exec(createOnly ...bool) error {
	if r.DB == nil {
		return nil
	}
	if len(r.list) == 0 {
		return errors.New("empty list")
	}

	// 查询已有数据
	var (
		messageList []LanguageMessage
		messageMap  = make(map[string]LanguageMessage)
	)
	DB().Model(&LanguageMessage{}).Find(&messageList)
	for _, item := range messageList {
		messageMap[fmt.Sprintf("%s_%s", item.LangCode, item.Key)] = item
	}
	quotedTableName := DB().NewScope(&LanguageMessage{}).QuotedTableName()

	// 执行SQL
	var (
		insertBase = fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s, %s) VALUES ",
			r.DB.Dialect().Quote(quotedTableName),
			r.DB.Dialect().Quote("created_at"),
			r.DB.Dialect().Quote("updated_at"),
			r.DB.Dialect().Quote("lang_code"),
			r.DB.Dialect().Quote("key"),
			r.DB.Dialect().Quote("value"),
		)
		insertBuffer bytes.Buffer
		insertVars   []interface{}
		now          = time.Now().Format("2006-01-02 15:04:05")
		batchSize    = 200
	)
	// 执行新增/更新
	for index, item := range r.list {
		var existing LanguageMessage
		if v, ok := messageMap[fmt.Sprintf("%s_%s", item.LangCode, item.Key)]; ok {
			existing = v
		}
		if existing.ID == 0 && item.ID == 0 {
			if insertBuffer.Len() == 0 {
				insertBuffer.WriteString(insertBase)
			}
			insertBuffer.WriteString("(?, ?, ?, ?, ?)")
			insertVars = append(insertVars, now, now, item.LangCode, item.Key, item.Value)
			if (index+1)%batchSize == 0 || index == len(r.list)-1 {
				if sql := insertBuffer.String(); sql != "" {
					if err := r.DB.Exec(sql, insertVars...).Error; err != nil {
						r.Reset()
						return err
					}
					insertBuffer.Reset()
					insertVars = insertVars[0:0]
				}
			} else {
				insertBuffer.WriteString(", ")
			}
		} else {
			if len(createOnly) > 0 && createOnly[0] {
				continue
			}
			id := item.ID
			if id == 0 {
				id = existing.ID
			}
			if id == 0 {
				continue
			}
			sql := fmt.Sprintf("UPDATE %s SET %s = ?, %s = ? WHERE %s = ?",
				DB().Dialect().Quote(quotedTableName),
				DB().Dialect().Quote("updated_at"),
				DB().Dialect().Quote("value"),
				DB().Dialect().Quote("id"),
			)
			if err := r.DB.Exec(sql, now, item.Value, id).Error; err != nil {
				r.Reset()
				return err
			}
		}
	}
	r.Reset()
	return nil
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
	//languageMessagesCache = make(map[string]LanguageMessagesMap)
	for _, item := range list {
		var mm LanguageMessagesMap
		if v, ok := languageMessagesCache.Load(item.LangCode); ok {
			mm = v.(LanguageMessagesMap)
		} else {
			mm = make(LanguageMessagesMap)
		}
		mm[item.Key] = item
		languageMessagesCache.Store(item.LangCode, mm)
	}
}

// GetUserLanguageMessages
func GetUserLanguageMessages(c *gin.Context, userLang ...string) LanguageMessagesMap {
	var count int
	languageMessagesCache.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	if count == 0 {
		RefreshLanguageMessagesCache()
	}
	var (
		messages LanguageMessagesMap
		lang     string
	)
	if c == nil {
		if ctx := GetRoutineRequestContext(); ctx != nil {
			c = ctx.Context
		}
	}
	if len(userLang) > 0 && userLang[0] != "" {
		lang = userLang[0]
	} else if c != nil {
		if sign := GetSignContext(c); sign.IsValid() && sign.Lang != "" {
			lang = sign.Lang
		} else {
			lang = ParseLang(c)
		}
	}
	if v, ok := languageMessagesCache.Load(lang); ok {
		messages = v.(LanguageMessagesMap)
	}
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
