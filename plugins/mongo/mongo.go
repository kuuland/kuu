package mongo

import (
	"log"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
	"github.com/kuuland/kuu/plugins/mongo/rest"
)

// defaultName 默认连接名
const defaultName = "default"

// connections 连接实例缓存
var connections = map[string]*Connection{}

var index = 0

// Connection 数据库连接
type Connection struct {
	Name    string
	URI     string
	UseDB   string
	session *mgo.Session
}

// Connect 数据库连接
func Connect(uri string) *mgo.Session {
	m := &Connection{
		URI: uri,
	}
	return New(m)
}

// New 创建数据库连接
func New(m *Connection) *mgo.Session {
	session, err := mgo.DialWithTimeout(m.URI, 10*time.Second)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	if m.UseDB == "" {
		m.UseDB = useDB(m.URI)
	}
	if m.Name == "" {
		m.Name = defaultName
	}
	m.session = session
	connections[m.Name] = m

	log.Println(kuu.Join("MongoDB '", m.UseDB, "' is connected."))
	return session
}

// useDB 从URI中截取数据库名
func useDB(uri string) string {
	s := strings.LastIndex(uri, "/") + 1
	e := strings.Index(uri, "?")
	if e == -1 {
		e = len(uri)
	}
	db := uri[s:e]
	return db
}

// SN 根据连接名获取会话
func SN(name string) *mgo.Session {
	if m := connections[name]; m != nil {
		return m.session.Clone()
	}
	return nil
}

// S 获取会话
func S() *mgo.Session {
	return SN(defaultName)
}

// C 获取集合对象
func C(name string) *mgo.Collection {
	if m := connections[defaultName]; m != nil {
		if s := m.session.Clone(); s != nil {
			return s.DB(m.UseDB).C(name)
		}
	}
	return nil
}

// Plugin 插件声明
var Plugin = &kuu.Plugin{
	Name: "mongo",
	OnLoad: func(k *kuu.Kuu) {
		if c := k.Config["mongo"]; c != nil {
			uri := c.(string)
			Connect(uri)
		}
	},
	OnModel: func(k *kuu.Kuu, schema *kuu.Schema) {
		MountRESTful(k, schema.Name)
	},
	Methods: kuu.Methods{
		"Connect": func(args ...interface{}) interface{} {
			uri := args[0].(string)
			return Connect(uri)
		},
		"New": func(args ...interface{}) interface{} {
			m := args[0].(*Connection)
			return New(m)
		},
		"C": func(args ...interface{}) interface{} {
			name := args[0].(string)
			return C(name)
		},
		"S": func(args ...interface{}) interface{} {
			return S()
		},
		"SN": func(args ...interface{}) interface{} {
			name := args[0].(string)
			return SN(name)
		},
	},
	InstMethods: kuu.InstMethods{
		"rest": func(k *kuu.Kuu, args ...interface{}) interface{} {
			if args != nil {
				for _, item := range args {
					name := item.(string)
					MountRESTful(k, name)
				}
			}
			return nil
		},
	},
}

// MountRESTful 挂载模型RESTful接口
func MountRESTful(k *kuu.Kuu, name string) {
	path := kuu.Join("/", strings.ToLower(name))
	k.POST(path, rest.Create(name))
	k.DELETE(path, rest.Remove(name))
	k.PUT(path, rest.Update(name))
	k.GET(path, rest.List(name))
	k.GET(kuu.Join(path, "/:id"), rest.ID(name))
}
