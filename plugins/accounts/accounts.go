package accounts

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

var (
	tokenKey        = "token"
	loginFunc       func(*gin.Context) (kuu.H, string)
	secretFunc      func(string) string
	secretResetFunc func(string, kuu.H)
	whilelist       = []string{
		"POST /login",
	}
	filterFunc = func(c *gin.Context) bool {
		return false
	}
	errorHandler = func(c *gin.Context) {
		data := kuu.StdDataError(kuu.SafeL(defaultMessages, c, "auth_error"))
		c.AbortWithStatusJSON(http.StatusOK, data)
	}
	defaultMessages = map[string]string{
		"login_error": "Login failed, please contact the administrator or try again later.",
		"auth_error":  "Your session has expired, please log in again.",
		"logout":      "Logout successful.",
	}
)

func init() {
	// kuu.Emit("OnPluginLoad", func(args ...interface{}) {
	// 	k := args[0].(*kuu.Kuu)
	// })
}

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
	var data jwt.MapClaims
	if token != "" && secret != "" {
		data = Decoded(token, secret)
	}
	return data
}

func whilelistFilter(c *gin.Context) bool {
	wl := false
	if len(whilelist) > 0 {
		value := kuu.Join(c.Request.Method, " ", c.Request.URL.String())
		value = strings.ToLower(value)
		for _, item := range whilelist {
			item = strings.ToLower(item)
			if strings.Contains(value, item) {
				wl = true
				break
			}
		}
	}
	return wl
}

// AuthMiddleware 令牌鉴权中间件
func AuthMiddleware(c *gin.Context) {
	data := encodeData(c)
	wl := whilelistFilter(c)
	if wl == false && (data == nil || data.Valid() != nil || !filterFunc(c)) {
		errorHandler(c)
		return
	}
	c.Next()
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
		c.JSON(http.StatusOK, kuu.StdDataError(kuu.SafeL(defaultMessages, c, "login_error")))
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
	jwtData := encodeData(c)
	if jwtData != nil && secretResetFunc != nil {
		token := parseToken(c)
		if b, err := json.Marshal(jwtData); err == nil {
			data := kuu.H{}
			json.Unmarshal(b, &data)
			secretResetFunc(token, data)
		}
	}
	c.JSON(http.StatusOK, kuu.StdDataError(kuu.SafeL(defaultMessages, c, "logout")))
}

// valid 验证路由
func valid(c *gin.Context) {
	data := encodeData(c)
	if data == nil {
		c.JSON(http.StatusOK, kuu.StdDataError(kuu.SafeL(defaultMessages, c, "auth_error")))
	} else {
		c.JSON(http.StatusOK, kuu.StdDataOK(data))
	}
}

// Plugin 插件声明
func Plugin(
	getUserData func(*gin.Context) (kuu.H, string),
	getUserSecretByToken func(string) string,
	setUserSecretByToken func(string, kuu.H),
	tokenFilter func(*gin.Context) bool,
	onError func(*gin.Context),
	customTokenKey string) *kuu.Plugin {
	// 必填参数
	if getUserData == nil || getUserSecretByToken == nil || setUserSecretByToken == nil {
		panic("Config required.")
	}
	loginFunc = getUserData
	secretFunc = getUserSecretByToken
	secretResetFunc = setUserSecretByToken
	// 可选参数
	if customTokenKey != "" {
		tokenKey = customTokenKey
	}
	if tokenFilter != nil {
		filterFunc = tokenFilter
	}
	if onError != nil {
		errorHandler = onError
	}
	return &kuu.Plugin{
		Name: "ac",
		Middleware: kuu.Middleware{
			AuthMiddleware,
		},
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "POST",
				Path:    "/login",
				Handler: login,
			},
			kuu.RouteInfo{
				Method:  "POST",
				Path:    "/logout",
				Handler: logout,
			},
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/valid",
				Handler: valid,
			},
		},
		KuuMethods: kuu.KuuMethods{
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
			"whilelist": func(args ...interface{}) interface{} {
				list := args[0].([]string)
				replace := false
				if len(args) > 1 {
					replace = args[1].(bool)
				}
				if list != nil && len(list) > 0 {
					m := map[string]bool{}
					z := make([]string, len(whilelist)+len(list))
					l := []([]string){}
					if replace {
						l = append(l, list)
					} else {
						l = append(l, whilelist)
						l = append(l, list)
					}
					offset := 0
					for _, arr := range l {
						for i, item := range arr {
							if m[item] {
								continue
							}
							m[item] = true
							z[i+offset] = item
						}
						offset += len(arr)
					}
					whilelist = z
				}
				return nil
			},
		},
	}
}
