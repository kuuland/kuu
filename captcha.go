package kuu

import (
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"math/rand"
	"strings"
	"time"
)

var (
	// CaptchaIDKey
	CaptchaIDKey = "captcha_id"
	// CaptchaValKey
	CaptchaValKey = "captcha_val"
	store         = &captchaStore{}
)

func init() {
	//init rand seed
	rand.Seed(time.Now().UnixNano())
}

type captchaStore struct{}

func (cs *captchaStore) Set(id string, value string) {
	DefaultCache.SetString(id, value)
}

func (cs *captchaStore) Get(id string, clear bool) string {
	v := DefaultCache.GetString(id)
	if clear {
		DefaultCache.Del(id)
	}
	return v
}

func (cs *captchaStore) Verify(id, answer string, clear bool) bool {
	v := cs.Get(id, clear)
	return v == answer
}

//NewCaptcha creates a captcha instance from driver and store
func NewCaptcha() *base64Captcha.Captcha {
	driver := &base64Captcha.DriverDigit{
		Height:   80,
		Width:    280,
		Length:   4,
		MaxSkew:  0.1,
		DotCount: 10,
	}
	return base64Captcha.NewCaptcha(driver, store)
}

// GenerateCaptcha create a digit captcha.
func GenerateCaptcha() (id string, base64Str string) {
	c := NewCaptcha()
	id, base64Str, err := c.Generate()
	if err != nil {
		ERROR(err)
	}
	return
}

// VerifyCaptcha Verify captcha value.
func VerifyCaptcha(idKey, value string) bool {
	return store.Verify(idKey, value, true)
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
