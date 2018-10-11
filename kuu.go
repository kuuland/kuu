package kuu

import (
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
)

// H 映射集合的别名
type H map[string]interface{}

// plugins 插件集合
var plugins = map[string]*Plugin{}

// contexts 应用缓存
var contexts = map[string]*Kuu{}

// methods 插件API
var methods = map[string]func(...interface{}) interface{}{}

// Kuu 应用
type Kuu struct {
	*gin.Engine
	Name    string
	Config  H
	Port    int16
	Env     string
	methods map[string]func(*Kuu, ...interface{}) interface{}
}

// AddM 模型注册
func (k *Kuu) AddM(m interface{}) {
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
			k.methods[key] = value
		}
		// 触发Onload
		if p.Onload != nil {
			p.Onload(k)
		}
	}
}

// D 调用实例上的插件API
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
		methods: make(map[string]func(*Kuu, ...interface{}) interface{}),
	}
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
	for i := 0; i < len(ps); i++ {
		p := ps[i]
		// 缓存插件
		plugins[p.Name] = p
		// 加载插件全局API
		for key, value := range p.Methods {
			if key == "" || value == nil {
				break
			}
			methods[key] = value
		}
	}
}

// Context 获取应用实例
func Context(name string) *Kuu {
	if name == "" {
		name = "kuu"
	}
	return contexts[name]
}

// Ctx 获取应用实例（快捷方式）
func Ctx() *Kuu {
	return Context("")
}

// D 调用插件API
func D(name string, args ...interface{}) interface{} {
	fn := methods[name]
	if fn == nil {
		return nil
	}
	return fn(args...)
}
