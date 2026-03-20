// reply.go 提供 Redis/redigo 风格 reply 转换为基础类型的工具函数。
package utils

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrNil 表示 Redis 返回 nil
var ErrNil = errors.New("redigo: nil returned")

// ToInt64 将 reply 转换为 int64
func ToInt64(reply any) (int64, error) {
	switch v := reply.(type) {
	case int64:
		return v, nil
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 64)
		return n, err
	case nil:
		return 0, ErrNil
	default:
		return 0, fmt.Errorf("cannot convert %T to int64", reply)
	}
}

// ToInt32 将 reply 转换为 int32
func ToInt32(reply any) (int32, error) {
	switch v := reply.(type) {
	case int64:
		x := int32(v)
		if int64(x) != v {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(n), nil
	case nil:
		return 0, ErrNil
	case error:
		return 0, v
	default:
		return 0, fmt.Errorf("cannot convert %T to int32", reply)
	}
}

// ToInt 将 reply 转换为 int
func ToInt(reply any, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch v := reply.(type) {
	case int64:
		x := int(v)
		if int64(x) != v {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(v), 10, 0)
		if err != nil {
			return 0, err
		}
		return int(n), nil
	case nil:
		return 0, ErrNil
	case error:
		return 0, v
	default:
		return 0, fmt.Errorf("cannot convert %T to int", reply)
	}
}

// ToString 将 reply 转换为 string
func ToString(reply any, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch v := reply.(type) {
	case []byte:
		return string(v), nil
	case string:
		return v, nil
	case nil:
		return "", ErrNil
	case error:
		return "", v
	default:
		return fmt.Sprint(v), nil
	}
}

// ToBool 将 reply 转换为 bool
func ToBool(reply any, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch v := reply.(type) {
	case bool:
		return v, nil
	case int64:
		return v != 0, nil
	case []byte:
		return len(v) > 0, nil
	case nil:
		return false, ErrNil
	case error:
		return false, v
	default:
		return false, fmt.Errorf("cannot convert %T to bool", reply)
	}
}

// GetMapValue 从 map 中获取指定类型的值
func GetMapValue[T any](m map[int]any, id int) T {
	var zero T
	v, ok := m[id]
	if !ok {
		return zero
	}
	if result, ok := v.(T); ok {
		return result
	}
	return zero
}

// BoolToInt bool 转 int32（true=1, false=0）
func BoolToInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// IntToBool int 转 bool（非0=true, 0=false）
func IntToBool(v int64) bool {
	return v != 0
}
