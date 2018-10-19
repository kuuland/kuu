package kuu

import (
	"fmt"
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

func ExampleL() {
	Langs["en"] = LangMessages{
		"signup": "Sign up",
	}
	Langs["zh-CN"] = LangMessages{
		"signup": "注册",
	}
	Langs["zh-TW"] = LangMessages{
		"signup": "註冊",
	}
	fmt.Println(L(nil, "signup"))
	fmt.Println(L(nil, "signup", "zh-CN"))
	fmt.Println(L(nil, "signup", "zh-TW"))
	// Output:
	// Sign up
	// 注册
	// 註冊
}

func ExampleL_template() {
	Langs["en"] = LangMessages{
		"success": "Congratulations, your ID {{.username}} registered successfully.",
	}
	Langs["zh-CN"] = LangMessages{
		"success": "恭喜，你的账号 {{.username}} 已注册成功！",
	}
	Langs["zh-TW"] = LangMessages{
		"success": "恭喜，你的賬號 {{.username}} 已註冊成功！",
	}
	fmt.Println(L(nil, "success", nil, H{
		"username": "kuu",
	}))
	fmt.Println(L(nil, "success", "zh-CN", H{
		"username": "kuu",
	}))
	fmt.Println(L(nil, "success", "zh-TW", H{
		"username": "kuu",
	}))
	// Output:
	// Congratulations, your ID kuu registered successfully.
	// 恭喜，你的账号 kuu 已注册成功！
	// 恭喜，你的賬號 kuu 已註冊成功！
}

func ExampleL_defaultLang() {
	fmt.Println(L(nil, "signup"))
	DefaultLang = "zh-CN"
	fmt.Println(L(nil, "signup"))
	DefaultLang = "zh-TW"
	fmt.Println(L(nil, "signup"))
	// Output:
	// Sign up
	// 注册
	// 註冊
}
