package kuu

import (
	"github.com/gin-gonic/gin"
	"time"
)

// LogMiddleware
func LogMiddleware(c *gin.Context) {
	// Start timer
	start := time.Now()

	// Process request
	c.Next()

	// 保存登录日志
	log := NewLog(LogTypeAPI, c)
	log.RequestCost = time.Now().Sub(start)
	log.ResponseStatusCode = c.Writer.Status()
	log.RequestErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
	log.ResponseBodySize = c.Writer.Size()

	// 认证信息
	if _, isLogin := c.Get("is_login"); isLogin {
		log.Type = LogTypeSign
		log.SignMethod = SignMethodLogin
	}
	if _, isLogout := c.Get("is_logout"); isLogout {
		log.Type = LogTypeSign
		log.SignMethod = SignMethodLogout
	}

	if v, exists := c.Get(SignContextKey); exists {
		// 用户信息
		signContext := v.(*SignContext)
		user := GetUserFromCache(signContext.UID)
		log.UID = signContext.UID
		log.SubDocID = signContext.SubDocID
		log.Token = signContext.Token
		log.SignType = signContext.Type
		log.SignPayload = JSONStringify(signContext.Payload)
		log.Username = user.Username
		log.RealName = user.Name
	}

	log.Save2Cache()
}
