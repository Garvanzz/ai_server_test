package conf

type DrawPool struct {
	Id           int32   `json:"Id"`
	Type         int32   `json:"Type"`
	Weight       []int32 `json:"Weight"`
	MiniNum      int32   `json:"MiniNum"`
	MiniValue    int32   `json:"MiniValue"`
	StageId      int32   `json:"StageId"`
	ActivityType int32   `json:"ActivityType"`
	StartTime    string  `json:"StartTime"`
	EndTime      string  `json:"EndTime"`
	HeroId       int32   `json:"HeroId"`
	Param        int32   `json:"Param"`
}

type DrawStageAward struct {
	Id       int32    `json:"Id"`
	PoolType int32    `json:"PoolType"`
	Progress []int32  `json:"Progress"`
	Award    []string `json:"Award"`
}
