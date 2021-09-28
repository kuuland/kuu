package kuu

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"gopkg.in/guregu/null.v3"
	"time"
)

// SignHistory
type SignHistory struct {
	gorm.Model `rest:"*" displayName:"登录历史"`
	UID        uint   `name:"用户ID"`
	SecretID   uint   `name:"密钥ID"`
	SecretData string `name:"密钥"`
	Token      string `name:"令牌" gorm:"NOT NULL;INDEX:kuu_token;size:767"`
	Method     string `name:"登录/登出"`
}

// SignSecret
type SignSecret struct {
	gorm.Model `rest:"*" displayName:"令牌密钥"`
	UID        uint      `name:"用户ID"`
	Username   string    `name:"用户账号"`
	SubDocID   uint      `name:"扩展档案ID"`
	Desc       string    `name:"令牌描述"`
	Secret     string    `name:"令牌密钥"`
	Payload    string    `name:"令牌数据(JSON-String)" gorm:"type:text"`
	Token      string    `name:"令牌" gorm:"NOT NULL;INDEX:kuu_token;size:767"`
	Iat        int64     `name:"令牌签发时间戳"`
	Exp        int64     `name:"令牌过期时间戳"`
	Method     string    `name:"登录/登出"`
	IsAPIKey   null.Bool `name:"是否API Key"`
	Type       string    `name:"令牌类型"`
}

// SignContext
type SignContext struct {
	Token    string
	Type     string
	Lang     string
	UID      uint
	Username string
	SubDocID uint
	Payload  jwt.MapClaims
	Secret   *SignSecret
}

// IsValid
func (s *SignContext) IsValid() (ret bool) {
	if s == nil {
		return
	}
	// 双重验证
	if err := s.Payload.Valid(); err == nil && s.verifyExp() && s.Token != "" && s.UID != 0 && s.Secret != nil {
		ret = true
	}
	return
}

func (s *SignContext) verifyExp() bool {
	exp := s.getPayloadExp()
	if exp == 0 {
		return true
	}
	return time.Now().Unix() <= exp
}

func (s *SignContext) getPayloadExp() int64 {
	var exp interface{}
	if v, ok := s.Payload["exp"]; ok {
		exp = v
	} else if v, ok := s.Payload["Exp"]; ok { // 兼容处理：部分已存在的令牌使用了大写
		exp = v
	}
	switch v := exp.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}
