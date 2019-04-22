// Package kuu is a pluggable Go web framework.
package kuu

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// ROOT 应用运行根目录
	ROOT string
	// ENV 应用运行环境
	ENV string
	// Schemas 数据模型集合
	Schemas = map[string]*Schema{}
	// Data 全局数据缓存集
	Data = H{}
)

var (
	apps          = []*Kuu{}
	modMiddleware = []gin.HandlerFunc{}
	modRoutes     = []RouteInfo{}
	modModels     = []interface{}{}
)

func init() {
	gin.SetMode(gin.ReleaseMode)
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
	Config      H
	Name        string
	RoutePrefix string
}

// GetSchema 获取模型信息
func GetSchema(name string) *Schema {
	return Schemas[name]
}

// RegisterModel 模型注册
func RegisterModel(args ...interface{}) {
	for _, m := range args {
		schema := parseSchema(m)
		Schemas[schema.Name] = schema
		Emit("OnModel", schema, schema.Config)
	}
}

//CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func parseSchema(m interface{}) *Schema {
	v := reflect.ValueOf(m)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		panic(`Model only accepts pointer. Example:
		Use: kuu.RegisterModel(&Struct{}) instead of kuu.RegisterModel(Struct{})
	`)
	}
	// 判断是否实现了配置接口
	config := H{}
	if s, ok := m.(IConfig); ok {
		config = s.Config()
	}
	t := v.Type()
	fields := make([]SchemaField, 0)
	schema := &Schema{
		Name:     t.Name(),
		FullName: Join(t.PkgPath(), "/", t.Name()),
		Fields:   &fields,
		Origin:   m,
		Config:   config,
	}
	// 解析字段列表
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tags := field.Tag
		sField := SchemaField{
			Tags:    make(map[string]string),
			IsArray: false,
		}
		parseModelTags(&sField, tags)
		sField.Code = field.Name
		if sField.Tags["name"] != "" {
			sField.Name = tags.Get("name")
		} else {
			sField.Name = sField.Code
		}
		if sField.Tags["join"] != "" {
			sField.JoinName, sField.JoinSelect = parseJoinSelect(tags.Get("join"))
		}
		if _, ok := sField.Tags["array"]; ok {
			sField.IsArray = true
		}
		sField.Default = tags.Get("default")
		if sField.Tags["type"] != "" {
			sField.Type = tags.Get("type")
		} else {
			sField.Type = field.Type.Name()
		}
		if _, ok := sField.Tags["required"]; ok {
			required, err := strconv.ParseBool(tags.Get("required"))
			if err != nil {
				sField.Required = required
			} else {
				sField.Required = false
			}
		}
		fields = append(fields, sField)
		if sField.Code == "ID" {
			if s, ok := sField.Tags["name"]; ok {
				config["name"] = s
			}
			if s, ok := sField.Tags["displayName"]; ok {
				config["displayName"] = s
			}
			if s, ok := sField.Tags["collection"]; ok {
				config["collection"] = s
			}
			if s, ok := sField.Tags["noauth"]; ok {
				config["noauth"] = s
			}
		}
	}
	// 模型配置
	if config["name"] != nil {
		schema.Name = config["name"].(string)
	}
	if config["noauth"] != nil {
		str := config["noauth"].(string)
		if v, err := strconv.ParseBool(str); err == nil {
			schema.NoAuth = v
		}
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
	if config["adapter"] != nil {
		if s, ok := config["adapter"].(IModel); ok {
			schema.Adapter = s
		}
	}
	if schema.Adapter == nil && ModelAdapter != nil {
		schema.Adapter = ModelAdapter
	}
	if schema.Adapter == nil {
		panic("Please register 'kuu.ModelAdapter' before using 'kuu.RegisterModel'")
	}
	return schema
}

func parseJoinSelect(join string) (name string, s map[string]int) {
	if join == "" {
		return
	}
	name = join
	start := strings.Index(join, "<")
	end := strings.LastIndex(join, ">")
	if start == -1 || end == -1 {
		return
	}
	if raw := join[start+1 : end]; raw != "" {
		name = join[0:start]
		s = make(map[string]int)
		fields := strings.Split(raw, ",")
		for _, item := range fields {
			v := 1
			if strings.HasPrefix(item, "-") {
				v = 0
				item = item[1:len(item)]
			}
			s[item] = v
		}
	}
	return
}

func parseModelTags(field *SchemaField, tag reflect.StructTag) {
	split := strings.Split(string(tag), " ")
	for _, item := range split {
		if item == "" {
			continue
		}
		t := strings.Split(item, ":")
		if t != nil && len(t) > 0 && t[0] != "" {
			var value string
			if len(t) > 1 {
				v, err := strconv.Unquote(t[1])
				if err == nil {
					value = v
				} else {
					value = t[1]
				}
			}
			field.Tags[t[0]] = value
		}
	}
}

// ParseKuuJSON 加载本地配置文件中的配置信息
func ParseKuuJSON() (H, error) {
	path := os.Getenv("KUU_CONFIG")
	if path == "" || !strings.HasSuffix(path, ".json") {
		path = "kuu.json"
	}
	config := H{}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config, nil
	}
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
	config, err := ParseKuuJSON()
	if err != nil {
		Error(err)
		return
	}
	if config != nil {
		for key, value := range config {
			k.Config[key] = value
		}
	}
	if k.Config["routePrefix"] != nil {
		k.RoutePrefix = k.Config["routePrefix"].(string)
	}
}

func (k *Kuu) loadMods() {
	// 挂载模块中间件
	for _, m := range modMiddleware {
		k.Use(m)
	}
	// 挂载模块路由
	for _, r := range modRoutes {
		k.Handle(r.Method, Join(k.RoutePrefix, r.Path), r.Handler)
	}
	// 挂载模块模型
	for _, m := range modModels {
		RegisterModel(m)
	}
}

// Run 重写启动函数
func (k *Kuu) Run(addr ...string) (err error) {
	Emit("BeforeRun", k)
	return k.Engine.Run(addr...)
}

// New 根据配置创建并返回一个新的应用实例，创建过程会自动加载已导入模块
func New(cfg ...H) *Kuu {
	config := resolveConfig(cfg)
	k := Kuu{
		Engine: gin.New(),
	}
	k.Use(gin.Logger(), gin.Recovery())
	if config == nil {
		config = H{}
	}
	if config["name"] != nil {
		k.Name = config["name"].(string)
	}
	apps = append(apps, &k)
	k.Config = config
	k.loadConfigFile()
	if _, has := k.Config["cors"]; has {
		k.Warn("Mounted CORS middleware.")
		k.Use(CORSMiddleware())
	}
	k.loadMods()
	Emit("OnNew", &k)
	return &k
}

// Import 导入一个或多个模块
func Import(ps ...*Mod) {
	for _, p := range ps {
		Emit("OnImport", p)
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

func resolveConfig(config []H) H {
	switch len(config) {
	case 0:
		return H{}
	case 1:
		return config[0]
	default:
		return H{}
	}
}
