package kuu

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
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

//TableName 设置表名
func (SignHistory) TableName() string {
	return "sys_SignHistory"
}

// SignSecret
type SignSecret struct {
	gorm.Model `rest:"*"`
	UID        uint
	Secret     string
	Token      string
	Iat        int64
	Exp        int64
	Method     string
}

//TableName 设置表名
func (SignSecret) TableName() string {
	return "sys_SignSecret"
}

// SignContext
type SignContext struct {
	Token   string
	UID     uint
	OrgID   uint
	Payload jwt.MapClaims
	Secret  *SignSecret
}

// IsValid
func (s *SignContext) IsValid() (ret bool) {
	if s == nil {
		return
	}
	if err := s.Payload.Valid(); err == nil && s.Token != "" && s.UID != 0 && s.Secret != nil {
		ret = true
	}
	return
}
