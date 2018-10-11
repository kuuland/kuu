package plugins

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/kuuland/kuu"
)

// Mongo 数据源
type Mongo struct {
	Str     string
	Opts    map[string]interface{}
	Session *mgo.Session
}

// Connect 实现数据库连接
func (m *Mongo) Connect() *mgo.Session {
	session, err := mgo.DialWithTimeout(m.Str, 10*time.Second)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}
	log.Println("MongoDB is connected.")
	m.Session = session
	return session
}

// if cfg["mongo"] != nil {
// 	mongo := &Mongo{
// 		Str: cfg["mongo"].(string),
// 	}
// 	mongo.Connect()
// 	k.mongo = mongo
// }

// // Sess 获取数据库会话的拷贝实例
// func (k *Kuu) Sess() *mgo.Session {
// 	return k.mongo.Session.Clone()
// }

// // Model 获取数据库会话的拷贝实例
// func (k *Kuu) Model(sess *mgo.Session, name string) *mgo.Collection {
// 	return sess.DB("k11-central").C(name)
// }

// A 插件A
func A(opts gin.H) *kuu.Plugin {
	return &kuu.Plugin{
		Name:        "A",
		Routes:      routes(),
		Middleware:  middleware(),
		Methods:     methods(),
		InstMethods: instMethods(),
	}
}

// routes 插件路由
func routes() map[string]*kuu.Route {
	return kuu.R{
		"list": &kuu.Route{
			Path: "/list",
			Handler: func(c *gin.Context) {
				c.String(200, "插件A list")
			},
		},
	}
}

// middleware 插件中间件
func middleware() map[string]gin.HandlerFunc {
	return kuu.M{
		"ma": func(c *gin.Context) {
			t := time.Now()
			c.Set("example", "12345")
			c.Next()
			latency := time.Since(t)
			log.Print(latency, "alslfkj")
			status := c.Writer.Status()
			log.Println(status)
		},
	}
}

func methods() map[string]func(...interface{}) interface{} {
	return kuu.Method{
		"sessA": func(args ...interface{}) interface{} {
			for _, v := range args {
				log.Println("sessA", v)
			}
			return 555
		},
	}
}

func instMethods() map[string]func(*kuu.Kuu, ...interface{}) interface{} {
	return kuu.InstMethod{
		"sessB": func(k *kuu.Kuu, args ...interface{}) interface{} {
			val := args[0]
			log.Println("instMethod----dlksjfljfls", k.Name, val)
			return 666
		},
	}
}
