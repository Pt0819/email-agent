// Package crypto 凭证加密工具
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
)

var (
	ErrInvalidKeyLength = errors.New("密钥长度必须为32字节或64字符的hex字符串")
	ErrInvalidCiphertext = errors.New("无效的密文")
	ErrInvalidIV        = errors.New("无效的IV")
	ErrDecryptFailed    = errors.New("解密失败")
)

// CredentialEncryptor 凭证加密器
// 使用 AES-256-GCM 算法进行加密
type CredentialEncryptor struct {
	key []byte // 32字节密钥，用于 AES-256
}

// NewCredentialEncryptor 创建凭证加密器
// masterKey 可以是32字节的原始字符串，或64字符的hex编码字符串
func NewCredentialEncryptor(masterKey string) (*CredentialEncryptor, error) {
	var key []byte

	// 尝试解析为hex字符串
	if len(masterKey) == 64 {
		decoded, err := hex.DecodeString(masterKey)
		if err == nil && len(decoded) == 32 {
			key = decoded
		}
	}

	// 如果不是hex，则直接使用原始字节
	if key == nil {
		if len(masterKey) != 32 {
			return nil, ErrInvalidKeyLength
		}
		key = []byte(masterKey)
	}

	return &CredentialEncryptor{key: key}, nil
}

// Encrypt 加密凭证
// 返回加密后的密文(Base64)和IV(Base64)
func (e *CredentialEncryptor) Encrypt(plaintext string) (encrypted, iv string, err error) {
	if plaintext == "" {
		return "", "", nil
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", "", err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// 生成随机 nonce (GCM 标准要求12字节)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", err
	}

	// 加密数据
	// Seal 会将 nonce 附加到密文前面
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// 返回 Base64 编码的结果
	encrypted = base64.StdEncoding.EncodeToString(ciphertext)
	iv = base64.StdEncoding.EncodeToString(nonce)

	return encrypted, iv, nil
}

// Decrypt 解密凭证
// encrypted: Base64 编码的密文
// iv: Base64 编码的 IV/nonce
func (e *CredentialEncryptor) Decrypt(encrypted, iv string) (string, error) {
	if encrypted == "" || iv == "" {
		return "", nil
	}

	// Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	nonce, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", ErrInvalidIV
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	// 创建 GCM 模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 验证 nonce 长度
	if len(nonce) != gcm.NonceSize() {
		return "", ErrInvalidIV
	}

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptFailed
	}

	return string(plaintext), nil
}

// GenerateKey 生成32字节的随机密钥
// 用于生成新的加密密钥
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return string(key), nil
}

// GenerateKeyBase64 生成32字节随机密钥并返回Base64编码
func GenerateKeyBase64() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
