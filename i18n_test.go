package kuu

import (
	"fmt"
	"testing"
)

func init() {
	Langs["en"] = LangMessages{
		"hello":   "Hello",
		"logout":  "Log out",
		"login":   "Log in",
		"signup":  "Sign up",
		"success": "Congratulations, your ID {{.username}} registered successfully.",
	}
	Langs["zh-CN"] = LangMessages{
		"hello":   "你好",
		"logout":  "退出",
		"login":   "登录",
		"signup":  "注册",
		"success": "恭喜，你的账号 {{.username}} 已注册成功！",
	}
	Langs["zh-TW"] = LangMessages{
		"hello":   "你好",
		"logout":  "退出",
		"login":   "登錄",
		"signup":  "註冊",
		"success": "恭喜，你的賬號 {{.username}} 已註冊成功！",
	}
}
func TestLang(t *testing.T) {
	fmt.Println(L(nil, "signup"))
	fmt.Println(L(nil, "signup", "zh-CN"))
	fmt.Println(L(nil, "signup", "zh-TW"))
}

func TestTemplate(t *testing.T) {
	fmt.Println(L(nil, "success", nil, H{
		"username": "kuu",
	}))
	fmt.Println(L(nil, "success", "zh-CN", H{
		"username": "kuu",
	}))
	fmt.Println(L(nil, "success", "zh-TW", H{
		"username": "kuu",
	}))
}

func TestSetLang(t *testing.T) {
	// default lang=en
	fmt.Println("en", L(nil, "signup"))
	// lang=zh-CN
	DefaultLang = "zh-CN"
	fmt.Println("zh-CN", L(nil, "signup"))
	// lang=zh-TW
	DefaultLang = "zh-TW"
	fmt.Println("zh-CN", L(nil, "signup"))
}
