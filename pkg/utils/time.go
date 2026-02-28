package utils

import (
	"errors"
	"time"
)

func GetTodayEndUnix() int64 {
	now := time.Now()
	year, month, day := now.Date()
	today_end := time.Date(year, month, day, 23, 59, 59, 0, time.Local).Unix()
	return today_end
}

// 获取当天最后一分钟
func GetTodayEndMinUnix() int64 {
	now := time.Now()
	year, month, day := now.Date()
	today_end := time.Date(year, month, day, 23, 59, 0, 0, time.Local).Unix()
	return today_end
}

func GetTargetDayEndUnix(t time.Time) int64 {
	year, month, day := t.Date()
	thatDayEnd := time.Date(year, month, day, 23, 59, 59, 0, time.Local).Unix()
	return thatDayEnd
}

func GetTargetDayStartUnix(t time.Time) int64 {
	year, month, day := t.Date()
	thatDayStart := time.Date(year, month, day, 0, 0, 0, 0, time.Local).Unix()
	return thatDayStart
}

// 获取当日某个小时的结束时间戳
func GetTodayEndUnixInHour(now *time.Time, hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return int64(0), errors.New("hour error")
	}

	if now == nil {
		*now = time.Now()
	}

	sec := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local).Unix()

	nowHour := now.Hour()
	if nowHour >= hour {
		sec += 24 * 3600
	}

	return sec, nil
}

// 获取当日某个小时的结束时间戳
func GetTodayUnixInHour(now *time.Time, hour int) (int64, error) {
	if hour < 0 || hour >= 24 {
		return int64(0), errors.New("hour error")
	}

	if now == nil {
		*now = time.Now()
	}

	sec := time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, time.Local).Unix()

	return sec, nil
}

func CheckIsSameDay(time1 *time.Time, time2 *time.Time, hour int) bool {
	if time1 == nil || time2 == nil {
		return false
	}

	t1EndSec, _ := GetTodayEndUnixInHour(time1, hour)
	t2EndSec, _ := GetTodayEndUnixInHour(time2, hour)

	return t1EndSec == t2EndSec
}

func CheckIsSameDayBySec(time1 int64, time2 int64, hour int) bool {
	t1 := time.Unix(time1, int64(0))
	t2 := time.Unix(time2, int64(0))
	return CheckIsSameDay(&t1, &t2, hour)
}

// 获取本地当天零点时间截
func TimestampToday() int64 {
	return time.Date(Year(), Month(), Day(), 0, 0, 0, 0, time.Local).Unix()
}

// 获取本地当天零点时间截毫秒
func TimestampTodayMillisecond() int64 {
	return (time.Date(Year(), Month(), Day(), 0, 0, 0, 0, time.Local).Unix()) * 1000
}

// 获取当前年
func Year() int {
	return time.Now().Year()
}

// 获取当前月
func Month() time.Month {
	return time.Now().Month()
}

// 获取当前天
func Day() int {
	return time.Now().Day()
}

// 计算2个时间之间的自然天数之差
func DaysBetweenTwoTimeUnix(time1, time2 int64) int32 {
	//如果时间1比时间2还要远 调换一下位置
	if time1 > time2 {
		temp := time1
		time1 = time2
		time2 = temp
	}

	//默认time1比time2时间小了
	//获取time1当前结束时间
	firstDayEndUnix := GetTargetDayEndUnix(time.Unix(time1, 0))
	if time2 <= firstDayEndUnix {
		return 0 //同一天
	} else {
		//肯定超过1天了
		timeDiff := time2 - firstDayEndUnix - 1
		return int32(timeDiff/(24*60*60)) + 1

	}

}

// DaysDiff 计算天数差（不同日期差一秒也算一天，要求是自然天数差）
func DaysDiff(t1, t2 int64) int32 {
	date1 := time.Unix(t1, 0)
	date2 := time.Unix(t2, 0)

	// 将日期的时间部分置为零
	date1 = time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	date2 = time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, date2.Location())

	// 计算日期差异
	daysDiff := int32(date2.Sub(date1).Hours() / 24)

	return daysDiff
}

//// 获取下周几的一个日期
//func GetNextWeekday(currentTime time.Time, weekday time.Weekday) time.Time {
//	// 计算与目标星期几的时间差
//	daysUntilTargetDay := int((weekday - currentTime.Weekday() + 7) % 7)
//
//	// 加上时间差得到下一个目标星期几的时间
//	return currentTime.Add(time.Duration(daysUntilTargetDay) * 24 * time.Hour)
//}

// GetNextWeekday 获取下周几的一个日期
func GetNextWeekday(today time.Time, weekday time.Weekday) time.Time {
	if weekday == 0 {
		weekday = 7
	}

	dayNum := (int(time.Sunday-today.Weekday())+7)%7 + int(weekday)

	return today.AddDate(0, 0, dayNum)
}

// GetWeekday 获取本周中特定星期的日期
func GetWeekday(t time.Time, weekday time.Weekday) time.Time {
	if weekday == 0 {
		weekday = 7
	}

	today := t.Weekday()
	offset := int(weekday - today)

	return t.AddDate(0, 0, offset)
}

// IsSameWeek 是否是同一周
func IsSameWeek(t1, t2 time.Time) bool {
	y1, w1 := t1.ISOWeek()
	y2, w2 := t2.ISOWeek()
	return y1 == y2 && w1 == w2
}

func IsSameWeekBySec(time1 int64, time2 int64) bool {
	t1 := time.Unix(time1, 0)
	t2 := time.Unix(time2, 0)
	return IsSameWeek(t1, t2)
}

// IsSameMonth 是否是同一月
func IsSameMonth(t1, t2 time.Time) bool {
	y1, m1, _ := t1.Date()
	y2, m2, _ := t2.Date()
	return y1 == y2 && m1 == m2
}

func IsSameMonthBySec(time1 int64, time2 int64) bool {
	t1 := time.Unix(time1, 0)
	t2 := time.Unix(time2, 0)
	return IsSameMonth(t1, t2)
}

func GetTimeNowFormat() string {
	// 获取当前时间
	now := time.Now()

	// 构造当天的 00:00:00
	midnight := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(), 0, now.Location())

	// 格式化为 "2006-01-02 15:04:05"
	formatted := midnight.Format("2006-01-02 15:04:05")
	return formatted
}
