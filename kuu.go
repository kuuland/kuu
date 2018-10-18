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

// ROOT 应用运行目录
var ROOT string

var (
	contexts         = map[string]*Kuu{}
	pluginMiddleware = []gin.HandlerFunc{}
	pluginRoutes     = []RouteInfo{}
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
	Name    string
	Config  H
	Port    int16
	Env     string
	Schemas map[string]*Schema
}

// Model 模型注册
func (k *Kuu) Model(args ...interface{}) {
	for _, m := range args {
		model(k, m)
	}
}

func model(k *Kuu, m interface{}) {
	v := reflect.ValueOf(m)
	config := H{}
	if s, ok := m.(IConfig); ok {
		config = s.Config()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()
	schema := &Schema{
		Name:     t.Name(),
		FullName: Join(t.PkgPath(), "/", t.Name()),
		Fields:   make([]*SchemaField, t.NumField()),
	}
	if config["displayName"] != nil {
		schema.DisplayName = config["displayName"].(string)
	}
	if config["collection"] != nil {
		schema.Collection = config["collection"].(string)
	}
	if config["name"] != nil {
		schema.Name = config["name"].(string)
	}
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
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
	Emit("OnModel", k, schema, config)
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
	// 挂载中间件
	for _, m := range pluginMiddleware {
		k.Use(m)
	}
	// 挂载路由
	for _, r := range pluginRoutes {
		k.Handle(r.Method, r.Path, r.Handler)
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
		Engine:  gin.New(),
		Schemas: make(map[string]*Schema),
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
		for _, m := range p.Middleware {
			if m != nil {
				pluginMiddleware = append(pluginMiddleware, m)
			}
		}
		for _, r := range p.Routes {
			if r.Path == "" || r.Handler == nil {
				continue
			}
			if r.Method == "" {
				r.Method = "GET"
			}
			pluginRoutes = append(pluginRoutes, r)
		}
	}
}

// App 通过应用名获取应用实例
func App(name string) *Kuu {
	return contexts[name]
}

// K 快速导出单实例应用
func K() *Kuu {
	return App("kuu")
}

// Std 按标准格式返回数据
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

// StdOK 返回数据
func StdOK(data interface{}) H {
	return Std(data, "", 0)
}

// StdOKWithMsg 返回数据和提示信息
func StdOKWithMsg(data interface{}, msg string) H {
	return Std(data, msg, 0)
}

// StdError 返回错误信息
func StdError(msg string) H {
	return Std(nil, msg, -1)
}

// StdErrorWithCode 返回错误信息和错误码
func StdErrorWithCode(msg string, code int) H {
	return Std(nil, msg, code)
}
