// Package kuu is a pluggable Go web framework.
package kuu

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ROOT 应用运行根目录
var ROOT string

// ENV 应用运行环境
var ENV string

var (
	apps          = []*Kuu{}
	modMiddleware = []gin.HandlerFunc{}
	modRoutes     = []RouteInfo{}
	modModels     = []interface{}{}
)

func init() {
	env := os.Getenv("KUU_ENV") // KUU_ENV = 'dev' | 'test' | 'prod'
	if env == "" {
		env = "dev"
	} else if env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	ENV = env

	if path, err := filepath.Abs(os.Args[0]); err == nil {
		ROOT = filepath.Dir(path)
	}
}

// H 是 map[string]interface{} 的快捷方式
type H map[string]interface{}

// Kuu 是框架的实例，它包含系统配置、数据模型等信息
type Kuu struct {
	*gin.Engine
	Config  H
	Name    string
	Schemas map[string]*Schema
}

// Model 模型注册
func (k *Kuu) Model(args ...interface{}) {
	for _, m := range args {
		v := reflect.ValueOf(m)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		} else {
			panic(`Model only accepts pointer. Example:
			Use: kuu.Model(&Struct{}) instead of kuu.Model(Struct{})
		`)
			}
		}
		// 判断是否实现了配置接口
		config := H{}
		if s, ok := m.(IConfig); ok {
			config = s.Config()
		}
		t := v.Type()
		schema := &Schema{
			Name:     t.Name(),
			FullName: Join(t.PkgPath(), "/", t.Name()),
			Fields:   make([]*SchemaField, t.NumField()),
			Origin:   m,
		}
		if config["name"] != nil {
			schema.Name = config["name"].(string)
		}
		if config["displayName"] != nil {
			schema.DisplayName = config["displayName"].(string)
		} else {
			schema.DisplayName = schema.Name
		}
		if config["collection"] != nil {
			schema.Collection = config["collection"].(string)
		} else {
			schema.Collection = schema.Name
		}
		schema.Config = config
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			tags := field.Tag
			sField := &SchemaField{}
			sField.Code = strings.ToLower(field.Name)
			if tags.Get("name") != "" {
				sField.Name = tags.Get("name")
			} else {
				sField.Name = sField.Code
			}
			if tags.Get("alias") != "" {
				sField.Name = tags.Get("alias")
			}
			sField.Default = tags.Get("default")
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
		Emit("OnModel", k, schema, config)
	}
}

// ParseLocalConfig 加载本地配置文件中的配置信息
func ParseLocalConfig() (H, error) {
	path := os.Getenv("KUU_CONFIG")
	if path == "" || !strings.HasSuffix(path, ".json") {
		path = "kuu.json"
	}
	config := H{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &config); err != nil {
		Error(err)
		return nil, err
	}
	return config, nil
}

func (k *Kuu) loadConfigFile() {
	config, err := ParseLocalConfig()
	if err != nil {
		Error(err)
		return
	}
	if config != nil {
		for key, value := range config {
			k.Config[key] = value
		}
	}
}

func (k *Kuu) loadMods() {
	// 挂载模块中间件
	for _, m := range modMiddleware {
		k.Use(m)
	}
	// 挂载模块路由
	for _, r := range modRoutes {
		k.Handle(r.Method, r.Path, r.Handler)
	}
	// 挂载模块模型
	for _, m := range modModels {
		k.Model(m)
	}
}

// Run 重写启动函数
func (k *Kuu) Run(addr ...string) (err error) {
	Emit("BeforeRun", k)
	return k.Engine.Run(addr...)
}

// New 根据配置创建并返回一个新的应用实例，创建过程会自动加载已导入模块
func New(cfg H) *Kuu {
	k := Kuu{
		Engine:  gin.New(),
		Schemas: make(map[string]*Schema),
	}
	k.Use(gin.Logger(), gin.Recovery())
	if cfg == nil {
		cfg = H{}
	}
	if cfg["name"] != nil {
		k.Name = cfg["name"].(string)
	}
	apps = append(apps, &k)
	k.Config = cfg
	k.loadConfigFile()
	k.loadMods()
	Emit("OnNew", &k)
	return &k
}

// Import 导入一个或多个模块
func Import(ps ...*Mod) {
	for _, p := range ps {
		for _, m := range p.Middleware {
			if m != nil {
				modMiddleware = append(modMiddleware, m)
			}
		}
		for _, r := range p.Routes {
			if r.Path == "" || r.Handler == nil {
				continue
			}
			if r.Method == "" {
				r.Method = "GET"
			}
			modRoutes = append(modRoutes, r)
		}
		for _, m := range p.Models {
			if m == nil {
				continue
			}
			modModels = append(modModels, m)
		}
	}
}

// App 根据应用名获取框架缓存实例
func App(name string) *Kuu {
	for _, app := range apps {
		if app.Name == name {
			return app
		}
	}
	return nil
}

// K 快捷访问首个缓存实例
func K() *Kuu {
	return apps[0]
}

// Std 将传入参数包装成Kuu响应格式的数据
/*
	Kuu响应格式约定如下：
	{
		"data": "",  表正常或异常情况下的数据信息，类型由具体接口决定，非必填
		"msg": "",   表正常或异常情况下的提示信息，字符串，非必填
		"code": 0    调用是否成功的明确标记，0表成功，非0表失败（默认异常值为-1），整型，必填
	}
*/
func Std(data interface{}, msg string, code int) H {
	json := H{}
	if data != nil {
		json["data"] = data
	}
	if msg != "" {
		json["msg"] = msg
	}
	json["code"] = code
	return json
}

// StdOK 只包含数据部分的Std
func StdOK(data interface{}) H {
	return Std(data, "", 0)
}

// StdOKWithMsg 只包含数据和提示部分的Std
func StdOKWithMsg(data interface{}, msg string) H {
	return Std(data, msg, 0)
}

// StdError 只包含错误信息的Std
func StdError(msg string) H {
	return Std(nil, msg, -1)
}

// StdErrorWithCode 只包含错误信息和错误码的Std
func StdErrorWithCode(msg string, code int) H {
	return Std(nil, msg, code)
}
