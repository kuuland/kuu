package utils

import (
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

var (
	// TokenKey 令牌键
	TokenKey = "Token"
	// UserIDKey 用户ID键
	UserIDKey = "UID"
	// WhiteList 白名单
	WhiteList = []string{}
	// ExpiresSeconds 令牌过期秒数
	ExpiresSeconds = 86400
	// LoginHandler 登录接口
	LoginHandler func(*gin.Context) (kuu.H, error)
	// ContextClaimsKey 请求上下文缓存中的Claims数据键
	ContextClaimsKey = "UserClaims"
	// ContextSecretKey 请求上下文缓存中的Secret数据键
	ContextSecretKey = "UserSecret"
)
