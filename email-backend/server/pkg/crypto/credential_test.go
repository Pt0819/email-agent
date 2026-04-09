package crypto

import (
	"strings"
	"testing"
)

func TestNewCredentialEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr error
	}{
		{
			name:    "valid key",
			key:     "12345678901234567890123456789012",
			wantErr: nil,
		},
		{
			name:    "key too short",
			key:     "short",
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "key too long",
			key:     "123456789012345678901234567890123",
			wantErr: ErrInvalidKeyLength,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: ErrInvalidKeyLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCredentialEncryptor(tt.key)
			if err != tt.wantErr {
				t.Errorf("NewCredentialEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	// 使用固定的32字节密钥
	key := "12345678901234567890123456789012"
	enc, err := NewCredentialEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "normal text",
			plaintext: "hello world",
		},
		{
			name:      "chinese text",
			plaintext: "这是一段中文文本",
		},
		{
			name:      "special characters",
			plaintext: "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "email credential",
			plaintext: "abcdefghijklmnop",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "long text",
			plaintext: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, iv, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			// 加密后密文和IV不应该为空(除非输入为空)
			if tt.plaintext != "" {
				if encrypted == "" {
					t.Error("Encrypted should not be empty")
				}
				if iv == "" {
					t.Error("IV should not be empty")
				}
				// 加密后密文应该和原文不同
				if encrypted == tt.plaintext {
					t.Error("Encrypted should be different from plaintext")
				}
			}

			decrypted, err := enc.Decrypt(encrypted, iv)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptProducesDifferentCiphertext(t *testing.T) {
	// 相同明文应该产生不同的密文(因为每次生成不同的IV)
	key := "12345678901234567890123456789012"
	enc, err := NewCredentialEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	plaintext := "same text"

	encrypted1, iv1, _ := enc.Encrypt(plaintext)
	encrypted2, iv2, _ := enc.Encrypt(plaintext)

	// IV 应该不同
	if iv1 == iv2 {
		t.Error("IVs should be different for same plaintext")
	}

	// 密文也应该不同
	if encrypted1 == encrypted2 {
		t.Error("Ciphertexts should be different for same plaintext")
	}

	// 但解密结果应该一致
	decrypted1, _ := enc.Decrypt(encrypted1, iv1)
	decrypted2, _ := enc.Decrypt(encrypted2, iv2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both decryptions should equal the plaintext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := "12345678901234567890123456789012"
	key2 := "abcdefghijklmnopqrstuvwxyz123456"

	enc1, _ := NewCredentialEncryptor(key1)
	enc2, _ := NewCredentialEncryptor(key2)

	plaintext := "secret credential"
	encrypted, iv, _ := enc1.Encrypt(plaintext)

	// 用不同的密钥解密应该失败
	_, err := enc2.Decrypt(encrypted, iv)
	if err == nil {
		t.Error("Decrypt with wrong key should fail")
	}
}

func TestDecryptWithCorruptedData(t *testing.T) {
	key := "12345678901234567890123456789012"
	enc, _ := NewCredentialEncryptor(key)

	plaintext := "test data"
	encrypted, iv, _ := enc.Encrypt(plaintext)

	tests := []struct {
		name      string
		encrypted string
		iv        string
		wantErr   error
	}{
		{
			name:      "empty encrypted",
			encrypted: "",
			iv:        iv,
			wantErr:   nil, // 空字符串直接返回空
		},
		{
			name:      "empty iv",
			encrypted: encrypted,
			iv:        "",
			wantErr:   nil, // 空字符串直接返回空
		},
		{
			name:      "invalid base64",
			encrypted: "not-valid-base64!!!",
			iv:        iv,
			wantErr:   ErrInvalidCiphertext,
		},
		{
			name:      "invalid iv base64",
			encrypted: encrypted,
			iv:        "not-valid-base64!!!",
			wantErr:   ErrInvalidIV,
		},
		{
			name:      "corrupted ciphertext",
			encrypted: "Y29ycnVwdGVkZGF0YQ==", // "corrupteddata" base64
			iv:        iv,
			wantErr:   ErrDecryptFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := enc.Decrypt(tt.encrypted, tt.iv)
			if err != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("GenerateKey() length = %d, want 32", len(key1))
	}

	// 生成两次应该不同
	key2, _ := GenerateKey()
	if key1 == key2 {
		t.Error("Two calls to GenerateKey() should produce different keys")
	}
}

func TestGenerateKeyBase64(t *testing.T) {
	key, err := GenerateKeyBase64()
	if err != nil {
		t.Fatalf("GenerateKeyBase64() error = %v", err)
	}

	// Base64 编码的32字节应该是44字符左右
	if len(key) < 40 {
		t.Errorf("GenerateKeyBase64() length = %d, seems too short", len(key))
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	enc, _ := NewCredentialEncryptor(key)
	plaintext := "benchmark test credential data"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Encrypt(plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := "12345678901234567890123456789012"
	enc, _ := NewCredentialEncryptor(key)
	plaintext := "benchmark test credential data"
	encrypted, iv, _ := enc.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc.Decrypt(encrypted, iv)
	}
}
