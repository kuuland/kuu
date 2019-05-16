package kuu

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"regexp"
	"time"
)

// SignHistory
type SignHistory struct {
	gorm.Model `rest:"*"`
	Request    string
	SecretID   uint
	SecretData string
	Token      string
	Method     string
}

// SignSecret
type SignSecret struct {
	gorm.Model `rest:"*"`
	UID        string
	Secret     string
	Token      string
	Iat        int64
	Exp        int64
	Method     string
}

// SignContext
type SignContext struct {
	Token   string
	UID     string
	Payload jwt.MapClaims
	Secret  *SignSecret
}

// IsValid
func (s *SignContext) IsValid() (ret bool) {
	if err := s.Payload.Valid(); err == nil && s.Token != "" && s.UID != "" {
		ret = true
	}
	return
}

// LoginHandlerFunc
type LoginHandlerFunc func(*gin.Context) (jwt.MapClaims, error)

var (
	TokenKey       = "Token"
	UIDKey         = "UID"
	WhiteList      = []string{"/login"}
	ExpiresSeconds = 86400
	LoginHandler   LoginHandlerFunc
	SignContextKey = "SignContext"
	loginHandler   LoginHandlerFunc
)

// AuthMiddleware
func AuthMiddleware(c *gin.Context) {
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	if whiteListCheck(c) == true {
		c.Next()
	} else {
		// 从请求参数中解码令牌
		sign, err := DecodedContext(c)
		if err != nil {
			ERROR(err)
			std := STDRender{
				Message: err.Error(),
				Code:    555,
			}
			c.AbortWithStatusJSON(http.StatusOK, &std)
			return
		}
		if err := sign.Payload.Valid(); err == nil {
			c.Set(SignContextKey, sign)
			c.Next()
		} else {
			ERROR(err)
			std := STDRender{
				Message: err.Error(),
				Code:    555,
			}
			c.AbortWithStatusJSON(http.StatusOK, &std)
		}
	}
}

func whiteListCheck(c *gin.Context) bool {
	if len(WhiteList) == 0 {
		return false
	}
	input := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path)
	for _, item := range WhiteList {
		reg := regexp.MustCompile(item)
		if reg.MatchString(input) {
			return true
		}
	}
	return false
}

func saveHistory(c *gin.Context, secretData *SignSecret) {
	var body map[string]interface{}
	if err := c.ShouldBindBodyWith(&body, binding.JSON); err != nil {
		ERROR(err)
	}
	history := SignHistory{
		Request: Stringify(map[string]interface{}{
			"headers": c.Request.Header,
			"query":   c.Request.URL.Query(),
			"body":    body,
		}, false),
		SecretID:   secretData.ID,
		SecretData: secretData.Secret,
		Token:      secretData.Token,
		Method:     secretData.Method,
	}
	DB().Create(&history)
}

func genKey(secretData *SignSecret) string {
	key := fmt.Sprintf("%s_%s_%s", RedisPrefix, secretData.UID, secretData.Token)
	return key
}

func deleteFromRedis(secretData *SignSecret) (err error) {
	key := genKey(secretData)
	_, err = RedisClient.Del(key).Result()
	return
}

func saveToRedis(secretData *SignSecret, expiration time.Duration) error {
	key := genKey(secretData)
	value := Stringify(secretData, false)
	if !RedisClient.SetNX(key, value, expiration).Val() {
		return errors.New("Token cache failed")
	}
	return nil
}

// ParseToken
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

// ParseUID
func ParseUID(c *gin.Context) string {
	// querystring > header > cookie
	var userID string
	userID = c.Query(UIDKey)
	if userID == "" {
		userID = c.GetHeader(UIDKey)
	}
	if userID == "" {
		userID, _ = c.Cookie(UIDKey)
	}
	return userID
}

// DecodedContext
func DecodedContext(c *gin.Context) (*SignContext, error) {
	token := ParseToken(c)
	uid := ParseUID(c)
	if token == "" {
		return nil, errors.New("Missing token")
	}
	data := SignContext{
		Token: token,
		UID:   uid,
	}
	var sign SignSecret
	key := fmt.Sprintf("%s_%s_%s", RedisPrefix, uid, token)
	if v, err := RedisClient.Get(key).Result(); err == nil {
		Parse(v, &sign)
	} else {
		DB().Where(&SignSecret{UID: uid, Token: token}).Find(&sign)
	}
	if sign.Secret == "" {
		return nil, errors.New(fmt.Sprintf("Secret is invalid: %s %s", uid, token))
	}
	if sign.Method == "LOGOUT" {
		return nil, errors.New(fmt.Sprintf("Token has expired: '%s'", token))
	}
	data.Payload = DecodedToken(token, sign.Secret)
	return &data, nil
}

