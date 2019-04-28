package models

// LoginOrg 组织登录信息
type LoginOrg struct {
	ID    string `json:"_id" displayName:"组织登录信息" noauth:"true"`
	Token string `name:"用户令牌"`
	UID   string `name:"用户ID"`
	Org   Org    `name:"登录组织" join:"Org<Code,Name>"`
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64       `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64       `name:"修改时间"`
}
