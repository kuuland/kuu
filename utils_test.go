package kuu

import "testing"

func TestEnsureDir(t *testing.T) {
	EnsureDir("logs")
}

func TestEmptyDir(t *testing.T) {
	EmptyDir("logs")
}
