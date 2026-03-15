package utils

import "time"

// CheckIsSameDayBySec 兼容旧调用：按指定小时划分判断是否同一天。
func CheckIsSameDayBySec(sec1, sec2 int64, hour int) bool {
	return IsSameDayBySecWithHour(sec1, sec2, hour)
}

// GetTodayEndUnix 兼容旧调用：获取当天最后一秒。
func GetTodayEndUnix() int64 {
	return TodayEndUnix()
}

// GetTodayEndMinUnix 兼容旧调用：获取当天最后一分钟。
func GetTodayEndMinUnix() int64 {
	return TodayEndMinuteUnix()
}

// TimestampToday 兼容旧调用：当天零点时间戳（秒）。
func TimestampToday() int64 {
	return TodayStartUnix()
}

// TimestampTodayMillisecond 兼容旧调用：当天结束时间戳（秒，用于 EXPIREAT）。
func TimestampTodayMillisecond() int64 {
	return TodayEndUnix()
}

// GetTargetDayStartUnix 兼容旧调用：目标日期零点时间戳。
func GetTargetDayStartUnix(t time.Time) int64 {
	return DayStartUnix(t)
}

// GetTodayEndUnixInHour 兼容旧调用：获取按 hour 划分的"当日结束"时间点。
func GetTodayEndUnixInHour(t *time.Time, hour int) (int64, error) {
	if t == nil {
		n := Now()
		t = &n
	}
	return dayEndAtHour(*t, hour)
}

// RemoveFirstInt32 兼容旧调用：删除切片中第一个指定值。
func RemoveFirstInt32(arr []int32, value int32) []int32 {
	return RemoveFirst(arr, value)
}

// MicsSlice 兼容旧调用：随机取 n 个不重复元素。
func MicsSlice[T any](arr []T, n int) []T {
	if n <= 0 || len(arr) == 0 {
		return nil
	}
	if n >= len(arr) {
		out := make([]T, len(arr))
		copy(out, arr)
		return out
	}
	shuffled := Shuffle(arr)
	return shuffled[:n]
}

// WeightedRandom 兼容旧调用：按权重随机选取 num 个元素（不重复）。
func WeightedRandom[W ~int | ~int32 | ~int64, T any](weights []W, values []T, num int) []T {
	return WeightedSample(weights, values, num)
}

// GetTodayUnixInHour 兼容旧调用：获取当天指定小时的时间戳。
func GetTodayUnixInHour(t *time.Time, hour int) (int64, error) {
	if t == nil {
		n := Now()
		t = &n
	}
	if hour < 0 || hour >= 24 {
		return 0, ErrInvalidHour
	}
	base := *t
	return time.Date(base.Year(), base.Month(), base.Day(), hour, 0, 0, 0, base.Location()).Unix(), nil
}

// DaysBetweenTwoTimeUnix 兼容旧调用：两个时间戳之间的自然天数差。
func DaysBetweenTwoTimeUnix(sec1, sec2 int64) int32 {
	return DaysBetween(sec1, sec2)
}

// ContainsString 兼容旧调用：判断字符串切片包含关系。
func ContainsString(arr []string, value string) bool {
	return Contains(arr, value)
}

// HasDuplicateInt32 兼容旧调用：是否有重复值，并返回一个重复值。
func HasDuplicateInt32(arr []int32) (bool, int32) {
	seen := make(map[int32]struct{}, len(arr))
	for _, v := range arr {
		if _, ok := seen[v]; ok {
			return true, v
		}
		seen[v] = struct{}{}
	}
	return false, 0
}

// Int64 兼容旧调用：将 reply 转为 int64。
func Int64(reply any) (int64, error) {
	return ToInt64(reply)
}

// GetTimeNowFormat 兼容旧调用：格式化当前时间。
func GetTimeNowFormat() string {
	return FormatNow()
}

// GetTimeOffset 兼容旧调用：获取当前游戏时间偏移。
func GetTimeOffset() time.Duration {
	return GetOffset()
}

// SetTimeOffset 兼容旧调用：设置游戏时间偏移。
func SetTimeOffset(offset time.Duration) {
	SetOffset(offset)
}

// TimeOffsetEnabled 兼容旧调用：是否允许时间偏移。
func TimeOffsetEnabled() bool {
	return IsOffsetEnabled()
}

// SetTimeOffsetEnabled 兼容旧调用：启用/禁用偏移功能。
func SetTimeOffsetEnabled(enabled bool) {
	InitClock(ClockConfig{AllowOffset: enabled})
}

// LoadTimeOffsetFromRedis 兼容旧调用：从存储层重新加载偏移。
func LoadTimeOffsetFromRedis() {
	ReloadOffset()
}
