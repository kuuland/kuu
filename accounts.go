package kuu

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"regexp"
	"strconv"
	"time"
)

// LoginHandlerFunc
type LoginHandlerFunc func(*gin.Context) (jwt.MapClaims, error)

var (
	TokenKey       = "Token"
	UIDKey         = "UID"
	WhiteList      = []string{"/login"}
	ExpiresSeconds = 86400
	SignContextKey = "SignContext"
	loginHandler   LoginHandlerFunc
)

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
		}),
		SecretID:   secretData.ID,
		SecretData: secretData.Secret,
		Token:      secretData.Token,
		Method:     secretData.Method,
	}
	DB().Create(&history)
}

func genKey(secretData *SignSecret) string {
	key := fmt.Sprintf("%s_%d_%s", RedisPrefix, secretData.UID, secretData.Token)
	return key
}

func deleteFromRedis(secretData *SignSecret) (err error) {
	key := genKey(secretData)
	_, err = RedisClient.Del(key).Result()
	return
}

func saveToRedis(secretData *SignSecret, expiration time.Duration) error {
	key := genKey(secretData)
	value := Stringify(secretData)
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
func ParseUID(c *gin.Context) (uid uint) {
	// querystring > header > cookie
	var userID string
	userID = c.Query(UIDKey)
	if userID == "" {
		userID = c.GetHeader(UIDKey)
	}
	if userID == "" {
		userID, _ = c.Cookie(UIDKey)
	}
	if v, err := strconv.ParseUint(userID, 10, 0); err != nil {
		ERROR(err)
	} else {
		uid = uint(v)
	}
	return
}

// DecodedContext
func DecodedContext(c *gin.Context) (*SignContext, error) {
	token := ParseToken(c)
	uid := ParseUID(c)
	if token == "" {
		return nil, errors.New(L(c, "Missing token"))
	}
	data := SignContext{
		Token: token,
		UID:   uid,
	}
	var sign SignSecret
	key := genKey(&SignSecret{UID: uid, Token: token})
	if v, err := RedisClient.Get(key).Result(); err == nil {
		Parse(v, &sign)
	} else {
		DB().Where(&SignSecret{UID: uid, Token: token}).Find(&sign)
	}
	if sign.Secret == "" {
		return nil, errors.New(LFull(c, "secret_invalid", "Secret is invalid: {{uid}} {{token}}", gin.H{"uid": uid, "token": token}))
	}
	if sign.Method == "LOGOUT" {
		return nil, errors.New(LFull(c, "token_expired", "Token has expired: '{{token}}'", gin.H{"token": token}))
	}
	data.Secret = &sign
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
