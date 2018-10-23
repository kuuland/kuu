package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
)

var (
	// TokenKey 令牌key
	TokenKey = "token"
	// DataGetter 处理登录逻辑，返回主体数据、GetSecret、SetSecret
	DataGetter func(*gin.Context) (kuu.H, func() string, func())
	// CustomAuthFilter 自定义授权验证
	CustomAuthFilter = func(c *gin.Context) bool {
		return false
	}
	// CustomErrorHandler 自定义错误处理
	CustomErrorHandler = func(c *gin.Context) {
		data := kuu.StdError(kuu.SafeL(defaultMessages, c, "auth_error"))
		c.AbortWithStatusJSON(http.StatusOK, data)
	}
	secretGetters = map[string]func() string{}
	secretSetters = map[string]func(){}
	whilelist     = []string{
		"POST /login",
	}
	defaultMessages = map[string]string{
		"login_error": "Login failed, please contact the administrator or try again later.",
		"auth_error":  "Your session has expired, please log in again.",
		"logout":      "Logout successful.",
	}
)

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

// GetDecodedData 从请求中获取解码数据
func GetDecodedData(c *gin.Context) jwt.MapClaims {
	var (
		secret string
		data   jwt.MapClaims
	)
	token := ParseToken(c)
	if getter := secretGetters[token]; getter != nil {
		secret = getter()
	}
	if token != "" && secret != "" {
		data = Decoded(token, secret)
	}
	return data
}

func whilelistFilter(c *gin.Context) bool {
	wl := false
	if len(whilelist) > 0 {
		value := kuu.Join(c.Request.Method, " ", c.Request.URL.Path)
		value = strings.ToLower(value)
		for _, item := range whilelist {
			if s := strings.Split(item, " "); len(s) == 1 {
				item = kuu.Join(c.Request.Method, " ", item)
			}
			item = strings.ToLower(item)
			if value == item {
				wl = true
				break
			}
		}
	}
	return wl
}

// AddWhilelist 添加白名单
func AddWhilelist(list []string, replace bool) []string {
	if list != nil && len(list) > 0 {
		z := make([]string, len(whilelist)+len(list))
		l := []([]string){}
		if replace {
			l = append(l, list)
		} else {
			l = append(l, whilelist)
			l = append(l, list)
		}
		exists := map[string]bool{}
		offset := 0
		for _, arr := range l {
			for i, item := range arr {
				if exists[item] {
					continue
				}
				exists[item] = true
				z[i+offset] = item
			}
			offset += len(arr)
		}
		whilelist = z
	}
	return whilelist
}

// AuthMiddleware 令牌鉴权中间件
func AuthMiddleware(c *gin.Context) {
	data := GetDecodedData(c)
	c.Set("JWTDecoded", data)
	wl := whilelistFilter(c)
	if wl == false && (data == nil || data.Valid() != nil || !CustomAuthFilter(c)) {
		CustomErrorHandler(c)
		return
	}
	c.Next()
}

// LoginHandler 登录路由
func LoginHandler(c *gin.Context) {
	data, getSecret, setSecret := DataGetter(c)
	var (
		secret string
		token  string
	)
	if data != nil || getSecret != nil || setSecret != nil {
		secretGetters[token] = getSecret
		secretSetters[token] = setSecret
		secret = getSecret()
	}
	if secret == "" {
		c.JSON(http.StatusOK, kuu.StdError(kuu.SafeL(defaultMessages, c, "login_error")))
		return
	}
	if b, err := json.Marshal(data); err == nil {
		data := jwt.MapClaims{}
		json.Unmarshal(b, &data)
		token = Encoded(data, secret)
	}
	if token != "" {
		data[TokenKey] = token
	}
	c.JSON(http.StatusOK, kuu.StdOK(data))
}

// LogoutHandler 退出登录路由
func LogoutHandler(c *gin.Context) {
	token := ParseToken(c)
	jwtData := GetDecodedData(c)
	setSecret := secretSetters[token]
	if token != "" && jwtData != nil && setSecret != nil {
		if b, err := json.Marshal(jwtData); err == nil {
			data := kuu.H{}
			json.Unmarshal(b, &data)
			setSecret()
		}
	}
	c.JSON(http.StatusOK, kuu.StdError(kuu.SafeL(defaultMessages, c, "logout")))
}

// ValidHandler 验证路由
func ValidHandler(c *gin.Context) {
	data := GetDecodedData(c)
	if data == nil {
		c.JSON(http.StatusOK, kuu.StdError(kuu.SafeL(defaultMessages, c, "auth_error")))
	} else {
		c.JSON(http.StatusOK, kuu.StdOK(data))
	}
}

// All 插件声明
func All(dataGetter func(*gin.Context) (kuu.H, func() string, func())) *kuu.Plugin {
	DataGetter = dataGetter
	return &kuu.Plugin{
		Middleware: kuu.Middleware{
			AuthMiddleware,
		},
		Routes: kuu.Routes{
			kuu.RouteInfo{
				Method:  "POST",
				Path:    "/login",
				Handler: LoginHandler,
			},
			kuu.RouteInfo{
				Method:  "POST",
				Path:    "/logout",
				Handler: LogoutHandler,
			},
			kuu.RouteInfo{
				Method:  "GET",
				Path:    "/valid",
				Handler: ValidHandler,
			},
		},
	}
}
