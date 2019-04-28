package models

// Audit 审计日志
type Audit struct {
	ID         string      `json:"_id" displayName:"审计日志"`
	Type       string      `name:"审计类型" dict:"sys_audit_type"`
	DataID     string      `name:"关键数据ID"`
	DataDetail interface{} `name:"关键数据详情"`
	Desc       string      `name:"描述" remark:"系统按指定规则生成的一段可读的描述"`
	Content    string      `name:"内容" remark:"用户可能填写的内容"`
	Attachs    []File      `name:"附件" remark:"用户可能上传的附件"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64       `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64       `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}
