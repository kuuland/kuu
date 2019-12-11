package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"strconv"
)

// Context
type Context struct {
	*gin.Context
	SignInfo      *SignContext
	PrisDesc      *PrivilegesDesc
	RoutineCaches RoutineCaches
}

// L
func (c *Context) L(key string, defaultMessage string, formattedContext ...interface{}) *LanguageMessage {
	msg := L(key, defaultMessage, formattedContext...)
	msg.c = c.Context
	return msg
}

// DB
func (c *Context) DB() *gorm.DB {
	return DB()
}

// GetPagination
func (c *Context) GetPagination() (int, int) {
	return GetPagination(c)
}

// Pagination
func GetPagination(ginContextOrKuuContext interface{}) (page int, size int) {
	var c *gin.Context
	if v, ok := ginContextOrKuuContext.(*gin.Context); ok {
		c = v
	} else if v, ok := ginContextOrKuuContext.(*Context); ok {
		c = v.Context
	} else {
		return
	}
	rawPage := c.DefaultQuery("page", "1")
	rawSize := c.DefaultQuery("size", "30")
	if v, err := strconv.Atoi(rawPage); err == nil {
		page = v
	}
	if v, err := strconv.Atoi(rawSize); err == nil {
		size = v
	}
	return
}

// WithTransaction
func (c *Context) WithTransaction(fn func(*gorm.DB) error) error {
	return WithTransaction(fn)
}

// STD
func (c *Context) STD(data interface{}, msg ...*LanguageMessage) *STDRender {
	return STD(c.Context, data, msg...)
}

// STDErr
func (c *Context) STDErr(msg *LanguageMessage, err ...interface{}) *STDRender {
	return STDErr(c.Context, msg, err...)
}

// STDHold
func (c *Context) STDHold(data interface{}, msg ...*LanguageMessage) *STDRender {
	return STDHold(c.Context, data, msg...)
}

// STDErrHold
func (c *Context) STDErrHold(msg *LanguageMessage, err ...interface{}) *STDRender {
	return STDErrHold(c.Context, msg, err...)
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

// PRINT
func (c *Context) PRINT(args ...interface{}) *Context {
	PRINT(args...)
	return c
}

// DEBUG
func (c *Context) DEBUG(args ...interface{}) *Context {
	DEBUG(args...)
	return c
}

// WARN
func (c *Context) WARN(args ...interface{}) *Context {
	WARN(args...)
	return c
}

// INFO
func (c *Context) INFO(args ...interface{}) *Context {
	INFO(args...)
	return c
}

// ERROR
func (c *Context) ERROR(args ...interface{}) *Context {
	ERROR(args...)
	return c
}

// FATAL
func (c *Context) FATAL(args ...interface{}) *Context {
	FATAL(args...)
	return c
}

// PANIC
func (c *Context) PANIC(args ...interface{}) *Context {
	PANIC(args...)
	return c
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
