// Package main 程序入口
package main

import (
	"fmt"
	"log"

	"email-backend/server/core"
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

	// 3. 设置Gin模式
	if core.GlobalConfig.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 4. 创建Gin实例
	r := gin.New()

	// 5. 设置路由
	router.Setup(r)

	// 6. 启动服务
	addr := fmt.Sprintf(":%d", core.GlobalConfig.Server.Port)
	log.Printf("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}