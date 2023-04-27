package kuu

import (
	"fmt"
	"gopkg.in/guregu/null.v3"
	"strings"
	"time"
)

const (
	MessageStatusDraft = "100"
	MessageStatusSent  = "200"
)

type Message struct {
	Model `rest:"*" displayName:"系统消息"`

	Subject     string      `name:"消息标题"`
	Content     null.String `name:"消息内容" gorm:"not null"`
	Attachments []File      `name:"消息附件" gorm:"polymorphic:Owner;polymorphic_value:Message.Attachments"`

	Status string `name:"消息状态" enum:"MessageStatus"`

	SenderID       uint      `name:"发送人ID"`
	SenderUsername string    `name:"发送人账号"`
	Sender         *User     `name:"发送人" gorm:"foreignkey:SenderID"`
	SenderSourceIP string    `name:"发送人IP地址"`
	SentAt         time.Time `name:"发送时间"`

	RecipientOrgIDs    string           `name:"关联接收组织ID（多个以英文逗号分隔）"`
	RecipientUserIDs   string           `name:"关联接收人ID（多个以英文逗号分隔）"`
	RecipientRoleCodes string           `name:"关联接收角色编码（多个以英文逗号分隔）"`
	RecipientReceipts  []MessageReceipt `name:"消息接收/阅读回执"`
}

// BeforeCreate
func (m *Message) BeforeCreate() {
	if m.Status == "" {
		m.Status = MessageStatusDraft
	}
	m.SentAt = time.Now()
	// TODO 处理IP的问题
	//if c := GetRoutineRequestContext(); c != nil {
	//	m.SenderSourceIP = c.ClientIP()
	//	m.SenderID = c.SignInfo.UID
	//	m.SenderUsername = c.SignInfo.Username
	//}
	m.RecipientUserIDs = strings.TrimSpace(m.RecipientUserIDs)
	m.RecipientRoleCodes = strings.TrimSpace(m.RecipientRoleCodes)
	m.RecipientOrgIDs = strings.TrimSpace(m.RecipientOrgIDs)
	if m.RecipientUserIDs != "" {
		m.RecipientUserIDs = fmt.Sprintf(",%s,", strings.Trim(m.RecipientUserIDs, ","))
	}
	if m.RecipientRoleCodes != "" {
		m.RecipientRoleCodes = fmt.Sprintf(",%s,", strings.Trim(m.RecipientRoleCodes, ","))
	}
	if m.RecipientOrgIDs != "" {
		m.RecipientOrgIDs = fmt.Sprintf(",%s,", strings.Trim(m.RecipientOrgIDs, ","))
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
