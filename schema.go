package kuu

import (
	"time"
)

// IConfig 模型配置接口，数据模型通过实现该接口来增强模型配置
type IConfig interface {
	Config() H
}

// Schema 数据模型结构
type Schema struct {
	Name        string
	DisplayName string
	FullName    string
	Collection  string
	Fields      []*SchemaField
	Origin      interface{}
	Config      H
}

// SchemaField 数据模型字段
type SchemaField struct {
	Code     string
	Name     string
	Type     string
	Required bool
	Default  interface{}
}

// Model 定义了数据模型的一些基本字段
type Model struct {
	CreatedBy interface{} `name:"创建人"`
	CreatedAt time.Time   `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人"`
	UpdatedAt time.Time   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
}

// IModel 定义了模型持久化操作接口
type IModel interface {
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
