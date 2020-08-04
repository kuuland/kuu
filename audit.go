package kuu

type EventLog struct {
	Model     `rest:"*" displayName:"事件日志"`
	EventID   string `name:"事件ID（UUID）" gorm:"NOT NULL"`
	EventTime int64  `name:"事件时间" gorm:"NOT NULL"`
	SourceIP  string `name:"来源IP" gorm:"NOT NULL"`

	UserID           uint   `name:"操作人ID" gorm:"NOT NULL"`
	UserName         string `name:"操作人名称" gorm:"NOT NULL"`
	UserEmailAddress string `name:"操作人邮箱"`
	UserPhoneNumber  string `name:"操作人手机"`

	EventClass   string          `name:"事件分类" gorm:"NOT NULL" enum:"EventLogClass"`
	EventSubject string          `name:"事件主题" gorm:"NOT NULL"`
	EventSummary string          `name:"事件摘要" gorm:"NOT NULL"`
	EventLabels  []EventLogLabel `name:"关联事件标签"`
	EventData    string          `name:"事件详情(JSON-String)"`
}

func (log *EventLog) BindData(dst interface{}) error {
	return JSONParse(log.EventData, dst)
}

type EventLogLabel struct {
	Model      `rest:"*" displayName:"事件日志标签"`
	EventLogID uint      `name:"关联事件日志ID" gorm:"NOT NULL"`
	EventLog   *EventLog `name:"关联事件日志"`

	EventClass string `name:"事件分类" gorm:"NOT NULL;INDEX:event_log_label" enum:"EventLogClass"`
	LabelKey   string `name:"标签名" gorm:"NOT NULL;INDEX:event_log_label"`
	LabelValue string `name:"标签值" gorm:"NOT NULL;INDEX:event_log_label"`
}
