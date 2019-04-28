package models

// Message 系统消息
type Message struct {
	ID            string      `json:"_id" displayName:"系统消息"`
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
	// 标准字段
	CreatedBy interface{} `name:"创建人" join:"User<Username,Name>"`
	CreatedAt int64       `name:"创建时间"`
	UpdatedBy interface{} `name:"修改人" join:"User<Username,Name>"`
	UpdatedAt int64       `name:"修改时间"`
	IsDeleted bool        `name:"是否已删除"`
	Remark    string      `name:"备注"`
}
