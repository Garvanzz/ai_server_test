package id

import (
	"math"
	"strings"
)

const (
	chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	hex   = 62
)

// Id 加密
func Itoa(id int64) string {
	bytes := []byte{}
	for id > 0 {
		bytes = append(bytes, chars[id%hex])
		id = id / hex
	}
	reverse(bytes)
	return string(bytes)
}

// ID 解密
func Atoi(strid string) int64 {
	var id int64
	n := len(strid)
	for i := 0; i < n; i++ {
		pos := strings.IndexByte(chars, strid[i])
		id += int64(math.Pow(hex, float64(n-i-1)) * float64(pos))
	}
	return id
}

func reverse(a []byte) {
	for left, right := 0, len(a)-1; left < right; left, right = left+1, right-1 {
		a[left], a[right] = a[right], a[left]
	}
}
