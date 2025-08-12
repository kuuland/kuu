package kuu

// AuthMiddleware
func AuthMiddleware(c *Context) *STDReply {
	if c.InWhitelist() == true {
		c.Next()
	} else {
		// 从请求参数中解码令牌
		sign, err := c.DecodedContext()
		if err != nil {
			return c.AbortErrWithCode(err, 555, "acc_please_login", "Please login")
		}
		if sign.IsValid() {
			if !c.validSignType(sign) {
				return c.AbortErrWithCode(err, 556, "acc_incorrect_token", "Incorrect token type")
			}

			// API权限检查
			if c.PrisDesc != nil && !c.PrisDesc.HasAPIPermission(c.Request.Method, c.Request.URL.Path) {
				return c.AbortErrWithCode(nil, 557, "api_permission_denied", "API access denied")
			}

			c.Next()
		} else {
			return c.AbortErrWithCode(err, 555, "acc_incorrect_token", "Incorrect token type")
		}
	}
	return nil
}
