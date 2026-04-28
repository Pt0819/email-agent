// Package storage 对象存储抽象层
package storage

import (
	"context"
	"io"
)

// Storage 对象存储接口
type Storage interface {
	// Upload 上传文件
	// key: 文件在存储桶中的路径/名称
	// reader: 文件内容
	// contentType: 文件MIME类型
	// 返回访问URL
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// GetURL 获取文件访问URL
	GetURL(key string) string

	// GetSignedURL 获取带签名的临时访问URL
	GetSignedURL(ctx context.Context, key string, expireSeconds int) (string, error)
}
