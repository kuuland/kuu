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

			var prisDesc *PrivilegesDesc

			// 在中间件模式下手动设置权限描述信息
			if c.PrisDesc == nil && sign != nil {
				prisDesc = GetPrivilegesDesc(sign)
			} else if c.PrisDesc != nil {
				prisDesc = c.PrisDesc
			}

			// API权限检查
			if prisDesc != nil && !prisDesc.HasAPIPermission(c.Request.Method, c.Request.URL.Path) {
				return c.AbortErrWithCode(nil, 557, "api_permission_denied", "API access denied")
			}

			c.Next()
		} else {
			return c.AbortErrWithCode(err, 555, "acc_incorrect_token", "Incorrect token type")
		}
	}
	return nil
}
