package main

import (
	"fmt"
	"log"
	"os"

	"email-backend/internal/pkg/config"
	"email-backend/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Printf("加载配置失败: %v，使用默认配置", err)
		cfg = &config.Config{}
	}

	// 设置Gin模式
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin实例
	r := gin.Default()

	// 中间件
	r.Use(corsMiddleware())

	// 路由
	setupRoutes(r)

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("服务启动在 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func setupRoutes(r *gin.Engine) {
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{
			"status":  "ok",
			"service": "email-backend",
		})
	})

	// API v1路由组
	v1 := r.Group("/api/v1")
	{
		// 邮件路由
		emails := v1.Group("/emails")
		{
			emails.GET("", listEmails)
			emails.GET("/:id", getEmail)
			emails.POST("/:id/classify", classifyEmail)
		}

		// 账户路由
		accounts := v1.Group("/accounts")
		{
			accounts.GET("", listAccounts)
			accounts.POST("", createAccount)
			accounts.DELETE("/:id", deleteAccount)
			accounts.POST("/:id/test", testAccount)
		}

		// 同步路由
		sync := v1.Group("/sync")
		{
			sync.POST("", triggerSync)
			sync.GET("/status", syncStatus)
		}
	}

	// 检查环境变量
	_ = os.Getenv("CREDENTIAL_KEY")
}

// listEmails 获取邮件列表
func listEmails(c *gin.Context) {
	// TODO: 实现邮件列表查询
	response.Success(c, gin.H{
		"list":       []interface{}{},
		"total":      0,
		"page":       1,
		"page_size":  20,
		"total_pages": 0,
	})
}

// getEmail 获取邮件详情
func getEmail(c *gin.Context) {
	// TODO: 实现邮件详情查询
	response.Success(c, gin.H{
		"id":         c.Param("id"),
		"subject":    "测试邮件",
		"sender":     "test@example.com",
		"category":   "unclassified",
		"priority":   "medium",
	})
}

// classifyEmail 分类邮件
func classifyEmail(c *gin.Context) {
	// TODO: 实现邮件分类
	response.Success(c, gin.H{
		"id":          c.Param("id"),
		"category":    "work_normal",
		"priority":    "medium",
		"confidence":  0.85,
	})
}

// listAccounts 获取账户列表
func listAccounts(c *gin.Context) {
	// TODO: 实现账户列表查询
	response.Success(c, gin.H{
		"list": []interface{}{},
	})
}

// createAccount 创建账户
func createAccount(c *gin.Context) {
	// TODO: 实现创建账户
	response.Created(c, gin.H{
		"id":      1,
		"email":   "user@126.com",
		"provider": "126",
	})
}

// deleteAccount 删除账户
func deleteAccount(c *gin.Context) {
	// TODO: 实现删除账户
	response.SuccessWithMessage(c, "删除成功", nil)
}

// testAccount 测试账户连接
func testAccount(c *gin.Context) {
	// TODO: 实现测试账户连接
	response.Success(c, gin.H{
		"id":      c.Param("id"),
		"status":  "connected",
		"message": "连接成功",
	})
}

// triggerSync 触发同步
func triggerSync(c *gin.Context) {
	// TODO: 实现触发同步
	response.Success(c, gin.H{
		"task_id": "sync_001",
		"status":  "started",
	})
}

// syncStatus 获取同步状态
func syncStatus(c *gin.Context) {
	// TODO: 实现获取同步状态
	response.Success(c, gin.H{
		"last_sync": "2026-04-08T10:00:00Z",
		"status":    "idle",
	})
}
