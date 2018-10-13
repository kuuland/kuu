package accounts

import (
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
)

func TestEncoded(t *testing.T) {
	token := Encoded(jwt.MapClaims{
		"Username": "yinfxs",
		"Password": "123456",
	}, "kuu")
	fmt.Println(token)
}

func TestDecoded(t *testing.T) {
	token := Decoded("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJQYXNzd29yZCI6IjEyMzQ1NiIsIlVzZXJuYW1lIjoieWluZnhzIn0.J5EwLITzO-fXoYu0QnL2W0j_h-TQh2XUl4ZKWOYoOBY", "kuu")
	fmt.Println(token["Username"])
}
