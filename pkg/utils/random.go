// Package utils 提供与业务无关的通用工具，供全项目使用。
// random.go 提供随机数生成、随机选择和随机字符串生成功能。
package utils

import (
	"math/rand"
	"time"
)

var rnd *rand.Rand

func init() {
	UpdateRand(time.Now().UnixNano())
}

// UpdateRand 更新随机数种子
func UpdateRand(t int64) {
	rnd = rand.New(rand.NewSource(t))
}

// Random 生成指定长度的纯数字字符串（首位不为0）
func Random(length int) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		if i == 0 {
			result[i] = byte('1' + rnd.Intn(9)) // 首位 1-9
		} else {
			result[i] = byte('0' + rnd.Intn(10)) // 其余位 0-9
		}
	}
	return string(result)
}

// RandInt 生成 [min, max] 区间的随机整数
func RandInt[T ~int | ~int32 | ~int64](min, max T) T {
	if min >= max {
		return max
	}

	// 转换为 int 计算，避免溢出
	minInt := int64(min)
	maxInt := int64(max)
	range_ := maxInt - minInt + 1

	return T(minInt + int64(rnd.Int63n(range_)))
}

// Shuffle 随机打乱切片顺序
func Shuffle[T any](arr []T) []T {
	tmp := make([]T, len(arr))
	copy(tmp, arr)
	rnd.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})
	return tmp
}

// Sample 从切片中随机选取指定数量的元素（不重复）
func Sample[T comparable](arr []T, count int) []T {
	if count <= 0 || len(arr) == 0 {
		return nil
	}
	if count >= len(arr) {
		return Shuffle(arr)
	}

	tmp := make([]T, len(arr))
	copy(tmp, arr)

	rnd.Shuffle(len(tmp), func(i, j int) {
		tmp[i], tmp[j] = tmp[j], tmp[i]
	})

	return tmp[:count]
}

// WeightedSample 按权重随机选取指定数量的元素（不重复）
func WeightedSample[T ~int | ~int32 | ~int64, E any](weights []T, values []E, num int) []E {
	if num <= 0 || len(weights) != len(values) || len(weights) == 0 {
		return nil
	}
	if num > len(weights) {
		num = len(weights)
	}

	cv := make([]E, len(values))
	copy(cv, values)
	wh := make([]T, len(weights))
	copy(wh, weights)

	result := make([]E, 0, num)
	for i := 0; i < num; i++ {
		idx := weightedIndex(wh)
		if idx < 0 {
			break
		}
		result = append(result, cv[idx])
		// 移除已选取的元素
		wh = append(wh[:idx], wh[idx+1:]...)
		cv = append(cv[:idx], cv[idx+1:]...)
	}
	return result
}

// WeightedChoice 按权重随机选择一个索引
func WeightedChoice[T ~int | ~int32 | ~int64](weights []T) int {
	return weightedIndex(weights)
}

// WeightIndex 兼容旧调用：按权重随机索引
func WeightIndex(weights []int32) int {
	return weightedIndex(weights)
}

// weightedIndex 内部实现：按权重选择索引
func weightedIndex[T ~int | ~int32 | ~int64](weights []T) int {
	if len(weights) == 0 {
		return -1
	}

	sum := T(0)
	for _, w := range weights {
		if w > 0 {
			sum += w
		}
	}
	if sum <= 0 {
		return -1
	}

	target := T(rnd.Float64() * float64(sum))
	acc := T(0)
	for i, w := range weights {
		acc += w
		if acc > target {
			return i
		}
	}
	return len(weights) - 1
}

// HitRate 按概率判断是否命中（upNum/downNum 表示几分之几的概率）
// 例如 HitRate(1, 100) 表示 1% 的概率返回 true
func HitRate(upNum, downNum int32) bool {
	if downNum <= 0 {
		return false
	}
	if upNum >= downNum {
		return true
	}
	if upNum <= 0 {
		return false
	}
	return rnd.Int31n(downNum) < upNum
}

// HitPercent 按百分比判断是否命中（0-100）
func HitPercent(percent int32) bool {
	return HitRate(percent, 100)
}

// RandStringKind 定义随机字符串类型

type RandStringKind int

const (
	RandNumeric      RandStringKind = iota // 纯数字
	RandLowerCase                          // 小写字母
	RandUpperCase                          // 大写字母
	RandAlphabetic                         // 字母（大小写）
	RandAlphanumeric                       // 字母数字混合
)

// RandomString 生成指定长度的随机字符串
func RandomString(length int, kind RandStringKind) string {
	if length <= 0 {
		return ""
	}

	const (
		numeric      = "0123456789"
		lowerCase    = "abcdefghijklmnopqrstuvwxyz"
		upperCase    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		alphabetic   = lowerCase + upperCase
		alphanumeric = numeric + alphabetic
	)

	var charset string
	switch kind {
	case RandNumeric:
		charset = numeric
	case RandLowerCase:
		charset = lowerCase
	case RandUpperCase:
		charset = upperCase
	case RandAlphabetic:
		charset = alphabetic
	case RandAlphanumeric:
		charset = alphanumeric
	default:
		charset = alphanumeric
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = charset[rnd.Intn(len(charset))]
	}
	return string(result)
}

// RandomNumeric 生成指定长度的纯数字字符串（可包含前导0）
func RandomNumeric(length int) string {
	return RandomString(length, RandNumeric)
}

// RandomAlphabetic 生成指定长度的纯字母字符串
func RandomAlphabetic(length int) string {
	return RandomString(length, RandAlphabetic)
}

// RandomAlphanumeric 生成指定长度的字母数字混合字符串
func RandomAlphanumeric(length int) string {
	return RandomString(length, RandAlphanumeric)
}

// RandomHex 生成指定长度的十六进制字符串
func RandomHex(length int) string {
	if length <= 0 {
		return ""
	}
	const hex = "0123456789abcdef"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = hex[rnd.Intn(len(hex))]
	}
	return string(result)
}

// CoinFlip 模拟抛硬币，返回 true 或 false，概率各 50%
func CoinFlip() bool {
	return rnd.Intn(2) == 0
}

// Roll 模拟掷骰子，返回 1-n 的随机整数
func Roll(n int) int {
	if n <= 1 {
		return 1
	}
	return rnd.Intn(n) + 1
}
