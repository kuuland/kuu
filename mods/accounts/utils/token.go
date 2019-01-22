package utils

import (
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/accounts/models"
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

// ParseUserID 从请求中解析UserID
func ParseUserID(c *gin.Context) string {
	// querystring > header > cookie
	var userID string
	userID = c.Query(UserIDKey)
	if userID == "" {
		userID = c.GetHeader(UserIDKey)
	}
	if userID == "" {
		userID, _ = c.Cookie(UserIDKey)
	}
	return userID
}

// DecodedContext 从请求中获取解码数据
func DecodedContext(c *gin.Context) (jwt.MapClaims, *models.UserSecret) {
	token := ParseToken(c)
	userID := ParseUserID(c)
	if token == "" {
		kuu.Error("No token found in the request!")
		return nil, nil
	}
	UserSecret := kuu.Model("UserSecret")
	var secret = &models.UserSecret{}
	UserSecret.One(kuu.H{
		"Cond": kuu.H{"UserID": userID},
		"Sort": "-UpdatedAt,-CreatedAt",
	}, secret)
	if secret == nil || secret.Secret == "" {
		kuu.Error("Secret not found based on token '%s'!", token)
		return nil, nil
	}
	claims := Decoded(token, secret.Secret)
	return claims, secret
}

// Encoded 加密
func Encoded(claims jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
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
	}
	kuu.Error(err)
	return nil
}
