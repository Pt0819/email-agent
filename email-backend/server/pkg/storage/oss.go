// Package storage 阿里云OSS存储实现
package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"email-backend/server/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

// OSSStorage 阿里云OSS存储实现
type OSSStorage struct {
	client   *oss.Client
	bucket   *oss.Bucket
	config   *config.StorageConfig
	ossConfig *config.OSSConfig
}

// NewOSSStorage 创建阿里云OSS存储实例
func NewOSSStorage(cfg *config.StorageConfig, ossCfg *config.OSSConfig) (*OSSStorage, error) {
	if ossCfg.Endpoint == "" {
		return nil, fmt.Errorf("OSS endpoint 未配置")
	}
	if ossCfg.AccessKeyID == "" {
		return nil, fmt.Errorf("OSS access_key_id 未配置")
	}
	if ossCfg.AccessKeySecret == "" {
		return nil, fmt.Errorf("OSS access_key_secret 未配置")
	}

	client, err := oss.New(ossCfg.Endpoint, ossCfg.AccessKeyID, ossCfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取OSS Bucket失败: %w", err)
	}

	return &OSSStorage{
		client:    client,
		bucket:    bucket,
		config:    cfg,
		ossConfig: ossCfg,
	}, nil
}

// SetBucketPublicRead 设置Bucket为公共读（私有写）
func (s *OSSStorage) SetBucketPublicRead() error {
	return s.client.SetBucketACL(s.config.Bucket, oss.ACLPublicRead)
}

// Upload 上传文件到OSS
func (s *OSSStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	// 使用进度条选项上传
	options := []oss.Option{
		oss.ContentType(contentType),
	}

	if err := s.bucket.PutObject(key, reader, options...); err != nil {
		return "", fmt.Errorf("OSS上传失败: %w", err)
	}

	return s.GetURL(key), nil
}

// Delete 从OSS删除文件
func (s *OSSStorage) Delete(ctx context.Context, key string) error {
	if err := s.bucket.DeleteObject(key); err != nil {
		return fmt.Errorf("OSS删除失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *OSSStorage) GetURL(key string) string {
	// 如果配置了自定义域名，使用自定义域名
	if s.config.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.config.CustomDomain, key)
	}

	// 如果配置了基础URL，使用基础URL
	if s.config.BaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), key)
	}

	// 默认使用OSS域名
	return fmt.Sprintf("https://%s.%s/%s", s.config.Bucket, s.ossConfig.Endpoint, key)
}

// GetSignedURL 获取带签名的临时访问URL
func (s *OSSStorage) GetSignedURL(ctx context.Context, key string, expireSeconds int) (string, error) {
	signedURL, err := s.bucket.SignURL(key, oss.HTTPGet, int64(expireSeconds))
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %w", err)
	}
	return signedURL, nil
}

// GetBucketName 获取存储桶名称
func (s *OSSStorage) GetBucketName() string {
	return s.config.Bucket
}

// GetEndpoint 获取Endpoint
func (s *OSSStorage) GetEndpoint() string {
	return s.ossConfig.Endpoint
}

// BuildKey 构建文件key，包含路径前缀
func BuildKey(prefix, filename string) string {
	if prefix == "" {
		return filename
	}
	return filepath.Join(prefix, filename)
}

// SanitizeFilename 清理文件名，移除不安全字符
func SanitizeFilename(filename string) string {
	// 替换不安全字符
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

// ParseURL 从完整URL中提取key
func ParseURL(fullURL string, bucket string, endpoint string) (string, error) {
	// 解析URL
	u, err := url.Parse(fullURL)
	if err != nil {
		return "", err
	}

	// 标准OSS URL格式: https://bucket.endpoint/key
	host := u.Host
	expectedHost := fmt.Sprintf("%s.%s", bucket, endpoint)

	if host == expectedHost {
		// 移除前导斜杠
		return strings.TrimPrefix(u.Path, "/"), nil
	}

	// 可能是自定义域名，直接返回路径
	return strings.TrimPrefix(u.Path, "/"), nil
}

// GenerateTimestampKey 生成带时间戳的唯一文件名
func GenerateTimestampKey(prefix, ext string) string {
	timestamp := time.Now().Format("20060102/150405")
	uniqueID := time.Now().UnixNano() % 10000
	return fmt.Sprintf("%s/%s_%04d%s", prefix, timestamp, uniqueID, ext)
}
