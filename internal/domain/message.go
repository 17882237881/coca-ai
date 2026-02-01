package domain

import "time"

// MessageRole 定义消息角色类型
type MessageRole string 

const (
	RoleUser      MessageRole = "user"      // 用户消息
	RoleAssistant MessageRole = "assistant" // AI 助手回复
	RoleSystem    MessageRole = "system"    // 系统提示 (System Prompt)
)

// Message 表示一条对话消息
type Message struct {
	ID        int64
	SessionID int64
	Role      MessageRole
	Content   string
	CreatedAt time.Time
}

// IsUser 判断是否为用户消息
func (m *Message) IsUser() bool {
	return m.Role == RoleUser
}

// IsAssistant 判断是否为 AI 回复
func (m *Message) IsAssistant() bool {
	return m.Role == RoleAssistant
}

// IsSystem 判断是否为系统消息
func (m *Message) IsSystem() bool {
	return m.Role == RoleSystem
}
