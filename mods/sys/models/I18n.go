package models

// I18n 国际化语言
type I18n struct {
	ID     string               `json:"_id" displayName:"国际化语言"`
	Code   string               `name:"语言编码"`
	Name   string               `name:"语言名称"`
	Values map[string]I18nValue `name:"国际化值"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}

// I18nValue 国际化值
type I18nValue struct {
	Value string `name:"国际化值"`
	Group string `name:"所属组别"`
}
