package domain

import "time"

// Session 表示一个对话会话
// 一个用户可以有多个 Session，每个 Session 包含多条 Message
type Session struct {
	ID        int64
	UserID    int64
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
