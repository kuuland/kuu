package kuu

import "time"

type Model struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	OrgID       uint
	CreatedByID uint
	UpdatedByID uint
	DeletedByID uint
	Org         Org  `gorm:"foreignkey:OrgID"`
	CreatedBy   User `gorm:"foreignkey:CreatedByID"`
	UpdatedBy   User `gorm:"foreignkey:UpdatedByID"`
	DeletedBy   User `gorm:"foreignkey:DeletedByID"`
	Remark      string
}

// User
type User struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	OrgID       uint
	CreatedByID uint
	UpdatedByID uint
	DeletedByID uint
	Remark      string

	Username    string       `name:"账号"`
	Password    string       `name:"密码"`
	Name        string       `name:"姓名"`
	Avatar      string       `name:"头像"`
	Sex         int          `name:"性别"`
	Mobile      string       `name:"手机号"`
	Email       string       `name:"邮箱"`
	Language    string       `name:"语言"`
	Disable     bool         `name:"是否禁用"`
	RoleAssigns []RoleAssign `name:"角色分配"`
	IsBuiltIn   bool         `name:"是否系统内置"`
}

// Org
type Org struct {
	ID          uint `gorm:"primary_key"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time `sql:"index"`
	OrgID       uint
	CreatedByID uint
	UpdatedByID uint
	DeletedByID uint
	CreatedBy   User `gorm:"foreignkey:CreatedByID"`
	UpdatedBy   User `gorm:"foreignkey:UpdatedByID"`
	DeletedBy   User `gorm:"foreignkey:DeletedByID"`
	Remark      string

	Code         string `name:"组织编码"`
	Name         string `name:"组织名称"`
	Pid          uint   `name:"父组织ID"`
	Sort         int    `name:"排序号"`
	FullPathPid  string `name:"完整层级路径"`
	FullPathName string `name:"完整层级路径"`
}

func (u *User) BeforeSave() {
	if len(u.Password) == 32 {
		u.Password = GenerateFromPassword(u.Password)
	}
	return
}

type RoleAssign struct {
	Model
	RoleID     string `name:"角色ID"`
	ExpiryUnix int64  `name:"过期时间"`
}

// Role
type Role struct {
	Model
	Code                string                `name:"角色编码"`
	Name                string                `name:"角色名称"`
	OperationPrivileges []OperationPrivileges `name:"操作权限"`
	DataPrivileges      []DataPrivileges      `name:"数据权限"`
	IsBuiltIn           bool                  `name:"是否系统内置"`
}

// OperationPrivileges
type OperationPrivileges struct {
	Model
	Permission string `name:"权限编码"`
	Desc       string `name:"权限描述"`
}

// DataPrivileges
type DataPrivileges struct {
	Model
	OrgID            uint   `name:"组织ID"`
	OrgName          string `name:"组织名称"`
	AllReadableRange string `name:"全局可读范围" dict:"sys_data_range"`
	AllWritableRange string `name:"全局可写范围" dict:"sys_data_range"`
	AuthObjects      []AuthObject
}

// AuthObject
type AuthObject struct {
	Model
	Name             string `name:"实体名称"`
	DisplayName      string `name:"实体显示名"`
	ObjReadableRange string `name:"实体可读范围" dict:"sys_data_range"`
	ObjWritableRange string `name:"实体可写范围" dict:"sys_data_range"`
}

// Menu
type Menu struct {
	Model
	Code          string
	Name          string
	URI           string
	Icon          string
	Pid           string
	Group         string
	Disable       bool
	IsLink        bool
	IsVirtual     bool
	Sort          int
	IsBuiltIn     bool
	IsDefaultOpen bool
	Closeable     bool
	Type          string
}

// Audit
type Audit struct {
	Model
	Type       string      `name:"审计类型" dict:"sys_audit_type"`
	DataID     string      `name:"关键数据ID"`
	DataDetail interface{} `name:"关键数据详情"`
	Desc       string      `name:"描述" remark:"系统按指定规则生成的一段可读的描述"`
	Content    string      `name:"内容" remark:"用户可能填写的内容"`
	Attachs    []File      `name:"附件" remark:"用户可能上传的附件"`
}

// AuthRule
type AuthRule struct {
	Model
	UID               string      `name:"用户ID"`
	Username          string      `name:"用户账号"`
	Name              string      `name:"用户姓名"`
	OrgID             string      `name:"登录组织"`
	OrgName           string      `name:"登录组织"`
	ObjectName        string      `name:"访问实体名称"`
	ObjectDisplayName string      `name:"访问实体显示名称"`
	ReadableScope     string      `name:"可读范围" dict:"sys_data_range"`
	WritableScope     string      `name:"可写范围" dict:"sys_data_range"`
	ReadableOrgIDs    []string    `name:"可读数据的组织ID"`
	WritableOrgIDs    []string    `name:"可写数据的组织ID"`
	HitAssign         interface{} `name:"命中规则"`
	Permissions       []string    `name:"可访问操作权限"`
}

// Dict
type Dict struct {
	Model
	Code      string      `name:"字典编码"`
	Name      string      `name:"字典名称"`
	Values    []DictValue `name:"字典值"`
	IsBuiltIn bool        `name:"是否系统内置"`
}

// DictValue
type DictValue struct {
	Model
	Label string `name:"字典标签"`
	Value string `name:"字典键值"`
	Sort  int    `name:"排序号"`
}

// File
type File struct {
	Model
	UID    string `json:"uid" name:"文件ID"`
	Type   string `json:"type" name:"文件类型"`
	Size   int64  `json:"size" name:"文件大小"`
	Name   string `json:"name" name:"文件名称"`
	Status string `json:"status" name:"文件状态"`
	URL    string `json:"url" name:"文件路径"`
	Path   string `json:"path" name:"本地路径"`
}

// LoginOrg
type LoginOrg struct {
	Model
	Token string `name:"用户令牌"`
	UID   string `name:"用户ID"`
}

//func (o *LoginOrg) IsValid() bool {
//	if o.Org != nil {
//		return true
//	}
//	return false
//}

// Message
type Message struct {
	Model
	Type          string      `name:"消息类型" dict:"sys_message_type"`
	Title         string      `name:"消息标题"`
	Content       string      `name:"消息内容"`
	Attachs       []File      `name:"消息附件"`
	BusType       string      `name:"业务类型" dict:"sys_message_bustype"`
	BusID         string      `name:"业务数据ID"`
	BusDetail     string      `name:"业务数据详情"`
	TryTimes      int32       `name:"重试次数"`
	Pusher        interface{} `name:"发送人" join:"User<Username,Name>"`
	PushTime      int64       `name:"推送时间"`
	PushStatus    string      `name:"推送状态" dict:"sys_push_status" remark:"待推送、推送中、重试中、已推送、已终止"`
	Receiver      interface{} `name:"接收人" join:"User<Username,Name>"`
	ReadingStatus string      `name:"阅读状态" dict:"sys_read_status" remark:"未读、已读"`
	ReadingTime   int64       `name:"阅读时间"`
}

// Param
type Param struct {
	Model
	Code      string `name:"参数编码"`
	Name      string `name:"参数名称"`
	Value     string `name:"参数值"`
	IsBuiltIn bool   `name:"是否系统内置"`
}
