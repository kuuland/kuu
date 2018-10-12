package kuu

// Schema 数据模型
type Schema struct {
	Name   string
	Fields []*SchemaField
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
