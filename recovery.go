package kuu

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Recovery defined kuu.Engine recovery from panic
// Rewrite `Recovery` if you need
// Tag: 在rest不处理error，除非业务需求(如事物)，直接抛出来
var Recovery gin.HandlerFunc

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}
				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}
				if brokenPipe {
					ERROR("%s\n%s", err, string(httpRequest))
				} else if gin.IsDebugging() {
					ERROR("kuu %s panic recovered:\n%s\n%s\n%s",
						timeFormat(time.Now()), strings.Join(headers, "\r\n"), err, stack)
				} else {
					ERROR("kuu %s panic recovered:\n%s\n%s",
						timeFormat(time.Now()), err, stack)
				}
				// If the connection is dead, we can't write a status to it.
				if brokenPipe {
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
				} else {
					c.String(http.StatusOK, "sorry, sys exception, please try again later")
				}
			}
		}()
		c.Next()
	}
}

func init() {
	Recovery = recovery()
}
