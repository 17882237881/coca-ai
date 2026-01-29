package domain

import "time"

// User 领域对象
// 对应业务概念中的“用户”，不依赖任何数据库标签
type User struct {
	Id       int64
	Email    string
	Password string
	Ctime    time.Time
}
