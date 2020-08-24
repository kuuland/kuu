package kuu

import (
	"testing"
)

func TestGetIPInfo(t *testing.T) {
	if info, err := GetIPInfo("13.229.188.59"); err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}
}
