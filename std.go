// Kuu response format:
// (JSON)
// data: 表正常或异常情况下的数据信息，类型由具体接口决定，Required
// msg: 表正常或异常情况下的提示信息，字符串，非必填，Optional
// code: 调用是否成功的明确标记，0表成功，非0表失败（异常时默认值为-1），整型，Optional

package kuu

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type STDRender struct {
	Context  *gin.Context `json:"-"`
	HTTPCode int          `json:"-"`
	Action   string       `json:"-"`
	Data     interface{}  `json:"data"`
	Code     int32        `json:"code"`
	Message  string       `json:"msg"`
}

// STD
func STD(c *gin.Context, data interface{}, msg ...string) {
	std := &STDRender{HTTPCode: http.StatusOK, Context: c, Action: "JSON"}
	std.Data = data
	if len(msg) > 0 {
		std.Message = msg[0]
	}
	std.Render()
}

// STDErr
func STDErr(c *gin.Context, msg string, code ...int32) {
	std := &STDRender{HTTPCode: http.StatusOK, Context: c, Action: "JSON"}
	std.Message = msg
	if len(code) > 0 {
		std.Code = code[0]
	} else {
		std.Code = -1
	}
	std.Render()
}

func (r *STDRender) Render() {
	r.Action = strings.TrimSpace(strings.ToUpper(r.Action))
	switch r.Action {
	case "JSON":
		r.Context.JSON(r.HTTPCode, r)
	case "ABORT":
		r.Context.AbortWithStatusJSON(r.HTTPCode, r)
	default:
		fmt.Errorf(`invalid action: %s, please use "c.%s" instead`, r.Action, r.Action)
	}
}
