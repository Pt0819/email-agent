// Package router 路由注册
package router

import (
	"email-backend/server/api/v1"
	"email-backend/server/config"
	"email-backend/server/global"
	"email-backend/server/middleware"
	"email-backend/server/pkg/agent"
	"email-backend/server/repository"
	"email-backend/server/service"

	"github.com/gin-gonic/gin"
)

// Setup 路由设置
func Setup(r *gin.Engine, cfg *config.Config) *service.SyncScheduler {
	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 健康检查（公开）
	r.GET("/health", v1.HealthCheck)

	// 创建Agent客户端
	agentClient := agent.NewClient(&cfg.Agent)

	// 创建Repository
	emailRepo := repository.NewEmailRepository(global.DB())
	userRepo := repository.NewUserRepository(global.DB())
	accountRepo := repository.NewAccountRepository(global.DB())
	steamRepo := repository.NewSteamRepository(global.DB())

	// 创建Service
	userService := service.NewUserService(userRepo)

	// 创建同步调度器
	scheduler := service.NewSyncScheduler(accountRepo, emailRepo, agentClient)

	// JWT中间件
	jwtAuth := middleware.JWTAuth(userService)

	// API v1
	v1Group := r.Group("/api/v1")
	{
		// 公开路由（无需认证）
		v1.SetupAuthRoutes(v1Group, userService)

		// 受保护路由（需要认证）
		protected := v1Group.Group("")
		protected.Use(jwtAuth)
		{
			// auth/me 也需要保护
			protected.GET("/auth/me", v1.NewAuthHandler(userService).Me)

			// 邮件路由
			v1.SetupEmailRoutes(protected, agentClient, emailRepo)

			// 账户路由
			v1.SetupAccountRoutes(protected)

			// 同步路由（传入调度器）
			v1.SetupSyncRoutes(protected, agentClient, scheduler)

			// 摘要路由
			v1.SetupSummaryRoutes(protected, agentClient, emailRepo)

		// Steam路由
		v1.SetupSteamRoutes(protected, agentClient, steamRepo, emailRepo)
		}
	}

	return scheduler
}
