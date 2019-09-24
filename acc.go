package kuu

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

// LoginHandlerFunc
type LoginHandlerFunc func(*Context) *LoginHandlerResponse

// LoginHandlerResponse
type LoginHandlerResponse struct {
	Username        string
	Password        string
	Payload         jwt.MapClaims
	Lang            string
	UID             uint
	Error           error
	LanguageMessage *LanguageMessage
}

var (
	TokenKey  = "Token"
	Whitelist = []interface{}{
		"GET /",
		"GET /favicon.ico",
		"GET /whitelist",
		"POST /login",
		"GET /enum",
		"GET /meta",
		"GET /model/docs",
		"GET /model/ws",
		"GET /language",
		"GET /langmsgs",
		"GET /captcha",
		regexp.MustCompile("GET /assets"),
	}
	ExpiresSeconds = 86400
	SignContextKey = "SignContext"
	loginHandler   = defaultLoginHandler
)

const (
	RedisSecretKey = "secret"
	RedisOrgKey    = "org"
)

// InWhitelist
func InWhitelist(c *gin.Context) bool {
	if len(Whitelist) == 0 {
		return false
	}
	input := fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path) // 格式为：GET /api/user
	for _, item := range Whitelist {
		if v, ok := item.(string); ok {
			// 字符串忽略大小写
			lowerInput := strings.ToLower(input)
			v = strings.ToLower(v)
			prefix := C().GetString("prefix")
			if v == lowerInput {
				// 完全匹配
				return true
			} else if C().DefaultGetBool("whitelist:prefix", true) && prefix != "" {
				// 加上全局prefix匹配
				old := strings.ToLower(fmt.Sprintf("%s ", c.Request.Method))
				with := strings.ToLower(fmt.Sprintf("%s %s", c.Request.Method, prefix))
				v = strings.Replace(v, old, with, 1)
				if v == lowerInput {
					return true
				}
			}
		} else if v, ok := item.(*regexp.Regexp); ok {
			// 正则匹配
			if v.MatchString(input) {
				return true
			}
		}
	}
	return false
}

// AddWhitelist support string and *regexp.Regexp.
func AddWhitelist(rules ...interface{}) {
	Whitelist = append(Whitelist, rules...)
}

func saveHistory(secretData *SignSecret) {
	history := SignHistory{
		SecretID:   secretData.ID,
		SecretData: secretData.Secret,
		Token:      secretData.Token,
		Method:     secretData.Method,
	}
	DB().Create(&history)
}

// ParseToken
var ParseToken = func(c *gin.Context) string {
	// querystring > header > cookie
	var token string
	token = c.Query(TokenKey)
	if token == "" {
		token = c.GetHeader(TokenKey)
		if token == "" {
			token = c.GetHeader("Authorization")
		}
		if token == "" {
			token = c.GetHeader("api_key")
		}
	}
	if token == "" {
		token, _ = c.Cookie(TokenKey)
	}
	return token
}

// DecodedContext
func DecodedContext(c *gin.Context) (sign *SignContext, err error) {
	token := ParseToken(c)
	if token == "" {
		return nil, ErrTokenNotFound
	}
	sign = &SignContext{Token: token, Lang: ParseLang(c)}
	// 解析UID
	var secret SignSecret
	if err = DB().Where(&SignSecret{Token: token}).Find(&secret).Error; err != nil {
		return
	}
	sign.UID = secret.UID
	// 验证令牌
	if secret.Secret == "" {
		err = ErrSecretNotFound
		return
	}
	if secret.Method == "LOGOUT" {
		err = ErrInvalidToken
		return
	}
	sign.Secret = &secret
	if secret.Type == "" {
		secret.Type = "ADMIN"
	}
	sign.Type = secret.Type
	sign.Payload = DecodedToken(token, secret.Secret)
	sign.SubDocID = secret.SubDocID
	if sign.IsValid() {
		c.Set(SignContextKey, sign)
	}
	return
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
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
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

// Acc
func Acc(handler ...LoginHandlerFunc) *Mod {
	if len(handler) > 0 {
		loginHandler = handler[0]
	}
	return &Mod{
		Code: "acc",
		Models: []interface{}{
			&SignSecret{},
			&SignHistory{},
		},
		Middlewares: gin.HandlersChain{
			AuthMiddleware,
		},
		Routes: RoutesInfo{
			LoginRoute,
			LogoutRoute,
			ValidRoute,
			APIKeyRoute,
			WhitelistRoute,
		},
	}
}
