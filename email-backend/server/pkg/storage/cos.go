// Package storage 腾讯云COS存储实现
package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"email-backend/server/config"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSStorage 腾讯云COS存储实现
type COSStorage struct {
	client    *cos.Client
	config    *config.StorageConfig
	cosConfig *config.COSConfig
	bucketURL string
}

// NewCOSStorage 创建腾讯云COS存储实例
func NewCOSStorage(cfg *config.StorageConfig, cosCfg *config.COSConfig) (*COSStorage, error) {
	if cosCfg.Region == "" {
		return nil, fmt.Errorf("COS region 未配置")
	}
	if cosCfg.SecretID == "" {
		return nil, fmt.Errorf("COS secret_id 未配置")
	}
	if cosCfg.SecretKey == "" {
		return nil, fmt.Errorf("COS secret_key 未配置")
	}

	// 构建Bucket URL
	bucketURL := fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cosCfg.Region)

	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("解析COS URL失败: %w", err)
	}

	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cosCfg.SecretID,
			SecretKey: cosCfg.SecretKey,
		},
	})

	return &COSStorage{
		client:    client,
		config:    cfg,
		cosConfig: cosCfg,
		bucketURL: bucketURL,
	}, nil
}

// Upload 上传文件到COS
func (s *COSStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err := s.client.Object.Put(ctx, key, reader, opt)
	if err != nil {
		return "", fmt.Errorf("COS上传失败: %w", err)
	}

	return s.GetURL(key), nil
}

// Delete 从COS删除文件
func (s *COSStorage) Delete(ctx context.Context, key string) error {
	_, err := s.client.Object.Delete(ctx, key)
	if err != nil {
		return fmt.Errorf("COS删除失败: %w", err)
	}
	return nil
}

// GetURL 获取文件访问URL
func (s *COSStorage) GetURL(key string) string {
	// 如果配置了自定义域名，使用自定义域名
	if s.config.CustomDomain != "" {
		return fmt.Sprintf("https://%s/%s", s.config.CustomDomain, key)
	}

	// 如果配置了基础URL，使用基础URL
	if s.config.BaseURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.config.BaseURL, "/"), key)
	}

	// 默认使用COS域名
	return fmt.Sprintf("%s/%s", s.bucketURL, key)
}

// GetSignedURL 获取带签名的临时访问URL
func (s *COSStorage) GetSignedURL(ctx context.Context, key string, expireSeconds int) (string, error) {
	presignedURL, err := s.client.Object.GetPresignedURL(ctx, http.MethodGet, key,
		s.cosConfig.SecretID, s.cosConfig.SecretKey, time.Duration(expireSeconds)*time.Second, nil)
	if err != nil {
		return "", fmt.Errorf("生成签名URL失败: %w", err)
	}
	return presignedURL.String(), nil
}

// GetBucketName 获取存储桶名称
func (s *COSStorage) GetBucketName() string {
	return s.config.Bucket
}

// GetRegion 获取区域
func (s *COSStorage) GetRegion() string {
	return s.cosConfig.Region
}
