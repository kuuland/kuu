package kuu

import (
	"bytes"
	"log"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
)

func init() {
	env := os.Getenv("KUU_ENV") // KUU_ENV = 'dev' | 'test' | 'prod'
	if env == "" {
		env = "dev"
	} else if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// H 映射集合的别名
type H map[string]interface{}

// plugins 插件集合
var plugins = map[string]*Plugin{}

// contexts 应用缓存
var contexts = map[string]*Kuu{}

// methods 插件API
var methods = Methods{}

// Kuu 应用
type Kuu struct {
	*gin.Engine
	Name    string
	Config  H
	Port    int16
	Env     string
	methods InstMethods
}

// Model 模型注册
func (k *Kuu) Model(m interface{}) {
	s := reflect.TypeOf(m).Elem()
	log.Println(s.Name())
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tags := field.Tag
		log.Println(field.Name, tags)
	}
}

// loadPlugins 加载插件
func (k *Kuu) loadPlugins() {
	for _, p := range plugins {
		// 插件名不能为空
		if p.Name == "" {
			break
		}
		// 挂载中间件
		for key, value := range p.Middleware {
			if key == "" ||
				value == nil {
				break
			}
			k.Use(value)
		}
		// 挂载路由
		for key, value := range p.Routes {
			if key == "" ||
				value.Path == "" ||
				value.Handler == nil {
				break
			}
			if value.Method == "" {
				value.Method = "GET"
			}
			k.Handle(value.Method, value.Path, value.Handler)
		}
		// 加载API
		for key, value := range p.InstMethods {
			if key == "" ||
				value == nil {
				break
			}
			k.methods[Join(p.Name, ":", key)] = value
		}
		// 触发Onload
		if p.Onload != nil {
			p.Onload(k)
		}
	}
}

// Join 基于字节实现的字符串拼接
func Join(args ...string) string {
	b := bytes.Buffer{}
	for _, item := range args {
		b.WriteString(item)
	}
	return b.String()
}

// D 调用插件实例API
func (k *Kuu) D(name string, args ...interface{}) interface{} {
	fn := k.methods[name]
	if fn == nil {
		return nil
	}
	return fn(k, args...)
}

// New 创建新应用
func New(cfg H) *Kuu {
	k := Kuu{
		Engine:  gin.New(),
		methods: make(InstMethods),
	}
	k.Use(gin.Logger(), gin.Recovery())
	if cfg == nil {
		cfg = H{}
	}
	if cfg["name"] == nil {
		k.Name = "kuu"
	} else {
		k.Name = cfg["name"].(string)
	}
	contexts[k.Name] = &k

	k.Config = cfg
	k.loadPlugins()
	return contexts[k.Name]
}

// Import 导入插件
func Import(ps ...*Plugin) {
	for _, p := range ps {
		// 插件名不能为空
		if p.Name == "" {
			break
		}
		// 缓存插件
		plugins[p.Name] = p
		// 加载插件全局API
		for key, value := range p.Methods {
			if key == "" || value == nil {
				break
			}
			methods[Join(p.Name, ":", key)] = value
		}
	}
}

// App 通过应用名获取应用实例
func App(name string) *Kuu {
	return contexts[name]
}

// K 获取应用实例（获取不指定Name所创建的应用）
func K() *Kuu {
	return App("kuu")
}

// D 调用插件API
func D(name string, args ...interface{}) interface{} {
	fn := methods[name]
	if fn == nil {
		return nil
	}
	return fn(args...)
}
