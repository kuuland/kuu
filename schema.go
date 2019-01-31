package kuu

// IConfig 模型配置接口，数据模型通过实现该接口来增强模型配置
type IConfig interface {
	Config() H
}

// Base 定义了数据模型的一些基本字段
type Base struct {
	CreatedBy interface{} `name:"创建人"`
	CreatedAt int64   `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人"`
	UpdatedAt int64   `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}

// Schema 数据模型结构
type Schema struct {
	Name        string
	DisplayName string
	FullName    string
	Collection  string
	Fields      []*SchemaField
	Origin      interface{} `json:"-"`
	Config      H
	Adapter     IModel `json:"-"`
}

// SchemaField 数据模型字段
type SchemaField struct {
	Code       string
	Name       string
	Type       string
	Required   bool
	Default    interface{}
	Tags       map[string]string
	IsArray    bool
	JoinName   string
	JoinSelect map[string]int
}
