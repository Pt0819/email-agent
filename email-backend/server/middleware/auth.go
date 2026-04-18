// Package middleware 认证中间件
package middleware

import (
	emailResponse "email-backend/server/model/response"
	"email-backend/server/service"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID 用户ID在context中的key
	ContextKeyUserID = "userID"
	// ContextKeyUserEmail 用户邮箱在context中的key
	ContextKeyUserEmail = "userEmail"
)

// JWTAuth JWT认证中间件
func JWTAuth(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// 解析Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			unauthorized(c, "invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			unauthorized(c, "empty token")
			c.Abort()
			return
		}

		// 验证Token
		claims, err := userService.ValidateToken(tokenString)
		if err != nil {
			unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// 将用户信息注入到context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyUserEmail, claims.Email)

		c.Next()
	}
}

// GetUserID 从context获取用户ID
func GetUserID(c *gin.Context) int64 {
	userID, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0
	}
	return userID.(int64)
}

// GetUserEmail 从context获取用户邮箱
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get(ContextKeyUserEmail)
	if !exists {
		return ""
	}
	return email.(string)
}

// unauthorized 返回未授权响应
func unauthorized(c *gin.Context, message string) {
	c.JSON(401, emailResponse.Response{
		Code:    emailResponse.CodeUnauthorized,
		Message: message,
	})
}
