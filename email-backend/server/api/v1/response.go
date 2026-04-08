// Package v1 公共响应函数
package v1

import (
	"net/http"

	respModel "email-backend/server/model/response"

	"github.com/gin-gonic/gin"
)

// success 成功响应
func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, respModel.Response{
		Code:    respModel.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// created 创建成功
func created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, respModel.Response{
		Code:    respModel.CodeSuccess,
		Message: "created",
		Data:    data,
	})
}

// badRequest 请求参数错误
func badRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, respModel.Response{
		Code:    respModel.CodeBadRequest,
		Message: message,
	})
}

// notFound 资源不存在
func notFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, respModel.Response{
		Code:    respModel.CodeNotFound,
		Message: message,
	})
}

// errorResp 通用错误
func errorResp(c *gin.Context, status int, message string) {
	c.JSON(status, respModel.Response{
		Code:    status,
		Message: message,
	})
}