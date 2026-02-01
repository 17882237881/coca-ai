package ioc

import (
	"coca-ai/internal/repository/dao"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	// Default DSN for local development
	dsn := "root:root@tcp(localhost:13306)/coca_db?charset=utf8mb4&parseTime=True&loc=Local"

	// Override with Docker environment if available
	if host := os.Getenv("MYSQL_HOST"); host != "" {
		port := os.Getenv("MYSQL_PORT")
		if port == "" {
			port = "3306"
		}
		dsn = "root:root@tcp(" + host + ":" + port + ")/coca_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	// 自动迁移表结构 (Auto Migration)
	err = db.AutoMigrate(&dao.User{}, &dao.Session{}, &dao.Message{})
	if err != nil {
		panic(err)
	}

	return db
}
