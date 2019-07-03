package kuu

import (
	"github.com/gin-gonic/gin"
	"regexp"
)

// OrgMiddleware
func OrgMiddleware(c *gin.Context) {
	// 解析登录信息
	var sign *SignContext
	if v, exists := c.Get(SignContextKey); exists {
		sign = v.(*SignContext)
	} else {
		if v, err := DecodedContext(c); err == nil {
			sign = v
		}
	}

	c.Next()

	reg := regexp.MustCompile("/login")
	if reg.MatchString(c.Request.RequestURI) {
		if v, exists := c.Get(SignContextKey); exists {
			sign = v.(*SignContext)
			if err := orgAutoLogin(c, sign); err != nil {
				STDErrHold(c, "组织自动登录失败", err).Abort()
				return
			}
		}
	}
}

func orgAutoLogin(c *gin.Context, sign *SignContext) error {
	if list, err := GetOrgList(c, sign.UID); err != nil {
		return err
	} else if len(*list) == 1 {
		orgs := *list
		first := (orgs)[0]
		_, err := ExecOrgLogin(sign, first.ID)
		return err
	}
	return nil
}
