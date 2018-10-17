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

var (
	ROOT       string
	plugins    = map[string]*Plugin{}
	contexts   = map[string]*Kuu{}
	kuuMethods = KuuMethods{}
)

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

// Kuu 应用
type Kuu struct {
	*gin.Engine
	Name       string
	Config     H
	Port       int16
	Env        string
	Schemas    map[string]*Schema
	appMethods AppMethods
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
	Emit("OnModel", k, schema)
}

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

func (k *Kuu) loadPlugins() {
	for _, p := range plugins {
		// 插件名不能为空
		if p.Name == "" {
			break
		}
		// 挂载中间件
		for _, value := range p.Middleware {
			if value == nil {
				break
			}
			k.Use(value)
		}
		// 挂载路由
		for _, value := range p.Routes {
			if value.Path == "" || value.Handler == nil {
				break
			}
			if value.Method == "" {
				value.Method = "GET"
			}
			k.Handle(value.Method, value.Path, value.Handler)
		}
		// 加载API
		for key, value := range p.AppMethods {
			if key == "" ||
				value == nil {
				break
			}
			k.appMethods[Join(p.Name, ":", key)] = value
		}
		Emit("OnPluginLoad", k)
	}
}

// Run 重写启动函数
func (k *Kuu) Run(addr ...string) (err error) {
	Emit("BeforeRun", k)
	return k.Engine.Run(addr...)
}

// New 创建新应用
func New(cfg H) *Kuu {
	k := Kuu{
		Engine:     gin.New(),
		Schemas:    make(map[string]*Schema),
		appMethods: make(AppMethods),
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
	Emit("OnNew", &k)
	return contexts[k.Name]
}

// Import 导入插件
func Import(ps ...*Plugin) {
	for _, p := range ps {
		Emit("OnImport", p)
		// 插件名不能为空
		if p.Name == "" {
			break
		}
		// 缓存插件
		plugins[p.Name] = p
		// 加载插件全局API
		for key, value := range p.KuuMethods {
			if key == "" || value == nil {
				break
			}
			kuuMethods[Join(p.Name, ":", key)] = value
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

// StdData 按标准格式返回数据
func StdData(data interface{}, msg string, code int) H {
	json := H{}
	if data != nil {
		json["data"] = msg
	}
	if msg != "" {
		json["msg"] = msg
	}
	json["code"] = code
	return json
}

// StdDataOK 返回数据
func StdDataOK(data interface{}) H {
	return StdData(data, "", 0)
}

// StdDataOKWithMsg 返回数据和提示信息
func StdDataOKWithMsg(data interface{}, msg string) H {
	return StdData(data, msg, 0)
}

// StdDataError 返回错误信息
func StdDataError(msg string) H {
	return StdData(nil, "", -1)
}

// StdDataErrorWithCode 返回错误信息和错误码
func StdDataErrorWithCode(msg string, code int) H {
	return StdData(nil, "", code)
}
