// Package core 核心功能模块
package core

import (
	"fmt"

	"email-backend/server/config"
	"email-backend/server/pkg/crypto"
	"email-backend/server/pkg/email/provider"
	"email-backend/server/pkg/storage"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// GlobalConfig 全局配置
	GlobalConfig *config.Config
	// GlobalDB 全局数据库连接
	GlobalDB *gorm.DB
	// GlobalEncryptor 全局凭证加密器
	GlobalEncryptor *crypto.CredentialEncryptor
	// GlobalStorage 全局对象存储服务
	GlobalStorage storage.Storage
)

// Config 返回全局配置
func Config() *config.Config {
	return GlobalConfig
}

// DB 返回全局数据库连接
func DB() *gorm.DB {
	return GlobalDB
}

// Encryptor 返回全局加密器
func Encryptor() *crypto.CredentialEncryptor {
	return GlobalEncryptor
}

// Storage 返回全局对象存储服务
func Storage() storage.Storage {
	return GlobalStorage
}

// InitStorage 初始化对象存储服务
func InitStorage() error {
	storageService, err := storage.CreateStorage(GlobalConfig)
	if err != nil {
		return fmt.Errorf("初始化存储服务失败: %w", err)
	}

	GlobalStorage = storageService
	fmt.Printf("存储服务初始化成功: 类型=%s, Bucket=%s\n", GlobalConfig.Storage.Type, GlobalConfig.Storage.Bucket)

	// 尝试设置Bucket为公共读（如果失败不影响服务启动）
	if ossStorage, ok := storageService.(*storage.OSSStorage); ok {
		if err := ossStorage.SetBucketPublicRead(); err != nil {
			fmt.Printf("警告: 设置Bucket公共读失败，请手动在OSS控制台设置: %v\n", err)
		} else {
			fmt.Printf("Bucket ACL已设置为公共读\n")
		}
	}

	return nil
}

// InitConfig 初始化配置
func InitConfig(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	GlobalConfig = cfg
	return nil
}

// InitEncryptor 初始化凭证加密器
func InitEncryptor() error {
	key := GlobalConfig.Security.CredentialKey
	if key == "" {
		return fmt.Errorf("凭证加密密钥未配置，请设置环境变量 CREDENTIAL_KEY")
	}

	enc, err := crypto.NewCredentialEncryptor(key)
	if err != nil {
		return fmt.Errorf("创建加密器失败: %w", err)
	}

	GlobalEncryptor = enc
	return nil
}

// InitProviders 初始化邮件Provider
func InitProviders() {
	// Provider 在各自的文件中通过 init() 自动注册
	// 这里可以添加日志记录
	fmt.Printf("已注册的邮件Provider: %v\n", provider.ListProviders())
}

// InitDB 初始化数据库连接
func InitDB() error {
	fmt.Printf("数据库配置: host=%s, port=%d, user=%s, dbname=%s, password=%s\n",
		GlobalConfig.Database.Host,
		GlobalConfig.Database.Port,
		GlobalConfig.Database.Username,
		GlobalConfig.Database.DBName,
		GlobalConfig.Database.Password,
	)

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		GlobalConfig.Database.Username,
		GlobalConfig.Database.Password,
		GlobalConfig.Database.Host,
		GlobalConfig.Database.Port,
		GlobalConfig.Database.DBName,
	)

	var gormLogger logger.Interface
	if GlobalConfig.Server.Mode == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(GlobalConfig.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(GlobalConfig.Database.MaxIdleConns)

	GlobalDB = db
	return nil
}

// Close 关闭资源
func Close() {
	if GlobalDB != nil {
		sqlDB, _ := GlobalDB.DB()
		sqlDB.Close()
	}
}