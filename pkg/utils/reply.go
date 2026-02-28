package utils

import (
	"errors"
	"fmt"
	"strconv"
)

var ErrNil = errors.New("redigo: nil returned")

func Int64(reply any) (int64, error) {
	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 64)
		return n, err
	case nil:
		return 0, ErrNil
	default:
		return 0, fmt.Errorf("reply to int64 type got type %T", reply)
	}
}

func Int32(reply interface{}) (int32, error) {
	switch reply := reply.(type) {
	case int64:
		x := int32(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 32)
		return int32(n), err
	case nil:
		return 0, ErrNil
	case error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int, got type %T", reply)
}

func Int(reply interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 0)
		return int(n), err
	case nil:
		return 0, ErrNil
	case error:
		return 0, reply
	}
	return 0, fmt.Errorf("redigo: unexpected type for Int, got type %T", reply)
}

func String(reply interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}
	switch reply := reply.(type) {
	case []byte:
		return string(reply), nil
	case string:
		return reply, nil
	case nil:
		return "", ErrNil
	case error:
		return "", reply
	}
	return "", fmt.Errorf("redigo: unexpected type for String, got type %T", reply)
}

func Bool(reply interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch reply := reply.(type) {
	case bool:
		return reply, nil
	case nil:
		return false, ErrNil
	case error:
		return false, reply
	}
	return false, fmt.Errorf("redigo: unexpected type for String, got type %T", reply)
}

func Convert[T any](m map[int]any, id int) T {
	var t T
	if v, ok := m[id]; !ok {
		return t
	} else {
		return v.(T)
	}
}

func BoolToInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
