package db

import (
	"log"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

const defaultName = "default"

var connections = map[string]*Connection{}

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
		m.UseDB = parseDB(m.URI)
	}
	if m.Name == "" {
		m.Name = defaultName
	}
	m.session = session
	connections[m.Name] = m

	log.Println(kuu.Join("MongoDB '", m.UseDB, "' is connected."))
	return session
}

// parseDB 从URI中截取数据库名
func parseDB(uri string) string {
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
