package kuu

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"regexp"
	"strings"
)

// LoginHandlerFunc
type LoginHandlerFunc func(*Context) (jwt.MapClaims, uint, error)

var (
	TokenKey       = "Token"
	WhiteList      = []string{"POST /api/login", "POST /login"}
	ExpiresSeconds = 86400
	SignContextKey = "SignContext"
	loginHandler   = defaultLoginHandler
)

const (
	RedisSecretKey = "secret"
	RedisOrgKey    = "org"
)

// InWhiteList
func InWhiteList(c *gin.Context) bool {
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

func saveHistory(c *Context, secretData *SignSecret) {
	var body map[string]interface{}
	c.ShouldBindBodyWith(&body, binding.JSON)
	history := SignHistory{
		Request: Stringify(map[string]interface{}{
			"headers": c.Request.Header,
			"query":   c.Request.URL.Query(),
			"body":    body,
		}),
		SecretID:   secretData.ID,
		SecretData: secretData.Secret,
		Token:      secretData.Token,
		Method:     secretData.Method,
	}
	DB().Create(&history)
}

// GenRedisKey
func RedisKeyBuilder(keys ...string) string {
	args := []string{RedisPrefix}
	for _, k := range keys {
		args = append(args, k)
	}
	return strings.Join(args, "_")
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

// DecodedContext
func DecodedContext(c *gin.Context) (*SignContext, error) {
	token := ParseToken(c)
	if token == "" {
		return nil, errors.New(L(c, "未找到令牌"))
	}
	data := SignContext{Token: token}
	// 解析UID
	var secret SignSecret
	if v, err := RedisClient.Get(RedisKeyBuilder(RedisSecretKey, token)).Result(); err == nil {
		Parse(v, &secret)
	} else {
		DB().Where(&SignSecret{Token: token}).Find(&secret)
	}
	data.UID = secret.UID
	// 解析OrgID
	var org SignOrg
	if v, err := RedisClient.Get(RedisKeyBuilder(RedisOrgKey, token)).Result(); err == nil {
		Parse(v, &org)
	} else {
		DB().Where(&SignOrg{Token: token}).Find(&org)
	}
	data.OrgID = org.OrgID
	// 验证令牌
	if secret.Secret == "" {
		return nil, errors.New(Lang(c, "secret_invalid", "Secret is invalid: {{uid}} {{token}}", gin.H{"uid": data.UID, "token": token}))
	}
	if secret.Method == "LOGOUT" {
		return nil, errors.New(Lang(c, "token_expired", "Token has expired: '{{token}}'", gin.H{"token": token}))
	}
	data.Secret = &secret
	data.Payload = DecodedToken(token, secret.Secret)
	if data.IsValid() {
		c.Set(SignContextKey, &data)
	}
	return &data, nil
}

// EncodedToken
func EncodedToken(claims jwt.MapClaims, secret string) (signed string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err = token.SignedString([]byte(secret))
	if err != nil {
		return
	}
	return
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

// Accounts
func Accounts(handler ...LoginHandlerFunc) *Mod {
	if len(handler) > 0 {
		loginHandler = handler[0]
	}
	return &Mod{
		Models: []interface{}{
			&SignSecret{},
			&SignHistory{},
		},
		Middleware: gin.HandlersChain{
			AuthMiddleware,
		},
		Routes: RoutesInfo{
			LoginRoute,
			LogoutRoute,
			ValidRoute,
		},
	}
}
