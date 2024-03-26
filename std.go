package kuu

type MessageType string

const (
	MessageTypeSilent       MessageType = "SILENT"
	MessageTypeWarn         MessageType = "WARN"
	MessageTypeInfo         MessageType = "INFO"
	MessageTypeError        MessageType = "ERROR"
	MessageTypeNotification MessageType = "NOTIFICATION"
)

type Reply struct {
	RequestId     string      `json:"rid"`
	Success       bool        `json:"success"`
	Code          int         `json:"code"`
	Data          any         `json:"data,omitempty"`
	Message       string      `json:"msg,omitempty"`
	MessageType   MessageType `json:"msg_type,omitempty"`
	MessageValues Map         `json:"msg_values,omitempty"`
}
