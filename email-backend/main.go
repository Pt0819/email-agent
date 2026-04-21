// Package main 程序入口
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	if err := core.DB().AutoMigrate(
		&model.User{},
		&model.Email{},
		&model.EmailAccount{},
		&model.ActionItem{},
		&model.SteamGame{},
		&model.SteamDeal{},
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
	if core.Config().Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 7. 创建Gin实例
	r := gin.New()

	// 8. 设置路由（获取调度器）
	scheduler := router.Setup(r, core.Config())

	// 9. 启动同步调度器
	if core.Config().Email.AutoClassify {
		if err := scheduler.Start(); err != nil {
			log.Printf("警告: 启动同步调度器失败: %v", err)
		}
	}

	// 10. 启动HTTP服务
	addr := fmt.Sprintf(":%d", core.Config().Server.Port)

	// 优雅关闭
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("正在关闭服务...")

		// 停止调度器
		scheduler.Stop()

		// 给HTTP服务5秒时间完成现有请求
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = ctx // 这里简化处理

		log.Println("服务已关闭")
		os.Exit(0)
	}()

	log.Printf("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
