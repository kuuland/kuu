package kuu

import (
	"testing"
)

func TestGetIPInfo(t *testing.T) {
	if info, err := GetIPInfo("61.140.27.206"); err != nil {
		t.Error(err)
	} else {
		t.Log(info)
	}
}
