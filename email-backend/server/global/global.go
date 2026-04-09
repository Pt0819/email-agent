// Package global 全局对象
package global

import (
	"email-backend/server/config"
	"email-backend/server/core"
	"email-backend/server/pkg/crypto"

	"gorm.io/gorm"
)

// DB 获取全局数据库实例
func DB() *gorm.DB {
	return core.GlobalDB
}

// Config 获取全局配置
func Config() *config.Config {
	return core.GlobalConfig
}

// Encryptor 获取全局凭证加密器
func Encryptor() *crypto.CredentialEncryptor {
	return core.GlobalEncryptor
}