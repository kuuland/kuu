package kuu

// IConfig 配置接口
type IConfig interface {
	Config() H
}

// Schema 数据模型
type Schema struct {
	Name        string
	DisplayName string
	FullName    string
	Collection  string
	Fields      []*SchemaField
}

// SchemaField 模型字段
type SchemaField struct {
	Code     string
	Name     string
	Type     string
	Required bool
	Default  interface{}
	Alias    string
}
