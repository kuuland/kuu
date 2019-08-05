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
