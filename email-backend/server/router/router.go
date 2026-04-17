// Package router 路由注册
package router

import (
	"email-backend/server/api/v1"
	"email-backend/server/config"
	"email-backend/server/global"
	"email-backend/server/middleware"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"

	"github.com/gin-gonic/gin"
)

// Setup 路由设置
func Setup(r *gin.Engine, cfg *config.Config) {
	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", v1.HealthCheck)

	// 创建Agent客户端
	agentClient := agent.NewClient(&cfg.Agent)

	// 创建Repository
	emailRepo := repository.NewEmailRepository(global.DB())

	// API v1
	v1Group := r.Group("/api/v1")
	{
		// 邮件路由
		v1.SetupEmailRoutes(v1Group, agentClient, emailRepo)

		// 账户路由
		v1.SetupAccountRoutes(v1Group)

		// 同步路由
		v1.SetupSyncRoutes(v1Group, agentClient)
	}
}