package accounts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

func safeL(c *gin.Context, key string) string {
	defaultMessages := map[string]string{
		"login_error": "Login failed, please contact the administrator or try again later.",
		"auth_error":  "Your session has expired, please log in again.",
		"logout":      "Logout successful.",
	}
	value := kuu.L(c, key)
	if value == "" {
		value = defaultMessages[key]
	}
	return value
}

// 全局配置
var (
	tokenKey   = "token"
	loginFunc  func(*gin.Context) (kuu.H, string)
	secretFunc func(string) string
	filterFunc = func(c *gin.Context) bool {
		return false
	}
	errorHandler = func(c *gin.Context) {
		c.JSON(http.StatusOK, kuu.StdDataError(safeL(c, "auth_error")))
	}
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

func parseToken(c *gin.Context) string {
	// querystring > header > cookie
	var token string
	token = c.Query(tokenKey)
	if token == "" {
		token = c.GetHeader(tokenKey)
	}
	if token == "" {
		token, _ = c.Cookie(tokenKey)
	}
	return token
}

func encodeData(c *gin.Context) jwt.MapClaims {
	token := parseToken(c)
	secret := secretFunc(token)
	data := Decoded(token, secret)
	return data
}

// AuthMiddleware 令牌鉴权中间件
func AuthMiddleware(c *gin.Context) {
	data := encodeData(c)
	if data == nil || data.Valid() != nil || !filterFunc(c) {
		errorHandler(c)
		c.Abort()
	}
}

// login 登录路由
func login(c *gin.Context) {
	userData, secret := loginFunc(c)
	var token string
	if secret == "" {
		if userData["secret"] != nil {
			secret = userData["secret"].(string)
			delete(userData, "secret")
		}
		if userData["password"] != nil {
			secret = userData["password"].(string)
			delete(userData, "password")
		}
		if userData["pass"] != nil {
			secret = userData["pass"].(string)
			delete(userData, "pass")
		}
	}
	if userData == nil || secret == "" {
		c.JSON(http.StatusOK, kuu.StdDataError(safeL(c, "login_error")))
		return
	}
	if b, err := json.Marshal(userData); err == nil {
		data := jwt.MapClaims{}
		json.Unmarshal(b, &data)
		token = Encoded(data, secret)
	}
	if token != "" {
		userData[tokenKey] = token
	}
	c.JSON(http.StatusOK, kuu.StdDataOK(userData))
}

// logout 退出登录路由
func logout(c *gin.Context) {
	c.JSON(http.StatusOK, kuu.StdDataError(safeL(c, "logout")))
}

// valid 验证路由
func valid(c *gin.Context) {
	data := encodeData(c)
	if data == nil {
		c.JSON(http.StatusOK, kuu.StdDataError(safeL(c, "auth_error")))
	} else {
		c.JSON(http.StatusOK, kuu.StdDataOK(data))
	}
}

// Plugin 插件声明
func Plugin(l func(*gin.Context) (kuu.H, string), s func(string) string, f func(*gin.Context) bool, eh func(*gin.Context), t string) *kuu.Plugin {
	// 必填参数
	if l == nil || s == nil {
		panic("Config required.")
	}
	loginFunc = l
	secretFunc = s
	// 可选参数
	if t != "" {
		tokenKey = t
	}
	if f != nil {
		filterFunc = f
	}
	if eh != nil {
		errorHandler = eh
	}
	return &kuu.Plugin{
		Name: "ac",
		Middleware: kuu.M{
			"auth": AuthMiddleware,
		},
		Routes: kuu.R{
			"login": &kuu.Route{
				Method:  "POST",
				Path:    "/login",
				Handler: login,
			},
			"logout": &kuu.Route{
				Method:  "POST",
				Path:    "/logout",
				Handler: logout,
			},
			"valid": &kuu.Route{
				Method:  "GET",
				Path:    "/valid",
				Handler: valid,
			},
		},
		Methods: kuu.Methods{
			"encoded": func(args ...interface{}) interface{} {
				if args != nil && len(args) == 2 {
					data := args[0].(jwt.MapClaims)
					secret := args[1].(string)
					return Encoded(data, secret)
				}
				return nil
			},
			"decoded": func(args ...interface{}) interface{} {
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
