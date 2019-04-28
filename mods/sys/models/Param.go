package models

// Param 系统参数
type Param struct {
	ID        string `json:"_id" displayName:"系统参数"`
	Code      string `name:"参数编码"`
	Name      string `name:"参数名称"`
	Value     string `name:"参数值"`
	IsBuiltIn bool   `name:"是否系统内置"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}
