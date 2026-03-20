// time.go 提供时间、日期计算工具函数。
// 本包使用 clock.Now() 作为时间源，支持游戏逻辑时间偏移。
package utils

import (
	"errors"
	"time"
)

var (
	ErrInvalidHour = errors.New("hour must be between 0 and 23")
)

// TodayEndUnix 获取当天最后一秒的时间戳（23:59:59）
func TodayEndUnix() int64 {
	now := Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, time.Local).Unix()
}

// TodayEndMinuteUnix 获取当天最后一分钟的时间戳（23:59:00）
func TodayEndMinuteUnix() int64 {
	now := Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 23, 59, 0, 0, time.Local).Unix()
}

// TodayStartUnix 获取当天零点时间戳（00:00:00）
func TodayStartUnix() int64 {
	now := Now()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
}

// TodayStartUnixMilli 获取当天零点时间戳（毫秒）
func TodayStartUnixMilli() int64 {
	return TodayStartUnix() * 1000
}

// DayEndUnix 获取指定日期最后一秒的时间戳
func DayEndUnix(t time.Time) int64 {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 0, t.Location()).Unix()
}

// DayStartUnix 获取指定日期零点时间戳
func DayStartUnix(t time.Time) int64 {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location()).Unix()
}

// DayEndUnixAtHour 获取当日指定小时的结束时间戳（该小时开始时刻）
// 如果当前时间已过该小时，则返回次日的该小时时间戳
func DayEndUnixAtHour(hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return 0, ErrInvalidHour
	}

	now := Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local)

	if now.Hour() >= hour {
		target = target.Add(24 * time.Hour)
	}

	return target.Unix(), nil
}

// DayUnixAtHour 获取当日指定小时的时间戳（该小时开始时刻）
func DayUnixAtHour(hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return 0, ErrInvalidHour
	}

	now := Now()
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local)
	return target.Unix(), nil
}

// IsSameDay 判断两个时间是否为同一天
func IsSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsSameDayBySec 判断两个时间戳（秒级）是否为同一天
func IsSameDayBySec(sec1, sec2 int64) bool {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)
	return IsSameDay(t1, t2)
}

// IsSameDayBySecWithHour 判断两个时间戳在指定小时划分下是否为同一天
func IsSameDayBySecWithHour(sec1, sec2 int64, hour int) bool {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)

	end1, _ := dayEndAtHour(t1, hour)
	end2, _ := dayEndAtHour(t2, hour)
	return end1 == end2
}

// dayEndAtHour 内部函数：计算时间在某小时划分下的结束时间
func dayEndAtHour(t time.Time, hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return 0, ErrInvalidHour
	}

	end := time.Date(t.Year(), t.Month(), t.Day(), hour, 0, 0, 0, t.Location())
	if t.Hour() >= hour {
		end = end.Add(24 * time.Hour)
	}
	return end.Unix(), nil
}

// DaysBetween 计算两个时间戳之间的自然天数差（绝对值）
func DaysBetween(sec1, sec2 int64) int32 {
	if sec1 > sec2 {
		sec1, sec2 = sec2, sec1
	}

	firstDayEnd := DayEndUnix(time.Unix(sec1, 0))
	if sec2 <= firstDayEnd {
		return 0 // 同一天
	}

	// 超过1天
	timeDiff := sec2 - firstDayEnd - 1
	return int32(timeDiff/(24*60*60)) + 1
}

// DaysDiff 计算两个时间戳之间的天数差（考虑正负，以日期为准）
func DaysDiff(sec1, sec2 int64) int32 {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)

	// 将时间部分置零，仅保留日期
	date1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, t1.Location())
	date2 := time.Date(t2.Year(), t2.Month(), t2.Day(), 0, 0, 0, 0, t2.Location())

	return int32(date2.Sub(date1).Hours() / 24)
}

// NextWeekday 获取下周指定星期的日期
func NextWeekday(weekday time.Weekday) time.Time {
	today := Now()
	if weekday == 0 {
		weekday = 7 // 将 Sunday 从 0 转为 7
	}

	// 计算距离下周目标日期的天数
	currentWeekday := int(today.Weekday())
	if currentWeekday == 0 {
		currentWeekday = 7
	}
	daysUntil := (7 - currentWeekday) + int(weekday)

	return today.AddDate(0, 0, daysUntil)
}

// GetWeekday 获取本周中特定星期的日期
func GetWeekday(weekday time.Weekday) time.Time {
	t := Now()
	if weekday == 0 {
		weekday = 7
	}
	currentWeekday := t.Weekday()
	if currentWeekday == 0 {
		currentWeekday = 7
	}
	offset := int(weekday - currentWeekday)
	return t.AddDate(0, 0, offset)
}

// IsSameWeek 判断两个时间是否为同一周（按 ISO 周标准）
func IsSameWeek(t1, t2 time.Time) bool {
	y1, w1 := t1.ISOWeek()
	y2, w2 := t2.ISOWeek()
	return y1 == y2 && w1 == w2
}

// IsSameWeekBySec 判断两个时间戳是否为同一周
func IsSameWeekBySec(sec1, sec2 int64) bool {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)
	return IsSameWeek(t1, t2)
}

// IsSameMonth 判断两个时间是否为同一月
func IsSameMonth(t1, t2 time.Time) bool {
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()
	return y1 == y2 && m1 == m2
}

// IsSameMonthBySec 判断两个时间戳是否为同一月
func IsSameMonthBySec(sec1, sec2 int64) bool {
	t1 := time.Unix(sec1, 0)
	t2 := time.Unix(sec2, 0)
	return IsSameMonth(t1, t2)
}

// CurrentYear 获取当前年份
func CurrentYear() int {
	return Now().Year()
}

// CurrentMonth 获取当前月份
func CurrentMonth() time.Month {
	return Now().Month()
}

// CurrentDay 获取当前日期（几号）
func CurrentDay() int {
	return Now().Day()
}

// CurrentHour 获取当前小时
func CurrentHour() int {
	return Now().Hour()
}

// FormatNow 返回当前时间的格式化字符串 "2006-01-02 15:04:05"
func FormatNow() string {
	return Now().Format("2006-01-02 15:04:05")
}

// FormatTime 格式化指定时间为 "2006-01-02 15:04:05"
func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化指定日期为 "2006-01-02"
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseTime 解析时间字符串 "2006-01-02 15:04:05"
func ParseTime(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.Local)
}

// ParseDate 解析日期字符串 "2006-01-02"
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, time.Local)
}

// AddDays 添加指定天数到时间戳
func AddDays(sec int64, days int) int64 {
	return sec + int64(days*24*60*60)
}

// AddHours 添加指定小时数到时间戳
func AddHours(sec int64, hours int) int64 {
	return sec + int64(hours*60*60)
}

// AddMinutes 添加指定分钟数到时间戳
func AddMinutes(sec int64, minutes int) int64 {
	return sec + int64(minutes*60)
}

// SecondsUntilTomorrow 计算距离明天零点还有多少秒
func SecondsUntilTomorrow() int64 {
	return TodayEndUnix() - Now().Unix() + 1
}

// SecondsSinceTodayStart 计算从今天零点开始经过了多少秒
func SecondsSinceTodayStart() int64 {
	return Now().Unix() - TodayStartUnix()
}
