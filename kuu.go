package kuu

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/jtolds/gls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	// PrisDescKey
	PrisDescKey = "PrisDesc"
	// SignInfoKey
	SignInfoKey = "SignInfo"
	// IgnoreAuthKey
	IgnoreAuthKey = "IgnoreAuth"
	// ValuesKey
	ValuesKey = "Values"
)

// KuuHandlerFunc defines the handler used by ok middleware as return value.
type KuuHandlerFunc func(*Context)

// KuuHandlersChain defines a KuuHandlerFunc array.
type KuuHandlersChain []KuuHandlerFunc

// RouteInfo represents a request route's specification which contains method and path and its handler.
type KuuRouteInfo struct {
	Method      string
	Path        string
	HandlerFunc KuuHandlerFunc
}

// RoutesInfo defines a RouteInfo array.
type KuuRoutesInfo []KuuRouteInfo

// Engine
type Engine struct {
	*gin.Engine
}

// Values
type Values map[string]interface{}

// Context
type Context struct {
	*gin.Context
	SignInfo *SignContext
	PrisDesc *PrivilegesDesc
	Values   *Values
}

// L
func (c *Context) L(defaultValue string, args ...interface{}) string {
	return L(c.Context, defaultValue, args...)
}

// Lang
func (c *Context) Lang(key string, defaultValue string, args interface{}) string {
	return Lang(c.Context, key, defaultValue, args)
}

// STD
func (c *Context) STD(data interface{}, msg ...string) *STDRender {
	return STD(c.Context, data, msg...)
}

// STDErr
func (c *Context) STDErr(msg string, code ...int32) *STDRender {
	return STDErr(c.Context, msg, code...)
}

// STDHold
func (c *Context) STDHold(data interface{}, msg ...string) *STDRender {
	return STDHold(c.Context, data, msg...)
}

// STDErrHold
func (c *Context) STDErrHold(msg string, code ...int32) *STDRender {
	return STDErrHold(c.Context, msg, code...)
}

// SetValue
func (c *Context) SetValue(key string, value interface{}) {
	(*c.Values)[key] = value
}

// DelValue
func (c *Context) DelValue(key string) {
	delete((*c.Values), key)
}

// GetValue
func (c *Context) GetValue(key string) interface{} {
	return (*c.Values)[key]
}

// IgnoreAuth
func (c *Context) IgnoreAuth(cancel ...bool) {
	if len(cancel) > 0 && cancel[0] == true {
		c.DelValue(IgnoreAuthKey)
	} else {
		c.SetValue(IgnoreAuthKey, true)
	}
}

// Default
func Default() (e *Engine) {
	e = &Engine{Engine: gin.Default()}
	onInit(e)
	return
}

// New
func New() (e *Engine) {
	e = &Engine{Engine: gin.New()}
	onInit(e)
	return
}

// SetValues
func SetValues(values gls.Values, call func()) {
	mgr.SetValues(values, call)
}

// GetValue
func GetValue(key interface{}) (value interface{}, ok bool) {
	return mgr.GetValue(key)
}

func shutdown(srv *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	INFO("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		FATAL("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		INFO("timeout of 5 seconds.")
	}
	Release()
	INFO("Server exiting")
}

// Run
func (e *Engine) Run(addr ...string) {
	address := resolveAddress(addr)
	srv := &http.Server{
		Addr:    address,
		Handler: e.Engine,
	}
	go func() {
		INFO("Listening and serving HTTP on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			FATAL("listen: %s", err)
		}
	}()
	shutdown(srv)
}

// RunTLS
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	srv := &http.Server{
		Addr:    addr,
		Handler: e.Engine,
	}
	go func() {
		INFO("Listening and serving HTTP on %s", addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			FATAL("listen: %s", err)
		}
	}()
	shutdown(srv)
}

// RESTful
func (e *Engine) RESTful(values ...interface{}) {
	if len(values) == 0 {
		return
	}
	for _, v := range values {
		if v != nil {
			RESTful(e, v)
		}
	}
}

func (e *Engine) convertGinHandlers(chain KuuHandlersChain) (handlers gin.HandlersChain) {
	handlers = make(gin.HandlersChain, len(chain))
	for index, handler := range chain {
		handlers[index] = func(c *gin.Context) {
			kc := &Context{
				Context: c,
			}
			if !InWhiteList(c) {
				sign := GetSignContext(c)
				desc := GetPrivilegesDesc(c)
				vals := make(Values)
				kc.SignInfo = sign
				kc.PrisDesc = desc
				kc.Values = &vals
			}
			glsVals := make(gls.Values)
			glsVals[SignInfoKey] = kc.SignInfo
			glsVals[PrisDescKey] = kc.PrisDesc
			glsVals[ValuesKey] = kc.Values
			SetValues(glsVals, func() { handler(kc) })
		}
	}
	return
}

// Overrite r.Group
func (e *Engine) Group(relativePath string, handlers ...KuuHandlerFunc) *gin.RouterGroup {
	return e.Engine.Group(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.Handle
func (e *Engine) Handle(httpMethod, relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.Handle(httpMethod, relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.POST
func (e *Engine) POST(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.POST(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.GET
func (e *Engine) GET(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.GET(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.DELETE
func (e *Engine) DELETE(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.DELETE(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.PATCH
func (e *Engine) PATCH(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.PATCH(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.PUT
func (e *Engine) PUT(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.PUT(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.OPTIONS
func (e *Engine) OPTIONS(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.OPTIONS(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.HEAD
func (e *Engine) HEAD(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.HEAD(relativePath, e.convertGinHandlers(handlers)...)
}

// Overrite r.Any
func (e *Engine) Any(relativePath string, handlers ...KuuHandlerFunc) gin.IRoutes {
	return e.Engine.Any(relativePath, e.convertGinHandlers(handlers)...)
}

func (e *Engine) initConfigs() {
	if _, exists := C().Get("cors"); exists {
		if C().GetBool("cors") {
			e.Use(cors.Default())
		} else {
			v := C().GetInterface("cors")
			var config cors.Config
			GetSoul(v, &config)
			e.Use(cors.New(config))
		}
	}

	if _, exists := C().Get("gzip"); exists {
		if C().GetBool("gzip") {
			e.Use(gzip.Gzip(gzip.DefaultCompression))
		} else {
			v := C().GetInt("gzip")
			if v != 0 {
				e.Use(gzip.Gzip(v))
			}
		}
	}
}

func resolveAddress(addr []string) string {
	switch len(addr) {
	case 0:
		return ":8080"
	case 1:
		return addr[0]
	default:
		panic("too much parameters")
	}
}

func connectedPrint(name, args string) {
	INFO("%-8s is connected: %s", name, args)
}

func onInit(e *Engine) {
	initDataSources()
	initRedis()
	e.initConfigs()

	gorm.DefaultCallback.Query().Before("gorm:query").Register("kuu:query", QueryCallback)
	gorm.DefaultCallback.Update().Before("gorm:update").Register("kuu:update", UpdateCallback)
	gorm.DefaultCallback.Delete().After("gorm:delete").Register("kuu:delete", DeleteCallback)
	gorm.DefaultCallback.Create().Before("gorm:create").Register("kuu:create", CreateCallback)
}
