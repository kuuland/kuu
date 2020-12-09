package kuu

import (
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	session "github.com/go-session/gin-session"
	"github.com/jtolds/gls"
)

var (
	// GLSPrisDescKey
	GLSPrisDescKey = "PrisDesc"
	// GLSSignInfoKey
	GLSSignInfoKey = "SignInfo"
	// GLSIgnoreAuthKey
	GLSIgnoreAuthKey = "IgnoreAuth"
	// GLSRoutineCachesKey
	GLSRoutineCachesKey = "RoutineCaches"
	// GLSRequestContextKey
	GLSRequestContextKey = "RequestContext"
	GLSRequestIDKey      = "Request ID"
	// Uptime
	Uptime time.Time
	// IsProduction
	IsProduction = os.Getenv("GIN_MODE") == "release" || os.Getenv("KUU_PROD") == "true"
	json         = jsoniter.ConfigCompatibleWithStandardLibrary
)

func init() {
	if v := os.Getenv("KUU_PROD"); v != "" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// M is a shortcut for map[string]interface{}
type D map[string]interface{}

// Engine
type Engine struct {
	*gin.Engine
}

// RoutineCaches
type RoutineCaches map[string]interface{}

// IgnoreAuth
func (v RoutineCaches) IgnoreAuth(cancel ...bool) {
	if len(cancel) > 0 && cancel[0] == true {
		delete(v, GLSIgnoreAuthKey)
	} else {
		v[GLSIgnoreAuthKey] = true
	}
}

// Default
func Default() (app *Engine) {
	app = &Engine{Engine: gin.Default()}
	app.RemoveExtraSlash = true
	app.UseGin(Recovery)
	if !C().DefaultGetBool("ignoreDefaultRootRoute", false) {
		app.GET("/", func(c *Context) *STDReply {
			return c.STD(fmt.Sprintf("%s is up.", GetAppName()), "kuu_up", "{{time}}", D{"time": Uptime.Format("2006-01-02 15:04:05")})
		})
	}
	app.init()
	return
}

// New
func New() (app *Engine) {
	app = &Engine{Engine: gin.New()}
	app.RemoveExtraSlash = true
	app.init()
	return
}

// SetGLSValues
func SetGLSValues(values gls.Values, call func()) {
	mgr.SetValues(values, call)
}

// GetGLSValue
func GetGLSValue(key interface{}) (value interface{}, ok bool) {
	return mgr.GetValue(key)
}

func beforeRun() {
	if !preflight() {
		for _, mod := range modMap {
			if mod.OnInit != nil {
				if err := mod.OnInit(); err != nil {
					// 模块初始化失败直接退出
					panic(err)
				}
			}
		}
	}
	DefaultCron.Start()
	RunAllRunAfterJobs()
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

// RegisterWhitelist
func (app *Engine) RegisterWhitelist(rules ...interface{}) {
	AddWhitelist(rules...)
}

// Run
func (app *Engine) Run(addr ...string) {
	Uptime = time.Now()
	address := resolveAddress(addr)
	srv := &http.Server{
		Addr:    address,
		Handler: app.Engine,
	}
	beforeRun()
	go func() {
		INFO("Listening and serving HTTP on %s", address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			FATAL("listen: %s", err)
		}
	}()
	shutdown(srv)
}

// RunTLS
func (app *Engine) RunTLS(addr, certFile, keyFile string) {
	Uptime = time.Now()
	srv := &http.Server{
		Addr:    addr,
		Handler: app.Engine,
	}
	beforeRun()
	go func() {
		INFO("Listening and serving HTTP on %s", addr)
		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			FATAL("listen: %s", err)
		}
	}()
	shutdown(srv)
}

func (app *Engine) convertHandlers(chain HandlersChain, isMiddleware ...bool) (handlers gin.HandlersChain) {
	handlers = make(gin.HandlersChain, len(chain))
	var middleware bool
	if len(isMiddleware) > 0 {
		middleware = isMiddleware[0]
	}
	for index, handler := range chain {
		handlers[index] = func(c *gin.Context) {
			var (
				kc = &Context{
					Context: c,
					app:     app,
				}
				v *STDReply
			)
			requestId := kc.RequestID()
			if middleware {
				v = handler(kc)
			} else {
				kc.RoutineCaches = make(RoutineCaches)
				var requestCache struct {
					SignInfo *SignContext
					PrisDesc *PrivilegesDesc
				}
				if s := GetCacheString(requestId); s != "" {
					_ = JSONParse(s, &requestCache)
				}
				var (
					sc  *SignContext
					err error
				)
				if requestCache.SignInfo == nil {
					sc, err = kc.DecodedContext()
					requestCache.SignInfo = sc
				} else {
					sc = requestCache.SignInfo
				}
				if err == nil && sc.IsValid() {
					var desc *PrivilegesDesc
					if requestCache.PrisDesc == nil {
						desc = GetPrivilegesDesc(sc)
						requestCache.PrisDesc = desc
					} else {
						desc = requestCache.PrisDesc
					}
					kc.PrisDesc = desc
					kc.SignInfo = sc
				}
				// 更新请求缓存
				SetCacheString(requestId, JSONStringify(requestCache), 30*time.Minute)
				glsVals := make(gls.Values)
				glsVals[GLSSignInfoKey] = kc.SignInfo
				glsVals[GLSPrisDescKey] = kc.PrisDesc
				glsVals[GLSRoutineCachesKey] = kc.RoutineCaches
				glsVals[GLSRequestContextKey] = kc
				glsVals[GLSRequestIDKey] = requestId
				SetGLSValues(glsVals, func() {
					if kc.InWhitelist() {
						IgnoreAuth()
					}
					v = handler(kc)
				})
			}
			if v != nil {
				switch vv := v.Data.(type) {
				case error:
				case []error:
					ERROR(vv)
				}
				if v.HTTPAction == nil {
					v.HTTPAction = c.JSON
				}
				if v.HTTPCode == 0 {
					v.HTTPCode = http.StatusOK
				}
				v.HTTPAction(v.HTTPCode, v)
			}
		}
	}
	return
}

func (app *Engine) Use(handlers ...HandlerFunc) *Engine {
	app.Engine.Use(app.convertHandlers(handlers, true)...)
	return app
}

func (app *Engine) UseGin(handlers ...gin.HandlerFunc) *Engine {
	app.Engine.Use(handlers...)
	return app
}

// Overrite r.Group
func (app *Engine) Group(relativePath string, handlers ...HandlerFunc) *gin.RouterGroup {
	return app.Engine.Group(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.Handle
func (app *Engine) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.Handle(httpMethod, relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.POST
func (app *Engine) POST(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.POST(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.GET
func (app *Engine) GET(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.GET(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.DELETE
func (app *Engine) DELETE(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.DELETE(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.PATCH
func (app *Engine) PATCH(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.PATCH(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.PUT
func (app *Engine) PUT(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.PUT(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.OPTIONS
func (app *Engine) OPTIONS(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.OPTIONS(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.HEAD
func (app *Engine) HEAD(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.HEAD(relativePath, app.convertHandlers(handlers)...)
}

// Overrite r.Any
func (app *Engine) Any(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return app.Engine.Any(relativePath, app.convertHandlers(handlers)...)
}

// GetRoutinePrivilegesDesc
func GetRoutinePrivilegesDesc() *PrivilegesDesc {
	raw, _ := GetGLSValue(GLSPrisDescKey)
	if !IsBlank(raw) {
		desc := raw.(*PrivilegesDesc)
		if desc.IsValid() {
			return desc
		}
	}
	return nil
}

// GetRoutineCaches
func GetRoutineCaches() RoutineCaches {
	raw, _ := GetGLSValue(GLSRoutineCachesKey)
	if !IsBlank(raw) {
		values := raw.(RoutineCaches)
		return values
	}
	if raw != nil {
		return raw.(RoutineCaches)
	}
	return nil
}

// SetRoutineCache
func SetRoutineCache(key string, value interface{}) {
	values := GetRoutineCaches()
	values[key] = value
}

// GetRoutineCache
func GetRoutineCache(key string) interface{} {
	values := GetRoutineCaches()
	return values[key]
}

// DelRoutineCache
func DelRoutineCache(key string) {
	values := GetRoutineCaches()
	delete(values, key)
}

// GetRoutineRequestContext
func GetRoutineRequestContext() *Context {
	raw, _ := GetGLSValue(GLSRequestContextKey)
	if !IsBlank(raw) {
		c := raw.(*Context)
		return c
	}
	return nil
}

// GetRoutineRequestID
func GetRoutineRequestID() string {
	raw, ok := GetGLSValue(GLSRequestIDKey)
	if ok {
		return raw.(string)
	}
	return ""
}

// IgnoreAuth
func IgnoreAuth(cancel ...bool) (success bool) {
	caches := GetRoutineCaches()
	if caches != nil {
		caches.IgnoreAuth(cancel...)
		success = true
	}
	return
}

func (app *Engine) initConfigs() {
	if C().Has("cors") {
		if C().GetBool("cors") {
			config := cors.DefaultConfig()
			config.AllowAllOrigins = true
			config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "api_key", "Authorization", TokenKey, ""}
			app.UseGin(cors.New(config))
		} else {
			var config cors.Config
			C().GetInterface("cors", &config)
			app.UseGin(cors.New(config))
		}
	}

	if C().Has("gzip") {
		if C().GetBool("gzip") {
			app.UseGin(gzip.Gzip(gzip.DefaultCompression))
		} else {
			v := C().GetInt("gzip")
			if v != 0 {
				app.UseGin(gzip.Gzip(v))
			}
		}
	}
}

func (app *Engine) initStatics() {
	statics := make(map[string]string)
	C().GetInterface("statics", &statics)
	if statics == nil || len(statics) == 0 {
		return
	}
	for key, val := range statics {
		stat, err := os.Lstat(val)
		if err != nil {
			ERROR("Static failed: %s", err.Error())
			continue
		}
		if stat.IsDir() {
			app.Static(key, val)
		} else {
			app.StaticFile(key, val)
		}
		AddWhitelist(regexp.MustCompile(fmt.Sprintf(`^GET\s%s`, key)))
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

func (app *Engine) init() {
	initDataSources()
	app.Use(func(c *Context) *STDReply {
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return nil
		}
		return nil
	})
	app.UseGin(session.New())
	app.initConfigs()
	app.initStatics()

	// Register default callbacks
	registerCallbacks()
}

// GetAppName
func GetAppName() string {
	name := C().GetString("name")
	if name == "" {
		PANIC("Application name is required")
	}
	return strings.ToLower(name)
}

// Release
func Release() {
	releaseDB()
	releaseCacheDB()
}
