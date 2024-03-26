package kuu

import (
	"encoding/json"
	"os"
	"strings"
)

var tokenKeys = []string{
	"MOD-TOKEN",
	"MOD-API-KEY",
}

func init() {
	if s := strings.TrimSpace(os.Getenv(ConfigTokenKey)); s != "" {
		keys := strings.Split(s, ",")
		if len(keys) > 0 {
			tokenKeys = keys
		}
	}
	Infof("TOKEN_KEYS: %s\n", strings.Join(tokenKeys, ","))
}

type Token struct {
	value string
	cache Cache
}

func NewToken(value string, cache Cache) *Token {
	return &Token{
		value: value,
		cache: cache,
	}
}

func (t *Token) Value() string {
	return t.value
}

func (t *Token) ContextString() string {
	v, err := t.cache.GetToken(t.value)
	if err != nil {
		return ""
	}
	return v
}

func (t *Token) ContextParser(dst any) error {
	s := t.ContextString()
	return json.Unmarshal([]byte(s), dst)
}

func (t *Token) Valid() bool {
	s := t.ContextString()
	return s != ""
}
