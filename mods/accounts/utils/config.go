package utils

import (
	"github.com/gin-gonic/gin"
)

var (
	// TokenKey 令牌键
	TokenKey = "token"
	// UserIDKey 用户ID键
	UserIDKey = "_id"
	// WhiteList 白名单
	WhiteList = []string{}
	// LoginHandler 登录接口
	LoginHandler func(*gin.Context) (interface{}, error)
)
