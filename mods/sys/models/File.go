package models

// File 系统文件
type File struct {
	ID     string `json:"_id" displayName:"系统文件"`
	UID    string `json:"uid" name:"文件ID"`
	Type   string `json:"type" name:"文件类型"`
	Size   int64  `json:"size" name:"文件大小"`
	Name   string `json:"name" name:"文件名称"`
	Status string `json:"status" name:"文件状态"`
	URL    string `json:"url" name:"文件路径"`
	Path   string `json:"path" name:"本地路径"`
	// 标准字段
	CreatedBy User   `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64  `name:"创建时间"`
	UpdatedBy User   `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64  `name:"修改时间"`
	IsDeleted bool   `name:"是否已删除"`
	Remark    string `name:"备注"`
}
