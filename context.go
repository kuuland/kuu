package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/kuuland/kuu/intl"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

type STDReply struct {
	HTTPAction func(code int, jsonObj interface{}) `json:"-"`
	HTTPCode   int                                 `json:"-"`
	Code       int                                 `json:"code"`
	Data       interface{}                         `json:"data,omitempty"`
	Message    string                              `json:"msg,omitempty"`
}

func (s *STDReply) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{"code": s.Code}
	if !IsNil(s.Data) && s.Code == 0 {
		data["data"] = s.Data
	}
	if s.Message != "" {
		data["msg"] = s.Message
	}

	return JSON().Marshal(data)
}

// Context
type Context struct {
	*gin.Context
	app           *Engine
	SignInfo      *SignContext
	PrisDesc      *PrivilegesDesc
	RoutineCaches RoutineCaches
}

func (c *Context) RequestID() string {
	cacheKey := "__request_id__"

	var idVal string
	if v, has := c.Get(cacheKey); has {
		idVal = v.(string)
	} else {
		idVal = uuid.NewV4().String()
		c.Set(cacheKey, idVal)
	}

	return idVal
}

// STD render a JSON body with code(default is 0), data and message.
//
// c.STD(data, localeMessageName, defaultLocaleMessageText, localeMessageValues)
//
// Examples:
//
// c.STD(data)
// c.STD(data, "hello")
// c.STD(data, "hello", "Hello")
// c.STD(data, "welcome", "Welcome {{name}}", mot.D{"name": "xxx"})
func (c *Context) STD(data interface{}, args ...interface{}) *STDReply {
	std := c.std(data, args)
	std.HTTPAction = c.JSON
	return std
}

func (c *Context) STDOK() *STDReply {
	return c.STD("ok")
}

// c.Abort(data, localeMessageName, defaultLocaleMessageText, localeMessageValues)
func (c *Context) Abort(data interface{}, args ...interface{}) *STDReply {
	std := c.std(data, args)
	std.HTTPAction = c.AbortWithStatusJSON
	return std
}

// STDErr render a JSON body with error message, error code(default is -1) and error detail.
//
// c.STDErr(err, localeMessageName, defaultLocaleMessageText, localeMessageValues)
//
// Examples:
//
// c.STDErr(err, "hello")
// c.STDErr(err, "hello", "Hello")
// c.STDErr(err, "welcome", "Welcome {{name}}", D{"name": "xxx"})
func (c *Context) STDErr(err interface{}, args ...interface{}) *STDReply {
	return c.STDErrWithCode(err, -1, args...)
}

// c.AbortErr(err, localeMessageName, defaultLocaleMessageText, localeMessageValues)
func (c *Context) AbortErr(err error, args ...interface{}) *STDReply {
	return c.AbortErrWithCode(err, -1, args)
}

// STDErrWithCode render a JSON body with error message, custom error code and error detail.
//
// c.STDErrWithCode(errData, code, localeMessageName, defaultLocaleMessageText, localeMessageValues)
//
// Examples:
//
// c.STDErrWithCode(errData, 555, "hello")
// c.STDErrWithCode(errData, 501, "hello", "Hello")
// c.STDErrWithCode(errData, 500, "welcome", "Welcome {{name}}", D{"name": "xxx"})
func (c *Context) STDErrWithCode(errData interface{}, code int, args ...interface{}) *STDReply {
	std := c.stdErr(errData, code, args)
	std.HTTPAction = c.JSON
	return std
}

// c.AbortErrWithCode(errData, code, localeMessageName, defaultLocaleMessageText, localeMessageValues)
func (c *Context) AbortErrWithCode(errData error, code int, args ...interface{}) *STDReply {
	std := c.stdErr(errData, code, args)
	std.HTTPAction = c.AbortWithStatusJSON
	return std
}

func (c *Context) std(data interface{}, args []interface{}) *STDReply {
	message := c.resolveLocaleMessage(args, c.Lang())
	reply := STDReply{
		Data:    data,
		Code:    0,
		Message: message,
	}
	return &reply
}

func (c *Context) stdErr(data interface{}, code int, args []interface{}) *STDReply {
	if v, ok := data.(*IntlError); ok {
		args = []interface{}{v.ID}
		if v.DefaultText != "" {
			args = append(args, v.DefaultText)
		}
		if v.ContextValues != nil {
			args = append(args, v.ContextValues)
		}
	}
	if v, ok := data.(error); ok {
		c.ERROR(v)
	}
	message := c.resolveLocaleMessage(args, c.Lang())
	if code == 0 {
		code = -1
	}
	body := STDReply{
		Data:    data,
		Code:    code,
		Message: message,
	}
	return &body
}

func (c *Context) GetIntlMessages() map[string]string {
	key := "__kuu_intl_messages__"
	var messages map[string]string
	if v, has := c.Get(key); has {
		messages = v.(map[string]string)
	} else {
		messages = GetIntlMessagesByLang(c.Lang())
	}
	return messages
}

func (c *Context) resolveLocaleMessage(args []interface{}, lang string) string {
	var (
		messageName    string
		messageContent string
		messageValues  interface{}
	)
	if len(args) > 0 {
		messageName = args[0].(string)
	}
	if len(args) > 1 {
		messageContent = args[1].(string)
	}
	if len(args) > 2 {
		messageValues = args[2]
	}
	if lang == "" {
		lang = c.Lang()
	}
	messages := c.GetIntlMessages()
	result := intl.FormatMessage(messages, messageName, messageContent, messageValues)
	return result
}

