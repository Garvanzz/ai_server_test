package utils

import "strconv"

// ParseInt64 将字符串解析为 int64，失败返回 error（不写日志，由调用方决定）
func ParseInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

// MustParseInt64 将字符串解析为 int64，失败返回 0（兼容原 core/common.StringToInt64 语义）
func MustParseInt64(str string) int64 {
	n, _ := ParseInt64(str)
	return n
}
