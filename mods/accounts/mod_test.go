package accounts

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/kuuland/kuu/mods/accounts/utils"
)

func TestEncoded(t *testing.T) {
	claims := jwt.MapClaims{
		"Username": "yinfxs",
		"Password": "123456",
		"exp":      time.Now().Add(time.Second * time.Duration(30)).Unix(),
		"iat":      time.Now().Unix(),
	}
	token := utils.Encoded(claims, "kuu")
	fmt.Println(token)
}

func TestDecoded(t *testing.T) {
	token := utils.Decoded("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJQYXNzd29yZCI6IjEyMzQ1NiIsIlVzZXJuYW1lIjoieWluZnhzIiwiZXhwIjoxNTQ4MDYzMjY1LCJpYXQiOjE1NDgwNjMyMzV9.hhnaxf2BB8mq5WsJk84ZLKNuD-ySGKK9GDiclMvKk4E", "kuu")
	fmt.Println(token.Valid())
	fmt.Println(token)
}
