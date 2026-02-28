package conf

type Buff struct {
	Id          int32   `json:"Id"`
	BuffType    int32   `json:"BuffType"`
	BuffTarget  int32   `json:"BuffTarget"`
	TargetLimit int32   `json:"TargetLimit"`
	TargetNum   int32   `json:"TargetNum"`
	BuffValue   []int32 `json:"BuffValue"`
	Trigger     int32   `json:"Trigger"`
	BuffRound   int32   `json:"BuffRound"`
	RemoveType  int32   `json:"RemoveType"`
	OverLayer   int32   `json:"OverLayer"`
	IsDebuff    bool    `json:"IsDebuff"`
	IsDispel    bool    `json:"IsDispel"`
}
