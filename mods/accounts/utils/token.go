package utils

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/models"
)

var (
	// ExpiresSeconds 过期秒数
	ExpiresSeconds = 7200
)

// ParseToken 从请求中解析token
func ParseToken(c *gin.Context) string {
	// querystring > header > cookie
	var token string
	token = c.Query(TokenKey)
	if token == "" {
		token = c.GetHeader(TokenKey)
	}
	if token == "" {
		token, _ = c.Cookie(TokenKey)
	}
	return token
}

// DecodedContext 从请求中获取解码数据
func DecodedContext(c *gin.Context) jwt.MapClaims {
	token := ParseToken(c)
	if token == "" {
		kuu.Error("No token found in the request!")
		return nil
	}
	UserSecret := kuu.Model("UserSecret")
	var secretData = &models.UserSecret{}
	UserSecret.One(kuu.H{
		"Cond": kuu.H{"Token": token},
		"Sort": "-UpdatedAt,-CreatedAt",
	}, secretData)
	if secretData == nil || secretData.Secret == "" {
		kuu.Error("Secret not found based on token '%s'!", token)
		return nil
	}
	return Decoded(token, secretData.Secret)
}

// Encoded 加密
func Encoded(data jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		kuu.Error(err)
	}
	return tokenString
}

// Decoded 解密
func Decoded(tokenString string, secret string) jwt.MapClaims {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims
	} else {
		kuu.Error(err)
	}
	return nil
}
