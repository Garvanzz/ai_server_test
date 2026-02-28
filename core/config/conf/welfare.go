package conf

type DaySign struct {
	Id        int64   `json:"Id"`
	Day       int32   `json:"Day"`
	Reward    []ItemE `json:"Reward"`
	AccSign   bool    `json:"AccSign"`
	CumReward []ItemE `json:"CumReward"`
}

type FunctionOpen struct {
	Id        int64    `json:"Id"`
	Type      string   `json:"Type"`
	Reward    []ItemE  `json:"Reward"`
	Condition []int32  `json:"Condition"`
	Param     []string `json:"Param"`
}
