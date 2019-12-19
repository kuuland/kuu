package kuu

import "testing"

func TestGenPassword(t *testing.T) {
	t.Log(GenPassword())
	t.Log(GenPassword(8))
	t.Log(GenPassword(10))
	t.Log(GenPassword(20))
}
