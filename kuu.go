package kuu

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ROOT 应用运行目录
var ROOT string

func init() {
	env := os.Getenv("KUU_ENV") // KUU_ENV = 'dev' | 'test' | 'prod'
	if env == "" {
		env = "dev"
	} else if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	if path, err := filepath.Abs(os.Args[0]); err == nil {
		ROOT = filepath.Dir(path)
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
	Schemas map[string]*Schema
	methods InstMethods
}

// Model 模型注册
func (k *Kuu) Model(m interface{}) {
	s := reflect.TypeOf(m).Elem()
	schema := &Schema{
		Name:   s.Name(),
		Fields: make([]*SchemaField, s.NumField()),
	}
	for i := 0; i < s.NumField(); i++ {
		field := s.Field(i)
		tags := field.Tag
		sField := &SchemaField{}
		sField.Code = strings.ToLower(field.Name)
		sField.Name = tags.Get("name")
		sField.Default = tags.Get("default")
		sField.Alias = tags.Get("alias")
		if tags.Get("type") != "" {
			sField.Type = tags.Get("type")
		} else {
			sField.Type = field.Type.Name()
		}
		if tags.Get("required") != "" {
			required, err := strconv.ParseBool(tags.Get("required"))
			if err != nil {
				sField.Required = required
			} else {
				sField.Required = false
			}
		}
		schema.Fields[i] = sField
	}
	k.Schemas[schema.Name] = schema
	k.eachPlugins(func(p *Plugin) {
		// 触发OnModel
		if p.OnModel != nil {
			p.OnModel(k, schema)
		}
	})
}

func (k *Kuu) eachPlugins(cb func(*Plugin)) {
	for _, p := range plugins {
		cb(p)
	}
}

// loadConfigFile 加载配置文件（配置文件优先）
func (k *Kuu) loadConfigFile() {
	path := os.Getenv("KUU_CONFIG")
	if path == "" || !strings.HasSuffix(path, ".json") {
		path = "kuu.json"
	}
	var config H
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}
	if config != nil {
		for key, value := range config {
			k.Config[key] = value
		}
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
		// 触发OnLoad
		if p.OnLoad != nil {
			p.OnLoad(k)
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

// Run 重写启动函数
func (k *Kuu) Run(addr ...string) (err error) {
	k.eachPlugins(func(p *Plugin) {
		// 触发OnModel
		if p.BeforeRun != nil {
			p.BeforeRun(k)
		}
	})
	return k.Engine.Run(addr...)
}

// New 创建新应用
func New(cfg H) *Kuu {
	k := Kuu{
		Engine:  gin.New(),
		Schemas: make(map[string]*Schema),
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
	k.loadConfigFile()
	k.loadPlugins()
	FireHooks("OnNew", &k, cfg)
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
