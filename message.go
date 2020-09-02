package kuu

import (
	"fmt"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

type Message struct {
	Model `rest:"*" displayName:"系统消息"`

	Subject     string `name:"消息标题"`
	Content     string `name:"消息内容"`
	Attachments []File `name:"消息附件" gorm:"polymorphic:Owner;polymorphic_value:Message.Attachments"`

	SenderID       uint      `name:"发送人ID"`
	SenderUsername string    `name:"发送人账号"`
	Sender         *User     `name:"发送人" gorm:"foreignkey:SenderID"`
	SenderSourceIP string    `name:"发送人IP地址"`
	SentAt         time.Time `name:"发送时间"`

	Range    MessageRange     `name:"消息接收范围"`
	Receipts []MessageReceipt `name:"消息接收/阅读回执"`
}

// BeforeCreate
func (m *Message) BeforeCreate() {
	m.SentAt = time.Now()
	if c := GetRoutineRequestContext(); c != nil {
		m.SenderSourceIP = c.ClientIP()
		m.SenderID = c.SignInfo.UID
		m.SenderUsername = c.SignInfo.Username
	}
}

type MessageRange struct {
	Model     `rest:"*" displayName:"消息通知范围"`
	MessageID uint     `name:"关联消息ID" gorm:"not null"`
	Message   *Message `name:"关联消息"`

	OrgIDs    string `name:"关联接收组织ID（多个以英文逗号分隔）"`
	UserIDs   string `name:"关联接收人ID（多个以英文逗号分隔）"`
	RoleCodes string `name:"关联接收角色编码（多个以英文逗号分隔）"`
}

// BeforeCreate
func (m *MessageRange) BeforeCreate() {
	m.UserIDs = strings.TrimSpace(m.UserIDs)
	m.RoleCodes = strings.TrimSpace(m.RoleCodes)
	m.OrgIDs = strings.TrimSpace(m.OrgIDs)
	if m.UserIDs != "" {
		m.UserIDs = fmt.Sprintf(",%s,", strings.Trim(m.UserIDs, ","))
	}
	if m.RoleCodes != "" {
		m.RoleCodes = fmt.Sprintf(",%s,", strings.Trim(m.RoleCodes, ","))
	}
	if m.OrgIDs != "" {
		m.OrgIDs = fmt.Sprintf(",%s,", strings.Trim(m.OrgIDs, ","))
	}
}

type MessageReceipt struct {
	Model     `rest:"*" displayName:"消息回执"`
	MessageID uint     `name:"关联消息ID" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	Message   *Message `name:"关联消息"`

	RecipientID       uint      `name:"接收人ID" gorm:"not null;UNIQUE_INDEX:kuu_unique"`
	RecipientUsername string    `name:"接收人账号"`
	Recipient         *User     `name:"接收人" gorm:"foreignkey:RecipientID"`
	RecipientSourceIP string    `name:"阅读人IP地址"`
	ReadAt            null.Time `name:"阅读时间"`
}
