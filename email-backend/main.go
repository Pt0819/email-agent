// Package main 程序入口
package main

import (
	"fmt"
	"log"

	"email-backend/server/core"
	"email-backend/server/model"
	"email-backend/server/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 初始化配置
	if err := core.InitConfig(""); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 2. 初始化数据库
	if err := core.InitDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}
	defer core.Close()

	// 3. 自动迁移表结构
	if err := core.GlobalDB.AutoMigrate(
		&model.Email{},
		&model.EmailAccount{},
		&model.ActionItem{},
	); err != nil {
		log.Printf("警告: 自动迁移失败: %v", err)
	}

	// 4. 初始化加密器
	if err := core.InitEncryptor(); err != nil {
		log.Fatalf("初始化加密器失败: %v", err)
	}

	// 5. 初始化邮件Provider
	core.InitProviders()

	// 6. 设置Gin模式
	if core.GlobalConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 7. 创建Gin实例
	r := gin.New()

	// 8. 设置路由（传入配置以初始化Agent客户端）
	router.Setup(r, core.GlobalConfig)

	// 9. 启动服务
	addr := fmt.Sprintf(":%d", core.GlobalConfig.Server.Port)
	log.Printf("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}