// EncodedToken
func EncodedToken(claims jwt.MapClaims, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		ERROR(err)
	}
	return tokenString
}

// DecodedToken
func DecodedToken(tokenString string, secret string) jwt.MapClaims {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if token != nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			return claims
		}
	}
	ERROR(err)
	return nil
}

// LoginHandler
var LoginRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/login",
	HandlerFunc: func(c *gin.Context) {
		// 调用登录处理器获取登录数据
		payload, err := LoginHandler(c)
		if err != nil {
			STDErr(c, err.Error())
			return
		}
		// 设置JWT令牌信息
		expiration := time.Second * time.Duration(ExpiresSeconds)
		iat := time.Now().Unix()
		exp := time.Now().Add(expiration).Unix()
		payload["iat"] = iat // 签发时间
		payload["exp"] = exp // 过期时间
		// 生成新密钥
		secretData := SignSecret{
			UID:    payload[UIDKey].(string),
			Secret: uuid.NewV4().String(),
			Iat:    iat,
			Exp:    exp,
			Method: "LOGIN",
		}
		// 签发令牌
		secretData.Token = EncodedToken(payload, secretData.Secret)
		payload[TokenKey] = secretData.Token
		DB().Create(&secretData)
		// 缓存secret至redis
		if err := saveToRedis(&secretData, expiration); err != nil {
			ERROR(err)
		}
		// 保存登入历史
		saveHistory(c, &secretData)
		// 设置Cookie
		c.SetCookie(TokenKey, secretData.Token, ExpiresSeconds, "/", "", false, true)
		c.SetCookie(UIDKey, secretData.UID, ExpiresSeconds, "/", "", false, true)
		STD(c, payload)
	},
}

// LogoutRoute
var LogoutRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/logout",
	HandlerFunc: func(c *gin.Context) {
		// 从上下文缓存中读取认证信息
		var sign *SignContext
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
		}
		if sign.IsValid() {
			var (
				secretData SignSecret
				db         = DB()
			)
			db.Where(&SignSecret{UID: sign.UID, Token: sign.Token}).First(&secretData)
			if !db.NewRecord(&secretData) {
				if errs := db.Model(&secretData).Updates(&SignSecret{Method: "LOGOUT"}).GetErrors(); len(errs) > 0 {
					ERROR(errs)
					STDErr(c, L(c, "logout_failed", "Logout failed"))
					return
				}
				// 删除redis缓存
				if err := deleteFromRedis(&secretData); err != nil {
					ERROR(err)
				}
				// 保存登出历史
				saveHistory(c, &secretData)
				// 设置Cookie过期
				c.SetCookie(TokenKey, secretData.Token, -1, "/", "", false, true)
				c.SetCookie(UIDKey, secretData.UID, -1, "/", "", false, true)
			}
		}
		STD(c, L(c, "logout_success", "Logout successful"))
	},
}

// ValidRoute
var ValidRoute = gin.RouteInfo{
	Method: "POST",
	Path:   "/valid",
	HandlerFunc: func(c *gin.Context) {
		var sign *SignContext
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
		}
		if sign.IsValid() {
			STD(c, sign.Payload)
		} else {
			STDErr(c, fmt.Sprintf("Token has expired: '%s'", sign.Token), 555)
		}
	},
}

// Accounts
func Accounts(handler LoginHandlerFunc, whiteList ...[]string) *Mod {
	if handler == nil {
		PANIC("Login handler is required")
	}
	loginHandler = handler
	if len(whiteList) > 0 {
		WhiteList = append(WhiteList, whiteList[0]...)
	}
	return &Mod{
		Models: []interface{}{
			&SignSecret{},
			&SignHistory{},
		},
		Middleware: gin.HandlersChain{
			AuthMiddleware,
		},
		Routes: gin.RoutesInfo{
			LoginRoute,
			LogoutRoute,
			ValidRoute,
		},
	}
}
