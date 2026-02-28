package conf

type Guild struct {
	Id         int32 `json:"id"`
	Num        int32 `json:"num"`
	Elder      int32 `json:"elder"`       // 长老数量
	ViceMaster int32 `json:"vice_master"` // 副会长数量
}

type GuildMaterial struct {
	Id   int32 `json:"Id"`
	Rare int32 `json:"rare"`
}

type GuildElement struct {
	Id               int32           `json:"Id"`
	Rare             int32           `json:"rare"`
	IsElement        bool            `json:"IsElement"`
	BasicTime        int32           `json:"BasicTime"`
	Title            int32           `json:"title"`
	BasicSuccessRare int32           `json:"BasicSuccessRare"`
	Stage            int32           `json:"stage"`
	Material         map[int32]int32 `json:"material"`
	MergeElement     []int32         `json:"mergeElement"`
}

type GuildTitle struct {
	Id             int32   `json:"Id"`
	DailyGive      []ItemE `json:"DailyGive"`
	PraySucRare    int32   `json:"PraySucRare"`
	PrayTime       int32   `json:"PrayTime"`
	PrayRangeType  []int32 `json:"PrayRangeType"`
	PrayRangeValue []int32 `json:"PrayRangeValue"`
}
