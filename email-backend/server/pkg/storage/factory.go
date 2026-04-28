// Package storage 对象存储工厂
package storage

import (
	"fmt"

	"email-backend/server/config"
)

const (
	// TypeOSS 阿里云OSS
	TypeOSS = "oss"
	// TypeCOS 腾讯云COS
	TypeCOS = "cos"
)

// Factory 存储服务工厂
type Factory struct {
	config *config.Config
}

// NewFactory 创建存储服务工厂
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{config: cfg}
}

// Create 创建存储服务实例
func (f *Factory) Create() (Storage, error) {
	return CreateStorage(f.config)
}

// CreateStorage 根据配置创建存储服务实例
func CreateStorage(cfg *config.Config) (Storage, error) {
	storageType := cfg.Storage.Type

	switch storageType {
	case TypeOSS:
		return NewOSSStorage(&cfg.Storage, &cfg.OSS)
	case TypeCOS:
		return NewCOSStorage(&cfg.Storage, &cfg.COS)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
	}
}

// CreateOSS 创建阿里云OSS存储实例
func CreateOSS(cfg *config.Config) (*OSSStorage, error) {
	return NewOSSStorage(&cfg.Storage, &cfg.OSS)
}

// CreateCOS 创建腾讯云COS存储实例
func CreateCOS(cfg *config.Config) (*COSStorage, error) {
	return NewCOSStorage(&cfg.Storage, &cfg.COS)
}
