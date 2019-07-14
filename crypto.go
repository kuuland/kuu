package kuu

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"golang.org/x/crypto/bcrypt"
)

// MD5 加密
func MD5(p string) (v string) {
	h := md5.New()
	h.Write([]byte(p))
	v = hex.EncodeToString(h.Sum(nil))
	return
}

// GenerateFromPassword 生成新密码
func GenerateFromPassword(p string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
	return string(hash), err
}

// CompareHashAndPassword 密码比对
func CompareHashAndPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		ERROR(err)
	}
	return err == nil
}

// Base64Encode Base64编码
func Base64Encode(p string) (v string) {
	v = base64.StdEncoding.EncodeToString([]byte(p))
	return
}

// Base64Decode Base64解码
func Base64Decode(p string) (v string) {
	decoded, err := base64.StdEncoding.DecodeString(p)
	if err != nil {
		ERROR(err)
	}
	v = string(decoded)
	return
}

// Base64URLEncode Base64 URL编码
func Base64URLEncode(p string) (v string) {
	v = base64.URLEncoding.EncodeToString([]byte(p))
	return
}

// Base64URLDecode Base64 URL解码
func Base64URLDecode(p string) (v string) {
	decoded, err := base64.URLEncoding.DecodeString(p)
	if err != nil {
		ERROR(err)
	}
	v = string(decoded)
	return
}
