// crypto.go 提供加密相关工具函数。
// 注意：ECB 模式仅用于兼容旧系统，新项目请使用 GCM、CBC 等更安全的模式。
package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
)

// ==================== Hash 工具 ====================

// MD5 计算字符串的 MD5 哈希值，返回十六进制字符串。
func MD5(data string) string {
	h := md5.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// MD5Bytes 计算字节数组的 MD5 哈希值，返回十六进制字符串。
func MD5Bytes(data []byte) string {
	h := md5.Sum(data)
	return hex.EncodeToString(h[:])
}

// SHA1 计算字符串的 SHA1 哈希值。
func SHA1(data string) string {
	h := sha1.Sum([]byte(data))
	return hex.EncodeToString(h[:])
}

// SHA256 计算字符串的 SHA256 哈希值。
func SHA256(data string) string {
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

// HMacSHA256 TODO: 实现 HMAC-SHA256

// ==================== Base64 工具 ====================

// Base64Encode 将字节数组编码为 Base64 字符串。
func Base64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// Base64Decode 将 Base64 字符串解码为字节数组。
func Base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Base64EncodeString 将字符串编码为 Base64。
func Base64EncodeString(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64DecodeString 将 Base64 解码为字符串。
func Base64DecodeString(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	return string(b), err
}

// Base64URLEncode URL 安全的 Base64 编码（替换 +/ 为 -_）。
func Base64URLEncode(data []byte) string {
	return base64.URLEncoding.EncodeToString(data)
}

// Base64URLDecode URL 安全的 Base64 解码。
func Base64URLDecode(s string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(s)
}

// ==================== AES-ECB（兼容旧系统） ====================

// aesECB 实现 ECB 模式的 AES 加密/解密。
// 警告：ECB 模式不安全，仅用于兼容旧系统。新项目请使用 GCM 或 CBC。
type aesECB struct {
	b         cipher.Block
	blockSize int
}

func newECB(b cipher.Block) *aesECB {
	return &aesECB{
		b:         b,
		blockSize: b.BlockSize(),
	}
}

// pkcs7Pad 使用 PKCS7 填充。
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7Unpad 移除 PKCS7 填充。
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, errors.New("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, errors.New("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// AESEncryptECB AES-ECB 加密（PKCS7 填充）。
// 警告：ECB 模式不安全，仅用于兼容旧系统。
func AESEncryptECB(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// PKCS7 填充
	padded := pkcs7Pad(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(padded))

	// ECB 模式加密
	for i := 0; i < len(padded); i += block.BlockSize() {
		block.Encrypt(ciphertext[i:i+block.BlockSize()], padded[i:i+block.BlockSize()])
	}

	return ciphertext, nil
}

// AESDecryptECB AES-ECB 解密（PKCS7 填充）。
// 警告：ECB 模式不安全，仅用于兼容旧系统。
func AESDecryptECB(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(ciphertext))

	// ECB 模式解密
	for i := 0; i < len(ciphertext); i += block.BlockSize() {
		block.Decrypt(plaintext[i:i+block.BlockSize()], ciphertext[i:i+block.BlockSize()])
	}

	// 移除 PKCS7 填充
	return pkcs7Unpad(plaintext)
}

// AESEncryptECBString AES-ECB 加密并返回 Base64 字符串。
func AESEncryptECBString(plaintext string, key []byte) (string, error) {
	encrypted, err := AESEncryptECB([]byte(plaintext), key)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// AESDecryptECBString 从 Base64 字符串解密 AES-ECB。
func AESDecryptECBString(ciphertext string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	decrypted, err := AESDecryptECB(data, key)
	if err != nil {
		return "", err
	}
	return string(decrypted), nil
}

// ==================== Hex 工具 ====================

// HexEncode 将字节数组编码为十六进制字符串。
func HexEncode(data []byte) string {
	return hex.EncodeToString(data)
}

// HexDecode 将十六进制字符串解码为字节数组。
func HexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

// ==================== XOR 工具 ====================

// XOR 对两个字节数组进行异或操作。
// 如果长度不同，以较短者为准。
func XOR(a, b []byte) []byte {
	length := len(a)
	if len(b) < length {
		length = len(b)
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = a[i] ^ b[i]
	}
	return result
}

// XORSingle 使用单字节对所有字节进行异或。
func XORSingle(data []byte, key byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key
	}
	return result
}

// ==================== Stream 工具 ====================

// HashReader 计算 io.Reader 的哈希值。
func HashReader(r io.Reader, hasher func() hash.Hash) (string, error) {
	h := hasher()
	_, err := io.Copy(h, r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// ==================== 常量定义 ====================

const (
	// AESKeySize128 AES 128 位密钥长度。
	AESKeySize128 = 16
	// AESKeySize192 AES 192 位密钥长度。
	AESKeySize192 = 24
	// AESKeySize256 AES 256 位密钥长度。
	AESKeySize256 = 32
)

// GenerateAESKey 从任意长度字符串生成 AES 密钥（使用 SHA256 派生）。
func GenerateAESKey(input string, keySize int) []byte {
	if keySize != AESKeySize128 && keySize != AESKeySize192 && keySize != AESKeySize256 {
		keySize = AESKeySize256
	}
	hash := sha256.Sum256([]byte(input))
	return hash[:keySize]
}
