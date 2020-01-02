package kuu

import (
	"testing"
)

func TestASCIISigner_Sign(t *testing.T) {
	s := &ASCIISigner{
		PrintRaw: true,
		ToUpper:  true,
		Alg:      ASCIISignerAlgMD5,
		Value: map[string]interface{}{
			"hello":  "world",
			"foo":    "bar",
			"bar":    10,
			"foobar": false,
			"remark": "",
			"price":  0,
		},
		OmitKeys: []string{"a", "b"},
	}
	signature := s.Sign()
	if signature == "" {
		t.Error("invalid signature")
		return
	}
	t.Log(signature)
}
