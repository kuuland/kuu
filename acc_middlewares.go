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
			c.Next()
		} else {
			return c.AbortErrWithCode(err, 555, "acc_incorrect_token", "Incorrect token type")
		}
	}
	return nil
}
