package sys

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/mods/sys/models"
	"github.com/kuuland/kuu/mods/sys/utils"
)

// LoginHandler 登录处理器
func LoginHandler(c *gin.Context) (data kuu.H, err error) {
	kuu.Info("进入登录处理器...")
	User := kuu.Model("User")
	// 解析请求参数
	var inputData map[string]string
	kuu.CopyBody(c, &inputData)
	if inputData == nil {
		return data, errors.New(kuu.L(c, "parse_login_form_error", "解析登录信息失败"))
	}
	var loginUser models.User
	err = User.One(kuu.H{
		"Cond": kuu.H{
			"Username": inputData["username"],
		},
	}, &loginUser)
	// 检测账号是否存在
	if err != nil {
		kuu.Error(err)
		return data, errors.New(kuu.L(c, "account_query_error", "账号查询失败，请稍后重试"))
	}
	if loginUser.ID == "" {
		return data, errors.New(kuu.L(c, "account_not_exist", "账号不存在"))
	}
	// 检测账号是否有效
	if loginUser.Disable {
		return data, errors.New(kuu.L(c, "account_disabled", "账号已被禁用"))
	}
	// 检测密码是否正确
	if !utils.CompareHashAndPassword(loginUser.Password, inputData["password"]) {
		return data, errors.New(kuu.L(c, "account_password_not_match", "账号密码不一致"))
	}
	data = kuu.H{
		"UID":       loginUser.ID,
		"Username":  loginUser.Username,
		"Name":      loginUser.Name,
		"Avatar":    loginUser.Avatar,
		"Sex":       loginUser.Sex,
		"Mobile":    loginUser.Mobile,
		"Email":     loginUser.Email,
		"IsBuiltIn": loginUser.IsBuiltIn,
		"CreatedAt": loginUser.CreatedAt,
		"UpdatedAt": loginUser.UpdatedAt,
	}
	return
}