// DB
func (c *Context) DB() *gorm.DB {
	return DB()
}

// GetPagination
func (c *Context) GetPagination(ignoreDefault ...bool) (int, int) {
	return GetPagination(c, ignoreDefault...)
}

// ParseCond
func (c *Context) ParseCond(cond interface{}, model interface{}) CondDesc {
	desc, _ := ParseCond(cond, model)
	return desc
}

// Pagination
func GetPagination(ginContextOrKuuContext interface{}, ignoreDefault ...bool) (page int, size int) {
	var c *gin.Context
	if v, ok := ginContextOrKuuContext.(*gin.Context); ok {
		c = v
	} else if v, ok := ginContextOrKuuContext.(*Context); ok {
		c = v.Context
	} else {
		return
	}

	var rawPage, rawSize string
	if len(ignoreDefault) > 0 && ignoreDefault[0] {
		rawPage = c.Query("page")
		rawSize = c.Query("size")
	} else {
		rawPage = c.DefaultQuery("page", "1")
		rawSize = c.DefaultQuery("size", "30")
	}

	if v, err := strconv.Atoi(rawPage); err == nil {
		page = v
	}
	if v, err := strconv.Atoi(rawSize); err == nil {
		size = v
	}
	return
}

func (c *Context) validSignType(sign *SignContext) bool {
	k := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
	info := routesMap[k]

	if len(info.SignType) == 0 {
		return true
	}

	for _, t := range info.SignType {
		if t == sign.Type {
			return true
		}
	}

	return false
}

// QueryCI
func (c *Context) QueryCI(key string) (v string) {
	query := c.Request.URL.Query()
	reg := regexp.MustCompile(fmt.Sprintf("(?i)%s", key))
	for key, values := range query {
		if reg.MatchString(key) {
			v = values[0]
			break
		}
	}
	return v
}

// QueryCI means to get case-insensitive query value
func QueryCI(c *gin.Context, key string) (v string) {
	query := c.Request.URL.Query()
	reg := regexp.MustCompile(fmt.Sprintf("(?i)%s", key))
	for key, values := range query {
		if reg.MatchString(key) {
			v = values[0]
			break
		}
	}
	return
}

func (c *Context) Lang() (lang string) {
	cacheKey := "__kuu_lang__"
	if v, has := c.Get(cacheKey); has {
		return v.(string)
	}

	keys := []string{"Lang", "lang", "l"}
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
		lang = c.parseAcceptLanguage()
	}

	if lang == "" {
		lang = "en"
	}

	lang = intl.ConvertLanguageCode(lang)
	c.Set(cacheKey, lang)
	return
}

func (c *Context) parseAcceptLanguage() string {
	header := c.GetHeader("Accept-Language")
	split := strings.Split(header, ",")
	// zh-CN,zh;q=0.9,zh-TW;q=0.8,en;q=0.7
	var s string
	for _, item := range split {
		item = strings.TrimSpace(item)
		s = strings.TrimSpace(strings.Split(item, ";")[0])
		if s != "" {
			break
		}
	}
	return s
}

// WithTransaction
func (c *Context) WithTransaction(fn func(*gorm.DB) error) error {
	return WithTransaction(fn)
}

// SetValue
func (c *Context) SetRoutineCache(key string, value interface{}) {
	SetRoutineCache(key, value)
}

// DelValue
func (c *Context) DelRoutineCache(key string) {
	DelRoutineCache(key)
}

// GetValue
func (c *Context) GetRoutineCache(key string) interface{} {
	return GetRoutineCache(key)
}

// IgnoreAuth
func (c *Context) IgnoreAuth(cancel ...bool) *Context {
	c.RoutineCaches.IgnoreAuth(cancel...)
	return c
}

// Scheme
func (c *Context) Scheme() string {
	// Can't use `r.Request.URL.Scheme`
	// See: https://groups.google.com/forum/#!topic/golang-nuts/pMUkBlQBDF0
	if c.Request.TLS != nil {
		return "https"
	}
	if scheme := c.Request.Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if scheme := c.Request.Header.Get("X-Forwarded-Protocol"); scheme != "" {
		return scheme
	}
	if ssl := c.Request.Header.Get("X-Forwarded-Ssl"); ssl == "on" {
		return "https"
	}
	if scheme := c.Request.Header.Get("X-Url-Scheme"); scheme != "" {
		return scheme
	}
	return "http"
}

// Origin
func (c *Context) Origin() string {
	return fmt.Sprintf("%s://%s", c.Scheme(), c.Request.Host)
}

// PRINT
func (c *Context) PRINT(v ...interface{}) *Context {
	PRINTWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// DEBUG
func (c *Context) DEBUG(v ...interface{}) *Context {
	DEBUGWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// WARN
func (c *Context) WARN(v ...interface{}) *Context {
	WARNWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// INFO
func (c *Context) INFO(v ...interface{}) *Context {
	INFOWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// ERROR
func (c *Context) ERROR(v ...interface{}) *Context {
	ERRORWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// PANIC
func (c *Context) PANIC(v ...interface{}) *Context {
	PANICWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}

// FATAL
func (c *Context) FATAL(v ...interface{}) *Context {
	FATALWithFields(logrus.Fields{"request_id": c.RequestID()}, v...)
	return c
}
