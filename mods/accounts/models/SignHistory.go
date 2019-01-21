package models

import "time"

// SignHistory 用户登入/登出历史
type SignHistory struct {
	ID        string `json:"_id" displayName:"用户登入/登出历史"`
	ReqData   string `name:"请求数据"`
	LoginData string `name:"登录数据"`
	Token     string `name:"令牌信息"`
	Method    string `name:"登入/登出"`
	// 标准字段
	CreatedAt time.Time `name:"创建时间"`
	UpdatedAt time.Time `name:"修改时间"`
	IsDeleted bool      `name:"是否已删除"`
	Remark    string    `name:"备注"`
}
