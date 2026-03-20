// math.go 提供数学计算工具函数。
package utils

import (
	"cmp"
	"math"
)

// Min 返回两数中较小者
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 返回两数中较大者
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Clamp 将 v 限制在 [lo, hi] 范围内
// 如果 lo > hi，会交换边界
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if lo > hi {
		lo, hi = hi, lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Abs 返回整数的绝对值
func Abs[T ~int | ~int32 | ~int64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// AbsFloat 返回浮点数的绝对值
func AbsFloat[T ~float32 | ~float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Sign 返回数值的符号：负数返回 -1，正数返回 1，零返回 0
func Sign[T cmp.Ordered](x T) int {
	var zero T
	if x < zero {
		return -1
	}
	if x > zero {
		return 1
	}
	return 0
}

// Between 判断 v 是否在 [min, max] 区间内（包含边界）
func Between[T cmp.Ordered](v, min, max T) bool {
	if min > max {
		min, max = max, min
	}
	return v >= min && v <= max
}

// Lerp 线性插值：返回 a 和 b 之间 t 比例处的值，t 应该在 [0, 1] 之间
func Lerp[T ~float32 | ~float64](a, b, t T) T {
	return a + (b-a)*t
}

// LerpInt 整数线性插值
func LerpInt(a, b int, t float64) int {
	return a + int(float64(b-a)*t)
}

// Round 四舍五入到最接近的整数
func Round[T ~float32 | ~float64](x T) int {
	return int(math.Floor(float64(x) + 0.5))
}

// Floor 向下取整
func Floor[T ~float32 | ~float64](x T) int {
	return int(math.Floor(float64(x)))
}

// Ceil 向上取整
func Ceil[T ~float32 | ~float64](x T) int {
	return int(math.Ceil(float64(x)))
}

// Pow 计算 x 的 y 次方
func Pow[T ~float32 | ~float64](x, y T) T {
	return T(math.Pow(float64(x), float64(y)))
}

// Sqrt 计算平方根
func Sqrt[T ~float32 | ~float64](x T) T {
	return T(math.Sqrt(float64(x)))
}

// IsPowerOfTwo 判断是否为 2 的幂
func IsPowerOfTwo[T ~int | ~int32 | ~int64](x T) bool {
	return x > 0 && (x&(x-1)) == 0
}

// NextPowerOfTwo 返回大于等于 x 的最小 2 的幂
func NextPowerOfTwo[T ~int | ~int32 | ~int64](x T) T {
	if x <= 1 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	if any(x).(int64) > 1<<32 {
		x |= x >> 32
	}
	return x + 1
}

// Percent 计算百分比（value/total * 100），避免除以零
func Percent[T ~int | ~int32 | ~int64 | ~float32 | ~float64](value, total T) float64 {
	if total == 0 {
		return 0
	}
	return float64(value) / float64(total) * 100
}

// PercentOf 计算 value 是 total 的百分之多少（0-100）
func PercentOf[T ~int | ~int32 | ~int64 | ~float32 | ~float64](value, total T) float64 {
	return Percent(value, total)
}

// SafeDiv 安全除法，避免除以零
func SafeDiv[T ~int | ~int32 | ~int64 | ~float32 | ~float64](a, b, defaultValue T) T {
	if b == 0 {
		return defaultValue
	}
	return a / b
}

// InRange 判断值是否在指定范围内（开区间）
func InRange[T cmp.Ordered](v, min, max T) bool {
	return v > min && v < max
}
