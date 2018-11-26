package kuu

import "fmt"

// IModel 定义了模型持久化操作接口
type IModel interface {
	New(*Schema) IModel
	Schema() *Schema
	Create(interface{}) ([]interface{}, error)
	Remove(interface{}) error
	RemoveEntity(interface{}) error
	RemoveAll(interface{}) (interface{}, error)
	RemoveWithData(interface{}, interface{}) error
	RemoveEntityWithData(interface{}, interface{}) error
	RemoveAllWithData(interface{}, interface{}) (interface{}, error)
	PhyRemove(interface{}) error
	PhyRemoveEntity(interface{}) error
	PhyRemoveAll(interface{}) (interface{}, error)
	Update(interface{}, interface{}) error
	UpdateEntity(interface{}) error
	UpdateAll(interface{}, interface{}) (interface{}, error)
	List(interface{}, interface{}) (H, error)
	ID(interface{}, interface{}) error
	One(interface{}, interface{}) error
}

// ModelAdapter 全局模型适配器实例
var ModelAdapter IModel

// Model 获取操作实例
func Model(name string) IModel {
	s := GetSchema(name)
	if s == nil {
		panic(fmt.Sprintf("Model '%s' not registered", name))
	}
	return s.Adapter.New(s)
}

// Op 获取操作实例（实时解析）
func Op(m interface{}) IModel {
	schema, _ := parseSchema(m)
	if schema != nil {
		return Model(schema.Name)
	}
	return nil
}
