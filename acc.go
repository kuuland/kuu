package kuu

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"regexp"
	"strings"
)

// LoginHandlerFunc
type LoginHandlerFunc func(*Context) *LoginHandlerResponse

// LoginHandlerResponse
type LoginHandlerResponse struct {
	Username                   string
	Password                   string
	Payload                    jwt.MapClaims
	Lang                       string
	UID                        uint
	Error                      error
	LocaleMessageID            string
	LocaleMessageDefaultText   string
	LocaleMessageContextValues interface{}
}

var (
	TokenKey  = "Token"
	LangKey   = "Lang"
	Whitelist = []interface{}{
		"GET /",
		"GET /favicon.ico",
		"POST /login",
		"GET /model/docs",
		"GET /model/ws",
		"GET /language",
		"GET /langmsgs",
		"GET /captcha",
		"GET /intl/languages",
		"GET /intl/messages",
		regexp.MustCompile("GET /assets"),
	}
	ExpiresSeconds = 86400
	loginHandler   = defaultLoginHandler
)

const (
	AdminSignType = "ADMIN"
)

// InWhitelist
func (c *Context) InWhitelist() bool {
	if len(Whitelist) == 0 {
		return false
	}
	cacheKey := "__kuu_path_in_whitelist__"
	if v, has := c.Get(cacheKey); has && v != nil {
		b, _ := v.(bool)
		return b
	}
	var (
		input  = fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path) // 格式为：GET /api/user
		result bool
	)
	for _, item := range Whitelist {
		if v, ok := item.(string); ok {
			// 字符串忽略大小写
			lowerInput := strings.ToLower(input)
			v = strings.ToLower(v)
			prefix := C().GetString("prefix")
			if v == lowerInput {
				// 完全匹配
				result = true
				break
			} else if C().DefaultGetBool("whitelist:prefix", true) && prefix != "" {
				// 加上全局prefix匹配
				old := strings.ToLower(fmt.Sprintf("%s ", c.Request.Method))
				with := strings.ToLower(fmt.Sprintf("%s %s", c.Request.Method, prefix))
				v = strings.Replace(v, old, with, 1)
				if v == lowerInput {
					result = true
					break
				}
			}
		} else if v, ok := item.(*regexp.Regexp); ok {
			// 正则匹配
			if v.MatchString(input) {
				result = true
				break
			}
		}
	}
	c.Set(cacheKey, result)
	return result
}

// AddWhitelist support string and *regexp.Regexp.
func AddWhitelist(rules ...interface{}) {
	Whitelist = append(Whitelist, rules...)
}

func saveHistory(secretData *SignSecret) {
	history := SignHistory{
		UID:        secretData.UID,
		SecretID:   secretData.ID,
		SecretData: secretData.Secret,
		Token:      secretData.Token,
		Method:     secretData.Method,
	}
	DB().Create(&history)
}

func (c *Context) Token() string {
	v := c.GetKey("Authorization", "api_key", TokenKey)
	return strings.TrimPrefix(v, "Bearer ")
}

func (c *Context) GetKey(names ...string) (value string) {
	if len(names) == 0 {
		return
	}
	cacheKey := fmt.Sprintf("__%s__", strings.Join(names, "_"))
	if v, has := c.Get(cacheKey); has {
		return v.(string)
	}
	// querystring > header > cookie
	for _, name := range names {
		if name == "" {
			continue
		}
		value = c.QueryCI(name)
		if value != "" {
			return value
		}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		value = c.GetHeader(name)
		if value != "" {
			return value
		}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		for _, s := range []string{strings.ToUpper(name), name} {
			value, _ = c.Cookie(s)
			if value != "" {
				return value
			}
		}
	}
	c.Set(cacheKey, value)
	return
}

// DecodedContext
func (c *Context) DecodedContext() (sign *SignContext, err error) {
	cacheKey := "__kuu_sign_context__"
	if v, has := c.Get(cacheKey); has {
		return v.(*SignContext), nil
	}

	token := c.Token()
	if token == "" {
		return nil, ErrTokenNotFound
	}
	sign = &SignContext{Token: token, Lang: c.Lang()}

	// 解析UID
	secret, err := GetSignSecret(token)
	if secret == nil || err != nil {
		return
	}
	sign.UID = secret.UID
	sign.Username = secret.Username
	// 验证令牌
	if secret.Secret == "" {
		err = ErrSecretNotFound
		return
	}
	if secret.Method == SignMethodLogout {
		err = ErrInvalidToken
		return
	}
	sign.Secret = secret
	if secret.Type == "" {
		secret.Type = AdminSignType
	}
	sign.Type = secret.Type
	if v, err := DecodedToken(token, secret.Secret); err != nil {
		return nil, err
	} else {
		sign.Payload = v
	}
	sign.SubDocID = secret.SubDocID
	if sign.SubDocID == 0 { // 当取SubDocID失败时，查用户数据（因为令牌签发可能在子档案创建之前）
		user := GetUserFromCache(sign.UID)
		sid, err := user.GetSubDocID(sign.Type)
		if err != nil {
			return nil, err
		}
		sign.SubDocID = sid
	}
	if !sign.IsValid() {
		return nil, ErrInvalidToken
	}
	var payload jwt.MapClaims
	if err := JSONParse(secret.Payload, &payload); err == nil {
		sign.Payload = payload
	}
	c.Set(cacheKey, sign)
	return
}

func GetSignSecret(token string, tx ...*gorm.DB) (*SignSecret, error) {
	if token == "" {
		return nil, nil
	}
	// 优先从缓存取
	var secret SignSecret
	s := GetCacheString(token)
	if s != "" {
		err := JSONParse(s, &secret)
		if err != nil {
			return nil, err
		}
		return &secret, nil
	}
	var db *gorm.DB
	if len(tx) > 0 && tx[0] != nil {
		db = tx[0]
	} else {
		db = DB()
	}
	// 没有再从数据库取
	err := db.Where(&SignSecret{Token: token}).Find(&secret).Error
	if err != nil {
		return nil, err
	}
	return &secret, nil
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
func DecodedToken(tokenString string, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if token != nil {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			return claims, nil
		}
	}
	ERROR(err)
	return nil, errors.New("token invalid")
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
		Middleware: HandlersChain{
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
