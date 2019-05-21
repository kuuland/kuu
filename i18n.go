package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hoisie/mustache"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// Language
type Language struct {
	gorm.Model `rest`
	Code       string
	Name       string
	Key        string
	Value      string
}

// L
func L(langOrContext interface{}, defaultValue string, args ...interface{}) string {
	if len(args) > 0 {
		return LFull(langOrContext, "", defaultValue, args[0])
	} else {
		return LFull(langOrContext, "", defaultValue, nil)
	}
}

// LFull
func LFull(langOrContext interface{}, key string, defaultValue string, args interface{}) string {
	lang := parseLang(langOrContext)
	if lang == "" {
		lang = "zh"
	}
	if key == "" {
		key = strings.Replace(defaultValue, "{", "", -1)
		key = strings.Replace(defaultValue, "}", "", -1)
		key = strings.Replace(defaultValue, "{", "", -1)
		key = strings.Replace(defaultValue, "}", "", -1)
		key = strings.Replace(defaultValue, " ", "_", -1)
		key = strings.TrimSpace(key)
		key = strings.ToLower(key)
	}
	return renderMessage(lang, key, defaultValue, args)
}

func parseLang(langOrContext interface{}) (lang string) {
	if c, ok := langOrContext.(*gin.Context); ok {
		lang = parseKuuLang(c)
		if lang == "" {
			lang = parseAcceptLanguage(c)
		}
	} else if v, ok := langOrContext.(string); ok {
		lang = v
	}
	return
}

func parseKuuLang(c *gin.Context) (lang string) {
	for _, key := range []string{"KuuLang", "Kuu-Lang", "Kuu_Lang"} {
		if lang = c.GetHeader(key); lang != "" {
			return
		}
	}
	return
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

func renderMessage(code string, name string, defaultValue string, args interface{}) string {
	key := strings.ToLower(fmt.Sprintf("%s_i18n_%s", RedisPrefix, code))
	// 如果该语言没有缓存，则查询数据库并缓存至redis
	if v := RedisClient.Exists(key).Val(); v == 0 {
		list := make([]Language, 0)
		DB().Where(&Language{Code: code}).Find(&list)
		fields := make(map[string]interface{})
		for _, item := range list {
			item.Key = strings.TrimSpace(item.Key)
			item.Value = strings.TrimSpace(item.Value)
			fields[item.Key] = item.Value
		}
		if len(fields) == 0 {
			fields["hello_kuu_i18n"] = "Hello Kuu"
		}
		if _, err := RedisClient.HMSet(key, fields).Result(); err != nil {
			ERROR(err)
		} else {
			if _, err := RedisClient.Expire(key, 24*time.Hour).Result(); err != nil {
				ERROR(err)
			}
		}
	}
	// 如果该语言下不存在指定键，则新增键到数据库并缓存至redis
	var template string
	if exists := RedisClient.HExists(key, name).Val(); exists {
		template = RedisClient.HGet(key, name).Val()
		if template == "" {
			template = defaultValue
		}
	} else {
		doc := Language{Code: code, Key: name, Value: defaultValue}
		DB().Create(&doc)
		if RedisClient.HSet(key, doc.Key, doc.Value).Val() {
			template = defaultValue
		}
	}
	return mustache.Render(template, args)
}
