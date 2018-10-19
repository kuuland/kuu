package kuu

import (
	"fmt"
	"testing"
)

func TestOn(t *testing.T) {
	On("hello", func(data ...interface{}) {
		fmt.Println(data[0])
	})
	Emit("hello", "kuu")
}

func ExampleOn() {
	On("hello", func(data ...interface{}) {
		fmt.Println(data[0])
	})
	Emit("hello", "kuu")
	// Output:
	// kuu
}
