package model

import (
	"xfx/proto/proto_welfare"
)

// 福利
type Welfare struct {
	DaySign      *DaySign
	MonthCard    map[int32]*MonthCard
	FunctionOpen []string
}

// 签到
type DaySign struct {
	Day          []int32
	IsDaySign    bool
	FirstDayTime int64
	SignTime     int64
	AccDay       []int32
}

// 月卡
type MonthCard struct {
	IsGet   bool
	GetTime int64
}

func ToWelfareMonthCardProto(maps map[int32]*MonthCard) map[int32]*proto_welfare.MonthCardOption {
	m := make(map[int32]*proto_welfare.MonthCardOption, 0)
	for k, v := range maps {
		m[k] = &proto_welfare.MonthCardOption{
			IsGet: v.IsGet,
		}
	}

	return m
}
