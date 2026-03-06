package utils

import "strconv"

// ParseInt64 将字符串解析为 int64，失败返回 error（不写日志，由调用方决定）
func ParseInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}
