package utils

// MinInt 返回两整数中较小者
func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxInt 返回两整数中较大者
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ClampInt 将 v 限制在 [lo, hi] 内
func ClampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// MinInt32 返回两数中较小者
func MinInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

// MaxInt32 返回两数中较大者
func MaxInt32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}

// ClampInt32 将 v 限制在 [lo, hi] 内
func ClampInt32(v, lo, hi int32) int32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// MinInt64 返回两数中较小者
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 返回两数中较大者
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ClampInt64 将 v 限制在 [lo, hi] 内
func ClampInt64(v, lo, hi int64) int64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
