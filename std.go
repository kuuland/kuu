// Kuu response format:
// (JSON)
// data: 表正常或异常情况下的数据信息，类型由具体接口决定，Required
// msg: 表正常或异常情况下的提示信息，字符串，非必填，Optional
// code: 调用是否成功的明确标记，0表成功，非0表失败（异常时默认值为-1），整型，Optional

package kuu

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

var (
	JSONAction  = "JSON"
	AbortAction = "ABORT"
)

// STDRender
type STDRender struct {
	c               *gin.Context
	httpCode        int
	action          string
	data            interface{}
	code            int32
	message         string
	languageMessage *LanguageMessage
}

func std(renderNow bool, c *gin.Context, data interface{}, msg ...*LanguageMessage) *STDRender {
	std := &STDRender{c: c}
	std.data = data
	if len(msg) > 0 {
		std.languageMessage = msg[0]
	}
	if renderNow {
		std.Render()
	}
	return std
}

func stdErr(renderNow bool, c *gin.Context, msg *LanguageMessage, err ...interface{}) *STDRender {
	std := &STDRender{c: c, code: -1}
	std.languageMessage = msg
	if len(err) > 0 {
		std.data = err[0]
	}
	if renderNow {
		std.Render()
	}
	return std
}

// STD
func STD(c *gin.Context, data interface{}, msg ...*LanguageMessage) *STDRender {
	return std(true, c, data, msg...)
}

// STDErr
func STDErr(c *gin.Context, msg *LanguageMessage, err ...interface{}) *STDRender {
	return stdErr(true, c, msg, err...)
}

// STDHold
func STDHold(c *gin.Context, data interface{}, msg ...*LanguageMessage) *STDRender {
	return std(false, c, data, msg...)
}

// STDErrHold
func STDErrHold(c *gin.Context, msg *LanguageMessage, err ...interface{}) *STDRender {
	return stdErr(false, c, msg, err...)
}

// Context
func (r *STDRender) Context(c *gin.Context) *STDRender {
	r.c = c
	return r
}

// GetContext
func (r *STDRender) GetContext() *gin.Context {
	return r.c
}

// Data
func (r *STDRender) Data(data interface{}) *STDRender {
	r.data = data
	return r
}

// Code
func (r *STDRender) Code(code int32) *STDRender {
	r.code = code
	return r
}

// Message
func (r *STDRender) Message(message *LanguageMessage) *STDRender {
	r.languageMessage = message
	return r
}

// Action
func (r *STDRender) Action(action string) *STDRender {
	r.action = action
	return r
}

// JSON
func (r *STDRender) JSON() {
	r.action = JSONAction
	r.Render()
}

// Abort
func (r *STDRender) Abort() {
	r.action = AbortAction
	r.Render()
}

// HTTPCode
func (r *STDRender) HTTPCode(httpCode int) *STDRender {
	r.httpCode = httpCode
	return r
}

// Render
func (r *STDRender) Render() {
	if r.c == nil {
		return
	}
	if r.action == "" {
		r.action = "JSON"
	}
	if r.httpCode == 0 {
		r.httpCode = http.StatusOK
	}
	if r.code != 0 {
		if v, ok := r.data.(error); ok {
			ERROR(v)
			r.data = v.Error()
		} else if v, ok := r.data.([]error); ok {
			ERROR(v)
			r.data = v[0].Error()
		}
	}
	r.action = strings.TrimSpace(strings.ToUpper(r.action))
	if r.languageMessage != nil && r.message == "" {
		r.message = r.languageMessage.Render()
	}
	ret := make(map[string]interface{})
	ret["code"] = r.code
	if !IsBlank(r.data) {
		ret["data"] = r.data
	}
	if r.message != "" {
		ret["msg"] = r.message
	}
	switch r.action {
	case JSONAction:
		r.c.JSON(r.httpCode, ret)
	case AbortAction:
		r.c.AbortWithStatusJSON(r.httpCode, ret)
	default:
		PANIC(`invalid action, try "c.%s"`, r.action)
	}
}
