package conf

type MonthCard struct {
	Id                 int32   `json:"Id"`
	Type               int32   `json:"Type"`
	Day                int32   `json:"Day"`
	Reward             []ItemE `json:"Reward"`
	ExchangeCount      int32   `json:"ExchangeCount"`
	BoxMissionCount    int32   `json:"BoxMissionCount"`
	LingyuMissionCount int32   `json:"LingyuMissionCount"`
	ClimbeTowerCount   int32   `json:"ClimbeTowerCount"`
	DanaotiangongCount int32   `json:"DanaotiangongCount"`
}
