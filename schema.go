package kuu

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
}

// SchemaField 数据模型字段
type SchemaField struct {
	Code     string
	Name     string
	Type     string
	Required bool
	Default  interface{}
	Alias    string
}
