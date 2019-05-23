package kuu

import (
	"context"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Engine struct {
	*gin.Engine
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

// Import
func (e *Engine) Import(mods ...*Mod) {
	Import(e.Engine, mods...)
}

func connectedPrint(name, args string) {
	INFO("%-8s is connected: %s", name, args)
}

func onInit(e *Engine) {
	initDataSources()
	initRedis()

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
