package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"strings"
)

var (
	// CaptchaIDKey
	CaptchaIDKey = "captcha_id"
	// CaptchaValKey
	CaptchaValKey = "captcha_val"
)

// GenerateCaptcha create a digit captcha.
func GenerateCaptcha(idKey string, configs ...interface{}) (id string, base64Str string) {
	var config interface{}
	if len(configs) > 0 {
		config = configs[0]
	} else {
		config = base64Captcha.ConfigDigit{
			Height:     30,
			Width:      100,
			MaxSkew:    1,
			DotCount:   90,
			CaptchaLen: 4,
		}
	}
	id, captchaInstance := base64Captcha.GenerateCaptcha(idKey, config)
	base64Str = base64Captcha.CaptchaWriteToBase64Encoding(captchaInstance)
	return
}

// VerifyCaptcha Verify captcha value.
func VerifyCaptcha(idKey, value string) bool {
	return base64Captcha.VerifyCaptcha(idKey, value)
}

// ParseCaptchaID
func ParseCaptchaID(c interface{}) string {
	var (
		ctx *gin.Context
		id  string
	)
	if v, ok := c.(*gin.Context); ok {
		ctx = v
	} else if v, ok := c.(*Context); ok {
		ctx = v.Context
	}
	if ctx != nil {
		id = ctx.Query(CaptchaIDKey)
		if id == "" {
			id = ctx.GetHeader(CaptchaIDKey)
		}
		if id == "" {
			id = ctx.GetHeader(strings.ReplaceAll(CaptchaIDKey, "-", ""))
		}
		id, _ = ctx.Cookie(CaptchaIDKey)
	}
	return id
}

// ParseCaptchaValue
func ParseCaptchaValue(c interface{}) string {
	var (
		ctx *gin.Context
		val string
	)
	if v, ok := c.(*gin.Context); ok {
		ctx = v
	} else if v, ok := c.(*Context); ok {
		ctx = v.Context
	}
	if ctx != nil {
		val = ctx.Query(CaptchaValKey)
		if val == "" {
			val = ctx.GetHeader(CaptchaValKey)
		}
		if val == "" {
			val = ctx.GetHeader(strings.ReplaceAll(CaptchaValKey, "-", ""))
		}
	}
	return val
}
