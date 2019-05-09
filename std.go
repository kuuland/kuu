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
	HTTPCode int          `json:"-"`
	Action   string       `json:"-"`
	Context  *gin.Context `json:"-"`
	Data     interface{}  `json:"data"`
	Code     int32        `json:"code"`
	Message  string       `json:"msg"`
}

// STD
func STD(c *gin.Context) *STDRender {
	return &STDRender{HTTPCode: http.StatusOK, Context: c, Action: "JSON"}
}

// OK
func (r *STDRender) OK(data interface{}, msg ...string) {
	r.Data = data
	if len(msg) > 0 {
		r.Message = msg[0]
	}
	r.render()
}

// ERROR
func (r *STDRender) ERROR(msg string, code ...int32) {
	r.Message = msg

	if len(code) > 0 {
		r.Code = code[0]
	} else {
		r.Code = -1
	}
	r.render()
}

func (r *STDRender) render() {
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
