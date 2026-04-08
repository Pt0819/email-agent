// Package global 全局对象
package global

import (
	"email-backend/server/core"

	"gorm.io/gorm"
)

// DB 获取全局数据库实例
func DB() *gorm.DB {
	return core.GlobalDB
}

// Config 获取全局配置
func Config() *core.GlobalConfig {
	return core.GlobalConfig
}