package accounts

import (
	"fmt"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

// Encoded 加密
func Encoded(data jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		log.Println(err)
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
		fmt.Println(err)
	}
	return nil
}

// tokenFromRequest 从请求信息中获取令牌
func tokenFromRequest(c *gin.Context) string {
	// querystring > body > header > cookie
	tokenKey := "token"
	var token string
	token = c.Query(tokenKey)
	if token == "" {
		var docs map[string]interface{}
		c.ShouldBindJSON(&docs)
		if docs != nil && docs[tokenKey] != nil {
			token = docs[tokenKey].(string)
		}
	}
	if token == "" {
		token = c.PostForm(tokenKey)
	}
	if token == "" {
		token = c.GetHeader(tokenKey)
	}
	if token == "" {
		token, _ = c.Cookie(tokenKey)
	}
	return token
}

// AuthMiddleware 鉴权中间件
func AuthMiddleware(c *gin.Context) {
	token := tokenFromRequest(c)
	log.Println(token)
}

// Install 导出插件
func Install() *kuu.Plugin {
	return &kuu.Plugin{
		Name: "ac",
		Middleware: kuu.M{
			"AuthMiddleware": AuthMiddleware,
		},
		Methods: kuu.Methods{
			"Encoded": func(args ...interface{}) interface{} {
				if args != nil && len(args) == 2 {
					data := args[0].(jwt.MapClaims)
					secret := args[1].(string)
					return Encoded(data, secret)
				}
				return nil
			},
			"Decoded": func(args ...interface{}) interface{} {
				if args != nil && len(args) == 2 {
					tokenString := args[0].(string)
					secret := args[1].(string)
					return Decoded(tokenString, secret)
				}
				return nil
			},
		},
	}
}
