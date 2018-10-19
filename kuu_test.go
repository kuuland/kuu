package kuu

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
)

func ExampleNew() {
	k := New(H{
		"name": "kuu",
	})
	k.GET("/", func(c *gin.Context) {
		c.String(200, "hello")
	})
	fmt.Printf("Hello %s.\n", k.Name)
	k.Run(":8080")
	// Output:
	// Hello kuu.
}

func TestStd(t *testing.T) {
	data, err := json.Marshal(Std(1001, "sign up success", 0))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
func ExampleStd() {
	data, err := json.Marshal(Std(1001, "sign up success", 0))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output:
	// {"code":0,"data":1001,"msg":"sign up success"}
}

func TestStdOK(t *testing.T) {
	data, err := json.Marshal(StdOK("kuu"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
func ExampleStdOK() {
	data, err := json.Marshal(StdOK("kuu"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output:
	// {"code":0,"data":"kuu"}
}
func TestStdError(t *testing.T) {
	data, err := json.Marshal(StdError("Login failed"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
}
func ExampleStdError() {
	data, err := json.Marshal(StdError("Login failed"))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(data))
	// Output:
	// {"code":-1,"msg":"Login failed"}
}
