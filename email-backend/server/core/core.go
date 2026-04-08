// Package core 核心功能模块
package core

import (
	"fmt"

	"email-backend/server/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	// GlobalConfig 全局配置
	GlobalConfig *config.Config
	// GlobalDB 全局数据库连接
	GlobalDB *gorm.DB
)

// InitConfig 初始化配置
func InitConfig(configPath string) error {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return err
	}
	GlobalConfig = cfg
	return nil
}

// InitDB 初始化数据库连接
func InitDB() error {
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