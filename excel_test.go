package kuu

import (
	"testing"
)

func TestGetXAxisNames(t *testing.T) {
	_10 := GetXAxisNames(10)
	if _10[len(_10)-1] != "J" {
		t.Errorf("wrong x-axis name: %v\n", _10)
	} else {
		t.Log(_10)
	}

	_27 := GetXAxisNames(27)
	if _27[len(_27)-1] != "AA" {
		t.Errorf("wrong x-axis name: %v\n", _27)
	} else {
		t.Log(_27)
	}

	_52 := GetXAxisNames(52)
	if _52[len(_52)-1] != "AZ" {
		t.Errorf("wrong x-axis name: %v\n", _52)
	} else {
		t.Log(_52)
	}
}
