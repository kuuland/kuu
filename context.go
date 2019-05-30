package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Context
type Context struct {
	*gin.Context
	SignInfo *SignContext
	PrisDesc *PrivilegesDesc
	Values   *Values
}

// L
func (c *Context) L(defaultValue string, args ...interface{}) string {
	return L(c.Context, defaultValue, args...)
}

// Lang
func (c *Context) Lang(key string, defaultValue string, args interface{}) string {
	return Lang(c.Context, key, defaultValue, args)
}

// DB
func (c *Context) DB() *gorm.DB {
	return DB()
}

// WithTransaction
func (c *Context) WithTransaction(fn func(*gorm.DB) error, with ...*gorm.DB) error {
	return WithTransaction(fn, with...)
}

// STD
func (c *Context) STD(data interface{}, msg ...string) *STDRender {
	return STD(c.Context, data, msg...)
}

// STDErr
func (c *Context) STDErr(msg string, code ...int32) *STDRender {
	return STDErr(c.Context, msg, code...)
}

// STDHold
func (c *Context) STDHold(data interface{}, msg ...string) *STDRender {
	return STDHold(c.Context, data, msg...)
}

// STDErrHold
func (c *Context) STDErrHold(msg string, code ...int32) *STDRender {
	return STDErrHold(c.Context, msg, code...)
}

// SetValue
func (c *Context) SetValue(key string, value interface{}) {
	(*c.Values)[key] = value
}

// DelValue
func (c *Context) DelValue(key string) {
	delete((*c.Values), key)
}

// GetValue
func (c *Context) GetValue(key string) interface{} {
	return (*c.Values)[key]
}

// PRINT
func (c *Context) PRINT(args ...interface{}) {
	PRINT(args...)
}

// DEBUG
func (c *Context) DEBUG(args ...interface{}) {
	DEBUG(args...)
}

// WARN
func (c *Context) WARN(args ...interface{}) {
	WARN(args...)
}

// INFO
func (c *Context) INFO(args ...interface{}) {
	INFO(args...)
}

// ERROR
func (c *Context) ERROR(args ...interface{}) {
	ERROR(args...)
}

// FATAL
func (c *Context) FATAL(args ...interface{}) {
	FATAL(args...)
}

// PANIC
func (c *Context) PANIC(args ...interface{}) {
	PANIC(args...)
}

// IgnoreAuth
func (c *Context) IgnoreAuth(cancel ...bool) {
	if len(cancel) > 0 && cancel[0] == true {
		c.DelValue(IgnoreAuthKey)
	} else {
		c.SetValue(IgnoreAuthKey, true)
	}
}
