package kuu

import (
	"context"
	"fmt"
	"github.com/json-iterator/go"
	uuid "github.com/satori/go.uuid"
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
	GLSRequestIDKey      = "RequestID"
	// RunTime
	RunTime time.Time
	// IsProduction
	IsProduction = os.Getenv("GIN_MODE") == "release" || os.Getenv("KUU_ENV") == "prod"
	json         = jsoniter.ConfigCompatibleWithStandardLibrary
)

// M is a shortcut for map[string]interface{}
type M map[string]interface{}

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
func Default() (e *Engine) {
	e = &Engine{Engine: gin.Default()}
	e.Use(Recovery)
	if !C().DefaultGetBool("ignoreDefaultRootRoute", false) {
		e.GET("/", func(c *Context) {
			msg := c.L("kuu_up", "{{time}}", M{"time": RunTime.Format("2006-01-02 15:04:05")})
			c.STD(fmt.Sprintf("%s is up.", GetAppName()), msg)
		})
	}
	onInit(e)
	return
}

// New
func New() (e *Engine) {
	e = &Engine{Engine: gin.New()}
	onInit(e)
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
	DefaultCron.Start()
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
func (e *Engine) RegisterWhitelist(rules ...interface{}) {
	AddWhitelist(rules...)
}

// Run
func (e *Engine) Run(addr ...string) {
	RunTime = time.Now()
	address := resolveAddress(addr)
	srv := &http.Server{
		Addr:    address,
		Handler: e.Engine,
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
func (e *Engine) RunTLS(addr, certFile, keyFile string) {
	RunTime = time.Now()
	srv := &http.Server{
		Addr:    addr,
		Handler: e.Engine,
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

// ConvertKuuHandlers
var ConvertKuuHandlers = func(chain HandlersChain) (handlers gin.HandlersChain) {
	handlers = make(gin.HandlersChain, len(chain))
	for index, handler := range chain {
		handlers[index] = func(c *gin.Context) {
			kc := &Context{
				Context:       c,
				RoutineCaches: make(RoutineCaches),
				SignInfo:      GetSignContext(c),
			}
			if kc.SignInfo.IsValid() {
				desc := GetPrivilegesDesc(kc.SignInfo)
				kc.PrisDesc = desc
			}
			glsVals := make(gls.Values)
			glsVals[GLSSignInfoKey] = kc.SignInfo
			glsVals[GLSPrisDescKey] = kc.PrisDesc
			glsVals[GLSRoutineCachesKey] = kc.RoutineCaches
			glsVals[GLSRequestContextKey] = kc
			glsVals[GLSRequestIDKey] = uuid.NewV4().String()
			SetGLSValues(glsVals, func() {
				if InWhitelist(c) {
					IgnoreAuth()
				}
				handler(kc)
			})
		}
	}
	return
}

// Overrite r.Group
func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *gin.RouterGroup {
	return e.Engine.Group(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.Handle
func (e *Engine) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.Handle(httpMethod, relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.POST
func (e *Engine) POST(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.POST(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.GET
func (e *Engine) GET(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.GET(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.DELETE
func (e *Engine) DELETE(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.DELETE(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.PATCH
func (e *Engine) PATCH(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.PATCH(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.PUT
func (e *Engine) PUT(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.PUT(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.OPTIONS
func (e *Engine) OPTIONS(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.OPTIONS(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.HEAD
func (e *Engine) HEAD(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.HEAD(relativePath, ConvertKuuHandlers(handlers)...)
}

// Overrite r.Any
func (e *Engine) Any(relativePath string, handlers ...HandlerFunc) gin.IRoutes {
	return e.Engine.Any(relativePath, ConvertKuuHandlers(handlers)...)
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

func (e *Engine) initConfigs() {
	if C().Has("cors") {
		if C().GetBool("cors") {
			config := cors.DefaultConfig()
			config.AllowAllOrigins = true
			config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "api_key", "Authorization", TokenKey, ""}
			e.Use(cors.New(config))
		} else {
			var config cors.Config
			C().GetInterface("cors", &config)
			e.Use(cors.New(config))
		}
	}

	if C().Has("gzip") {
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

func (e *Engine) initStatics() {
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
			e.Static(key, val)
		} else {
			e.StaticFile(key, val)
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

func onInit(app *Engine) {
	app.Use(session.New())
	initDataSources()
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
