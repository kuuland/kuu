package kuu

import (
	"testing"
)

// TestGenerateRandomCode
func TestGenerateRandomCode(t *testing.T) {
	t.Log(GenerateRandomCode(4))
	t.Log(GenerateRandomCode(6))
	t.Log(GenerateRandomCode())
	t.Log(GenerateRandomCode(10))
}
