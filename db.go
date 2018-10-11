package kuu

// IDB 数据库接口
type IDB interface {
	Connect(info interface{}) interface{}
}
