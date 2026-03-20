// Package id 提供分布式唯一 ID 生成器（Twitter Snowflake 算法变体）。
package id

import (
	"strings"
)

const (
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base62      = 62
)

// Itoa 将 int64 ID 转换为 62 进制短字符串（缩短长度）。
// 使用字符集：0-9a-zA-Z
func Itoa(id int64) string {
	if id == 0 {
		return "0"
	}

	// 处理负数
	negative := id < 0
	if negative {
		id = -id
	}

	var buf [32]byte // 足够容纳任何 int64
	i := len(buf)

	for id > 0 {
		i--
		buf[i] = base62Chars[id%base62]
		id /= base62
	}

	if negative {
		i--
		buf[i] = '-'
	}

	return string(buf[i:])
}

// Atoi 将 62 进制字符串还原为 int64 ID。
func Atoi(s string) int64 {
	if s == "" {
		return 0
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	var result int64
	for i := 0; i < len(s); i++ {
		c := s[i]
		idx := strings.IndexByte(base62Chars, c)
		if idx < 0 {
			// 非法字符，返回已解析的部分
			break
		}
		result = result*base62 + int64(idx)
	}

	if negative {
		result = -result
	}

	return result
}

// MustAtoi 将 62 进制字符串还原为 int64 ID，出错时返回 0。
func MustAtoi(s string) int64 {
	return Atoi(s)
}

// IsValidIDString 检查字符串是否为有效的 62 进制 ID。
func IsValidIDString(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(base62Chars, s[i]) < 0 {
			return false
		}
	}
	return true
}
