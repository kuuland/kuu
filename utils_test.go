package kuu

import (
	"testing"
)

// TestRandCode
func TestRandCode(t *testing.T) {
	t.Log(RandCode(4))
	t.Log(RandCode(6))
	t.Log(RandCode())
	t.Log(RandCode(10))
}

func TestSetUrlQuery(t *testing.T) {
	var u string

	u = "/mp"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))

	u = "/mp?a=1&b=2"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))

	u = "https://www.example.com/mp"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))

	u = "https://www.example.com/mp?a=1&b=2"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))

	u = "https://www.example.com/admin/?a=1&b=2#/test?h=1&j=ok"
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}))
	t.Log(SetUrlQuery(u, map[string]interface{}{"t": 1, "s": "hello"}, true))
}
